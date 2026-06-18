package main

import (
	"fmt"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// MessagePayload is a simple payload that carries a text message.
type MessagePayload struct {
	Content string
}

// EventType implements the event.Payload interface.
func (p MessagePayload) EventType() event.Type {
	return "message.created"
}

func main() {
	// Create a done channel to control the bus lifetime
	done := make(chan struct{})
	defer close(done)

	// Create an in-memory event bus
	bus := inmemory.NewBus(done, nil)

	// Create a subscriber
	sub := bus.Subscribe()

	// Register a callback for message events
	sub.On(event.Is("message.created"), func(e event.Event) {
		if payload, ok := e.Payload().(MessagePayload); ok {
			fmt.Printf("Received message: %s (Event ID: %s)\n", payload.Content, e.ID())
		}
	})

	// Start listening with 1 worker
	sub.ListenWithWorkers(1)

	// Publish some messages
	messages := []string{
		"Hello, World!",
		"This is a test message",
		"Event-driven architecture is powerful!",
	}

	for _, msg := range messages {
		evt := event.New(MessagePayload{Content: msg})
		fmt.Printf("Publishing: %s (Event ID: %s)\n", msg, evt.ID())
		bus.Publish(evt)
	}

	// Wait for all messages to be processed
	time.Sleep(100 * time.Millisecond)

	fmt.Println("All messages processed!")
}
