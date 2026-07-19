package main

import (
	"context"
	"fmt"
	"time"

	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// SimplePayload is a basic event payload for demonstration.
type SimplePayload struct {
	Name string
}

func (p SimplePayload) EventType() event.Type {
	return "demo.event"
}

// DonePayload indicates completion.
type DonePayload struct {
	Count int
}

func (p DonePayload) EventType() event.Type {
	return "demo.done"
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bus := inmemory.NewBus(ctx)

	// Subscribe to demo events
	sub := bus.Subscribe()

	sub.On(event.Is("demo.event"), func(e event.Event) {
		if payload, ok := e.Payload().(SimplePayload); ok {
			fmt.Printf("  Received: %s\n", payload.Name)
		}
		// Emit followup so carrier knows the event is complete
		// Only emit for original events (not followups) to avoid infinite loop
		if e.ChainPosition() == 0 {
			bus.Publish(e.NewFollowup(e.Payload()))
		}
	})

	sub.On(event.Is("demo.done"), func(e event.Event) {
		if payload, ok := e.Payload().(DonePayload); ok {
			fmt.Printf("Done: processed %d events\n", payload.Count)
		}
	})

	sub.ListenWithWorkers(1)

	// Create some simple events to carry
	events := []event.Event{
		event.New(SimplePayload{Name: "Event 1"}),
		event.New(SimplePayload{Name: "Event 2"}),
		event.New(SimplePayload{Name: "Event 3"}),
	}

	fmt.Println("=== All Carrier (Parallel Dispatch) ===")
	fmt.Println("Publishing carrier with 3 events...")

	allCarrier := carrier.NewAll(
		events,
		func(evtCarrier event.Event, received []event.Event) event.Event {
			// This function is called when all carried events are completed
			// For this demo, we'll just return a done event
			return event.New(DonePayload{Count: len(received)})

			// You also can use the evtCarrier reference itself to return a followup event:
			// return evtCarrier.NewFollowup(DonePayload{Count: len(received)})
			// This way, it's possible to chain a carrier with other carriers (for example, multiple `all` carriers into a `sequence`)
		},
		event.New(SimplePayload{Name: "Timeout event"}), // This would be published on timeout
		carrier.WithMaxConcurrency(2),
		carrier.WithTimeout(2*time.Second),
	)

	bus.Publish(allCarrier)

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Note: All carrier dispatches all events in parallel.")
}
