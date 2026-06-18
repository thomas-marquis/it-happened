package carrier

import (
	"context"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

// Sequence is a carrier that emits a sequence of events on the bus.
// The next event is emitted only when the previous one has been received/resolved/.
type Sequence struct {
	Carried             []event.ChainableEvent
	DoneEventFactory    func(received []event.Event) event.Event
	OnTimeout           event.Event
	CompletionCondition CompletionCondition

	timeout time.Duration
}

func NewSequence(carried []event.ChainableEvent, doneEventFactory func(received []event.Event) event.Event, onTimeout event.Event, opts ...Option) event.Event {
	c := &Sequence{
		Carried:          carried,
		DoneEventFactory: doneEventFactory,
		OnTimeout:        onTimeout,
	}

	cfg := &carrierConfig{
		timeout:             defaultCarrierTimeout,
		completionCondition: CompletedOnFollowupReceived,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	c.timeout = cfg.timeout
	c.CompletionCondition = cfg.completionCondition

	return event.New(c)
}

func (c *Sequence) EventType() event.Type {
	return TypePrefix + ".sequence"
}

func (c *Sequence) Dispatch(bus event.Bus) {
	if len(c.Carried) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

	go func() {
		defer cancel()
		receivedEvents := c.doDispatch(ctx, bus)
		bus.Publish(c.DoneEventFactory(receivedEvents))
	}()
}

func (c *Sequence) doDispatch(ctx context.Context, bus event.Bus) (receivedEvents []event.Event) {
	workload := make(chan event.ChainableEvent, 1)
	defer close(workload)

	var currIdx int
	workload <- c.Carried[currIdx]

	var mu sync.Mutex

	for {
		select {
		case evt := <-workload:
			finished := make(chan struct{})
			sub := bus.Subscribe().
				On(event.IsFollowupOf(evt), func(received event.Event) {
					if c.CompletionCondition(evt, received) {
						mu.Lock()
						defer mu.Unlock()
						receivedEvents = append(receivedEvents, received)
						close(finished)
					}
				})
			sub.ListenWithWorkers(1)
			bus.Publish(evt)

			select {
			case <-finished:
				currIdx++
				if currIdx == len(c.Carried) {
					sub.Detach()
					return
				}
				workload <- c.Carried[currIdx]
			case <-ctx.Done():
				bus.Publish(c.OnTimeout)
				sub.Detach()
				return
			}

			sub.Detach()
		case <-ctx.Done():
			return
		}
	}
}
