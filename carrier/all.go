package carrier

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

// All is a carrier that emits all carried events on the bus.
// The carried events order is not preserved. They are dispatched in parallel (under a max concurrency threshold).
type All struct {
	Carried             []event.Event
	DoneEventFactory    func(received []event.Event) event.Event
	OnTimeout           event.Event
	CompletionCondition CompletionCondition
	maxConcurrency      int
	timeout             time.Duration
}

var (
	_ Carrier = (*All)(nil)
)

// NewAll creates a new event carrier that dispatches all events in the given slice to the event Bus.
// All carried events must have unique Ref (that means they must not be followup from each other), otherwise the behavior is undefined.
// This event carrier has a blocking Dispatch method.
func NewAll(
	carried []event.Event,
	doneEventFactory func(received []event.Event) event.Event,
	onTimeout event.Event,
	opts ...Option,
) event.Event {
	var uniqueRefset = make(map[string]struct{})
	for _, evt := range carried {
		if _, exists := uniqueRefset[evt.Ref]; exists {
			log.Printf("duplicate event ref: %s, undefined behaviour mey will append", evt.Ref)
			continue
		}
		uniqueRefset[evt.Ref] = struct{}{}
	}

	c := &All{
		Carried:          carried,
		DoneEventFactory: doneEventFactory,
		OnTimeout:        onTimeout,
	}

	cfg := &carrierConfig{
		maxConcurrency:      defaultCarrierConcurrency,
		timeout:             defaultCarrierTimeout,
		completionCondition: CompletedOnFollowupReceived,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	c.maxConcurrency = cfg.maxConcurrency
	c.timeout = cfg.timeout
	c.CompletionCondition = cfg.completionCondition

	return event.New(c)
}

func (c *All) Dispatch(bus event.Bus) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	evtProcessed := make(map[string]bool)
	evtByRef := make(map[string]event.Event)
	receivedEvents := make([]event.Event, 0, len(c.Carried))
	for _, evt := range c.Carried {
		evtByRef[evt.Ref] = evt
	}
	var mu sync.Mutex

	workload := make(chan event.Event)
	for range c.maxConcurrency {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case evt, ok := <-workload:
					if !ok {
						return
					}
					mu.Lock()
					evtProcessed[evt.Ref] = false
					mu.Unlock()
					bus.Publish(evt) //TODO; won't prevent to overwhelming the event bus
				}
			}
		}()
	}

	sub := bus.Subscribe().
		On(event.IsFollowupOf(c.Carried...), func(received event.Event) {
			mu.Lock()
			if processed, ok := evtProcessed[received.Ref]; ok &&
				!processed &&
				c.CompletionCondition(evtByRef[received.Ref], received) {
				evtProcessed[received.Ref] = true
				receivedEvents = append(receivedEvents, received)
			}
			mu.Unlock()
		})
	sub.ListenWithWorkers(1)
	defer sub.Detach()

	var done bool
	for _, evt := range c.Carried {
		select {
		case <-ctx.Done():
			done = true
		case workload <- evt:
		}
		if done {
			break
		}
	}
	close(workload)

	// Wait for completion or timeout
	t := time.NewTicker(10 * time.Millisecond) // polling may not be the better option...
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			bus.Publish(c.OnTimeout)
			return
		case <-t.C:
			if len(evtProcessed) == len(c.Carried) && allEventsHasBeenProcessed(evtProcessed) {
				bus.Publish(c.DoneEventFactory(receivedEvents))
				return
			}
		}
	}
}

func (c *All) Type() event.Type {
	return TypePrefix + ".all"
}

func allEventsHasBeenProcessed(eventMap map[string]bool) bool {
	for _, processed := range eventMap {
		if !processed {
			return false
		}
	}
	return true
}
