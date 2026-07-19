package carrier

import (
	"context"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const (
	PipelineType     event.Type = TypePrefix + ".pipeline"
	PipelineStopType event.Type = TypePrefix + ".pipeline.stop"
)

// PipelineStop is a special payload that can be used to interrupt a pipeline execution.
// It wrapp the actual event user-defined that will be triggered.
// PipelineStop.Event can be left nil (not recommended)
type PipelineStop struct {
	Event event.Event
}

func (p PipelineStop) EventType() event.Type {
	return PipelineStopType
}

type Pipeline struct {
	InitEvent event.Event                                 `json:"initEvent"`
	Pipeline  []func(prev event.Event) (next event.Event) `json:"-"`
	OnTimeout event.Event                                 `json:"onTimeout,omitempty"`

	completionCondition CompletionCondition
	timeout             time.Duration
	evtCarrier          event.Event
}

var (
	_ Carrier = (*Pipeline)(nil)
)

func NewPipeline(
	initEvent event.Event,
	pipeline []func(prev event.Event) (next event.Event),
	onTimeout event.Event,
	opts ...Option,
) event.Event {
	c := &Pipeline{
		InitEvent: initEvent,
		OnTimeout: onTimeout,
		Pipeline:  pipeline,
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

	evt := event.New(c)

	c.evtCarrier = evt
	return evt
}

func (c *Pipeline) EventType() event.Type {
	return PipelineType
}

type pipelineItem struct {
	pipelineFunc func(prev event.Event) (next event.Event)
	next         event.Event
}

func (c *Pipeline) Dispatch(bus event.Bus) {
	if len(c.Pipeline) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

	go func() {
		defer cancel()

		workload := make(chan pipelineItem, 1)
		defer close(workload)

		var currIdx int
		workload <- pipelineItem{
			pipelineFunc: c.Pipeline[currIdx],
			next:         c.InitEvent,
		}

		for {
			select {
			case wl := <-workload:
				finished := make(chan event.Event)

				sub := bus.Subscribe().
					On(event.IsFollowupOf(wl.next), func(received event.Event) {
						if c.completionCondition(wl.next, received) {
							finished <- received
						}
					})
				sub.ListenWithWorkers(1)
				bus.Publish(wl.next)

				select {
				case prev := <-finished:
					newNext := wl.pipelineFunc(prev)

					if stop, ok := newNext.Payload().(PipelineStop); ok {
						if lastEvt := stop.Event; lastEvt != nil {
							bus.Publish(lastEvt)
						}
						bus.Unsubscribe(sub)
						close(finished)
						return
					}

					currIdx++
					if currIdx == len(c.Pipeline) {
						bus.Publish(newNext)
						bus.Unsubscribe(sub)
						close(finished)
						return
					}

					workload <- pipelineItem{
						pipelineFunc: c.Pipeline[currIdx],
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
	}()
}
