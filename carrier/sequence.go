package carrier

import (
	"context"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

const SequenceType event.Type = TypePrefix + ".sequence"

// Sequence is a carrier that emits a sequence of events on the bus.
// The next event is emitted only when the previous one has been received/resolved.
// This ensures ordered processing of events in the sequence.
type Sequence struct {
	// Carried contains the events to be dispatched in sequence.
	Carried []event.Event `json:"carried"`
	// DoneEventFactory creates the completion event when all carried events are processed.
	DoneEventFactory DoneFactory `json:"-"`
	// OnTimeout is the event to publish if the sequence times out.
	OnTimeout event.Event `json:"onTimeout,omitempty"`
	// CompletionCondition determines when a sent event is considered complete.
	CompletionCondition CompletionCondition `json:"-"`

	timeout    time.Duration
	evtCarrier event.Event
}

// NewSequence creates a new Sequence carrier that dispatches events in order.
//
// The events are dispatched one at a time, with each subsequent event only being
// dispatched after the previous one has been completed according to the completion condition.
//
// Parameters:
//
//	carried - The events to dispatch in sequence
//	doneEventFactory - Function to create the completion event
//	onTimeout - Event to publish if the sequence times out
//	opts - Optional configuration options
//
// Returns:
//
//	A new event that wraps the Sequence carrier
func NewSequence(
	carried []event.Event,
	doneEventFactory DoneFactory,
	onTimeout event.Event,
	opts ...Option,
) event.Event {
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
	c.evtCarrier = event.New(c)

	return c.evtCarrier
}

// EventType returns the event type for Sequence carrier events.
// All Sequence carriers have the same type prefix.
func (c *Sequence) EventType() event.Type {
	return SequenceType
}

// Dispatch implements the Carrier interface for Sequence.
//
// It starts dispatching the carried events in sequence, waiting for each event
// to be completed before dispatching the next one. When all events are processed
// or a timeout occurs, it publishes the appropriate completion event.
//
// Parameters:
//
//	bus - The event bus to dispatch events to
func (c *Sequence) Dispatch(bus event.Bus) {
	if len(c.Carried) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)

	go func() {
		defer cancel()
		receivedEvents := c.doDispatch(ctx, bus)
		doneEvt := c.DoneEventFactory(c.evtCarrier, receivedEvents)
		if doneEvt != nil {
			bus.Publish(doneEvt)
		}
	}()
}

// doDispatch handles the actual sequential dispatching of events.
// It returns a slice of all received events that matched the completion condition.
func (c *Sequence) doDispatch(ctx context.Context, bus event.Bus) (receivedEvents []event.Event) {
	workload := make(chan event.Event, 1)
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
					bus.Unsubscribe(sub)
					return
				}
				workload <- c.Carried[currIdx]
			case <-ctx.Done():
				bus.Publish(c.OnTimeout)
			}

			bus.Unsubscribe(sub)
		case <-ctx.Done():
			return
		}
	}
}
