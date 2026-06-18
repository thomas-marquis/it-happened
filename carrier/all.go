package carrier

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

// All is a carrier that emits all carried events on the bus.
// The carried events order is not preserved. They are dispatched in parallel
// under a maximum concurrency threshold.
type All struct {
	// Carried contains the events to be dispatched in parallel.
	Carried []event.ChainableEvent
	// DoneEventFactory creates the completion event when all carried events are processed.
	DoneEventFactory func(received []event.Event) event.Event
	// OnTimeout is the event to publish if the carrier times out.
	OnTimeout event.Event
	// CompletionCondition determines when a sent event is considered complete.
	CompletionCondition CompletionCondition
	maxConcurrency      int
	timeout             time.Duration
}

// Ensure All implements the Carrier interface.
var (
	_ Carrier = (*All)(nil)
)

// NewAll creates a new event carrier that dispatches all events in the given slice to the event Bus.
//
// All carried events must have unique Ref (that means they must not be followup from each other),
// otherwise the behavior is undefined.
//
// This event carrier has a blocking Dispatch method that waits for all events to be processed
// or for a timeout to occur.
//
// Parameters:
//
//	carried - The events to dispatch in parallel
//	doneEventFactory - Function to create the completion event
//	onTimeout - Event to publish if the carrier times out
//	opts - Optional configuration options
//
// Returns:
//
//	A new event that wraps the All carrier
func NewAll(
	carried []event.ChainableEvent,
	doneEventFactory func(received []event.Event) event.Event,
	onTimeout event.Event,
	opts ...Option,
) event.Event {
	var uniqueRefset = make(map[string]struct{})
	for _, evt := range carried {
		if _, exists := uniqueRefset[evt.ChainRef()]; exists {
			log.Printf("duplicate event ref: %s, undefined behaviour mey will append", evt.ChainRef())
			continue
		}
		uniqueRefset[evt.ChainRef()] = struct{}{}
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

// Dispatch implements the Carrier interface for All.
//
// It dispatch all carried events in parallel (up to maxConcurrency), waits for all
// events to be completed or for a timeout to occur, then publishes the appropriate
// completion or timeout event.
//
// Parameters:
//
//	bus - The event bus to dispatch events to
func (c *All) Dispatch(bus event.Bus) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	evtProcessed := make(map[string]bool)
	evtByRef := make(map[string]event.ChainableEvent)
	receivedEvents := make([]event.Event, 0, len(c.Carried))
	for _, evt := range c.Carried {
		evtByRef[evt.ChainRef()] = evt
	}
	var mu sync.Mutex

	workload := make(chan event.ChainableEvent)
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
					evtProcessed[evt.ChainRef()] = false
					mu.Unlock()
					bus.Publish(evt) //TODO; won't prevent to overwhelming the event bus
				}
			}
		}()
	}

	sub := bus.Subscribe().
		On(event.IsFollowupOf(c.Carried...), func(received event.Event) {
			e, ok := received.(event.ChainableEvent)
			if !ok {
				return
			}
			mu.Lock()
			if processed, ok := evtProcessed[e.ChainRef()]; ok &&
				!processed &&
				c.CompletionCondition(evtByRef[e.ChainRef()], e) {
				evtProcessed[e.ChainRef()] = true
				receivedEvents = append(receivedEvents, e)
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

// EventType returns the event type for All carrier events.
// All All carriers have the same type prefix.
func (c *All) EventType() event.Type {
	return TypePrefix + ".all"
}

// allEventsHasBeenProcessed checks if all events in the map have been processed.
//
// Parameters:
//
//	eventMap - Map of event refs to their processed status
//
// Returns:
//
//	true if all events have been processed, false otherwise
func allEventsHasBeenProcessed(eventMap map[string]bool) bool {
	for _, processed := range eventMap {
		if !processed {
			return false
		}
	}
	return true
}
