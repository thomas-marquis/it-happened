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

	fmt.Println("=== Sequence Carrier (Sequential Dispatch) ===")
	fmt.Println("Publishing carrier with 3 events...")

	sequenceCarrier := carrier.NewSequence(
		events,
		func(carrier event.Event, received []event.Event) event.Event {
			return event.New(DonePayload{Count: len(received)})
		},
		event.New(SimplePayload{Name: "Timeout event"}),
		carrier.WithTimeout(2*time.Second),
	)

	bus.Publish(sequenceCarrier)

	// Wait for events to be processed sequentially
	time.Sleep(400 * time.Millisecond)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Note: Sequence carrier dispatches events one at a time, waiting for each to complete.")
}
