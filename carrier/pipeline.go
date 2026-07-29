package carrier

import (
	"context"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const (
	PipelineType     event.Type = TypePrefix + ".pipeline"
	PipelineStopType event.Type = TypePrefix + ".pipeline.stop"
)

type PipelineStage func(prev event.Event) (next event.Event)

// pipelineStop is a special payload that can be used to interrupt a pipeline execution.
// It wrapp the actual event user-defined that will be triggered.
// pipelineStop.Event can be left nil.
// Note that the pipelineStop payload itself will never be published.
type pipelineStop struct {
	// Wrapped event that will be published.
	Event event.Event
}

func (p pipelineStop) EventType() event.Type {
	return PipelineStopType
}

// StopPipeline interrupts the pipeline execution at the stage it returned from.
// The pipeline will no longer emit events.
func StopPipeline() event.Event {
	return event.New(pipelineStop{})
}

// StopPipelineWithEvent interrupts the pipeline execution at the stage it returned from.
// The pipeline will be stopped and the specified event will be published.
func StopPipelineWithEvent(evt event.Event) event.Event {
	return event.New(pipelineStop{Event: evt})
}

// Pipeline is a carrier that emits events sequentially thanks to a list of functions.
// Each function of the pipeline takes the previously received completion event as input and returns the next event to be processed.
type Pipeline struct {
	InitEvent event.Event     `json:"initEvent"`
	Stages    []PipelineStage `json:"-"`
	OnTimeout event.Event     `json:"onTimeout,omitempty"`

	completionCondition CompletionCondition
	timeout             time.Duration
	evtCarrier          event.Event
}

var (
	_ Carrier = (*Pipeline)(nil)
)

// NewPipeline creates a new Pipeline carrier.
//
// The initEvent is dispatched first. Once its completion event is received
// (by default, its direct followup, but you can change this behavior with an option),
// is forwarded as an argument for the first function (stage) of the pipeline. The function is supposed
// to return the next event to be processed, and so on.
//
// Parameters:
//
//	initEvent - The first event to be published
//	stages - The sequence of functions to be executed in the pipeline
//	onTimeout - The event to be published when the pipeline times out
//	opts - Optional configuration options
func NewPipeline(
	initEvent event.Event,
	stages []PipelineStage,
	onTimeout event.Event,
	opts ...Option,
) event.Event {
	c := &Pipeline{
		InitEvent: initEvent,
		OnTimeout: onTimeout,
		Stages:    stages,
	}

	cfg := &carrierConfig{
		maxConcurrency:      defaultCarrierConcurrency,
		timeout:             defaultCarrierTimeout,
		completionCondition: CompletedOnFollowupReceived,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	c.completionCondition = cfg.completionCondition
	c.timeout = cfg.timeout

	c.evtCarrier = makeEvent(c, cfg)

	return c.evtCarrier
}

func (c *Pipeline) EventType() event.Type {
	return PipelineType
}

type pipelineItem struct {
	currentStage PipelineStage
	next         event.Event
}

// Dispatch is used by the event bus to dispatch the carried event.
// You are not supposed to call this method directly.
func (c *Pipeline) Dispatch(bus event.Bus) {
	if len(c.Stages) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

	defer cancel()

	workload := make(chan pipelineItem, 1)
	defer close(workload)

	var currIdx int
	workload <- pipelineItem{
		currentStage: c.Stages[currIdx],
		next:         c.InitEvent,
	}

	for {
		select {
		case wl := <-workload:
			var once sync.Once
			finished := make(chan event.Event)

			sub := bus.Subscribe().
				On(event.IsFollowupOf(wl.next), func(received event.Event) {
					if c.completionCondition(wl.next, received) {
						once.Do(func() {
							select {
							case finished <- received:
							default:
							}
						})
					}
				})
			sub.ListenWithWorkers(1)
			bus.Publish(wl.next)

			select {
			case prev := <-finished:
				newNext := wl.currentStage(prev)

				if stop, ok := newNext.Payload().(pipelineStop); ok {
					if lastEvt := stop.Event; lastEvt != nil {
						bus.Publish(lastEvt)
					}
					bus.Unsubscribe(sub)
					close(finished)
					return
				}

				currIdx++
				if currIdx == len(c.Stages) {
					bus.Publish(newNext)
					bus.Unsubscribe(sub)
					close(finished)
					return
				}

				workload <- pipelineItem{
					currentStage: c.Stages[currIdx],
					next:         newNext,
				}
			case <-ctx.Done():
				bus.Publish(c.OnTimeout)
			}

			close(finished)
			bus.Unsubscribe(sub)

		case <-ctx.Done():
			return
		}
	}
}
