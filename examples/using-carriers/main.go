package main

import (
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
	done := make(chan struct{})
	defer close(done)

	bus := inmemory.NewBus(done, nil)

	// Subscribe to demo events
	sub := bus.Subscribe()

	// Count how many demo events we receive
	count := 0
	sub.On(event.Is("demo.event"), func(e event.Event) {
		count++
		if payload, ok := e.Payload().(SimplePayload); ok {
			fmt.Printf("  Received: %s\n", payload.Name)
		}
	})

	sub.On(event.Is("demo.done"), func(e event.Event) {
		if payload, ok := e.Payload().(DonePayload); ok {
			fmt.Printf("Done: processed %d events\n", payload.Count)
		}
	})

	sub.ListenWithWorkers(1)

	// Create some simple events to carry
	events := []event.Payload{
		SimplePayload{Name: "Event 1"},
		SimplePayload{Name: "Event 2"},
		SimplePayload{Name: "Event 3"},
	}

	// Convert to ChainableEvent
	var carriedEvents []event.ChainableEvent
	for _, e := range events {
		carriedEvents = append(carriedEvents, event.New(e))
	}

	// Example: Using All carrier to dispatch all events
	// All carrier dispatches all carried events in parallel (up to max concurrency)
	fmt.Println("=== All Carrier (Parallel Dispatch) ===")
	fmt.Println("Publishing carrier with 3 events...")

	allCarrier := carrier.NewAll(
		carriedEvents,
		func(received []event.Event) event.Event {
			// This function is called when all carried events are completed
			// For this demo, we'll just return a done event
			return event.New(DonePayload{Count: len(received)})
		},
		event.New(SimplePayload{Name: "Timeout event"}), // This would be published on timeout
		carrier.WithMaxConcurrency(2),
		carrier.WithTimeout(2*time.Second),
	)

	bus.Publish(allCarrier)

	// Wait a moment for events to be processed
	time.Sleep(100 * time.Millisecond)

	// Reset counter for sequence demo
	count = 0

	// Create fresh events for sequence carrier
	var carriedEvents2 []event.ChainableEvent
	for _, e := range events {
		carriedEvents2 = append(carriedEvents2, event.New(e))
	}

	// Example: Using Sequence carrier to dispatch events one at a time
	fmt.Println("\n=== Sequence Carrier (Sequential Dispatch) ===")
	fmt.Println("Publishing carrier with 3 events...")

	sequenceCarrier := carrier.NewSequence(
		carriedEvents2,
		func(received []event.Event) event.Event {
			return event.New(DonePayload{Count: len(received)})
		},
		event.New(SimplePayload{Name: "Timeout event"}),
		carrier.WithTimeout(2*time.Second),
	)

	bus.Publish(sequenceCarrier)

	// Wait for processing
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== Demo Complete ===")
	fmt.Println("Note: Carriers dispatch multiple events as a single unit.")
	fmt.Println("      All carrier: parallel dispatch")
	fmt.Println("      Sequence carrier: sequential dispatch")
}
