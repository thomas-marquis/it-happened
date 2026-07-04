package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// MessagePayload is a simple payload for our messages.
type MessagePayload struct {
	Content string
}

// EventType implements the event.Payload interface.
func (p MessagePayload) EventType() event.Type {
	return "message.created"
}

// ActivityNotifier demonstrates how to create a custom notifier that tracks all bus activity.
type ActivityNotifier struct {
	eventCount int
	mu         sync.Mutex
}

// NotifyPublished is called whenever an event is published to the bus.
func (n *ActivityNotifier) NotifyPublished(evt event.Event) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.eventCount++
	fmt.Printf("[NOTIFIER] Event #%d published: type=%s, id=%s\n", n.eventCount, evt.Type(), evt.ID())
}

// NotifySubscribed is called whenever a new subscriber is created.
func (n *ActivityNotifier) NotifySubscribed(sub *event.Subscriber) {
	fmt.Printf("[NOTIFIER] Subscriber joined the bus\n")
}

// NotifyUnsubscribed is called whenever a subscriber is removed.
func (n *ActivityNotifier) NotifyUnsubscribed(sub *event.Subscriber) {
	fmt.Printf("[NOTIFIER] Subscriber left the bus\n")
}

func main() {
	fmt.Println("=== it-happened: Custom Notifiers Example ===")
	fmt.Println()

	// Create context to control the bus lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create our custom notifier that embeds NopNotifier
	activityNotifier := &ActivityNotifier{}

	// Create the bus with our custom notifier
	// You can only have one notifier per bus, but you can create a composite notifier
	// that delegates to multiple notifiers if needed
	bus := inmemory.NewBus(ctx, inmemory.WithNotifier(activityNotifier))

	fmt.Println("Setting up subscribers...")

	// Create first subscriber
	sub1 := bus.Subscribe()
	sub1.On(event.Is("message.created"), func(evt event.Event) {
		if payload, ok := evt.Payload().(MessagePayload); ok {
			fmt.Printf("  [SUB1] Received: %q\n", payload.Content)
		}
	})
	sub1.ListenWithWorkers(1)
	defer sub1.Detach()

	// Create second subscriber
	sub2 := bus.Subscribe()
	sub2.On(event.IsAny(), func(evt event.Event) {
		fmt.Printf("  [SUB2] All events receiver: type=%s\n", evt.Type())
	})
	sub2.ListenWithWorkers(1)
	defer sub2.Detach()

	fmt.Println()
	fmt.Println("Publishing events...")

	// Publish some messages
	messages := []string{
		"Hello, World!",
		"This is a test message",
		"Event-driven architecture is powerful!",
	}

	for i, msg := range messages {
		fmt.Printf("Publishing message %d: %q\n", i+1, msg)
		evt := event.New(MessagePayload{Content: msg})
		bus.Publish(evt)
	}

	fmt.Println()
	fmt.Println("Unsubscribing first subscriber...")

	// Unsubscribe the first subscriber
	bus.Unsubscribe(sub1)

	fmt.Println()
	fmt.Println("Publishing more events after unsubscribe...")

	// Publish more messages after unsubscribing sub1
	moreMessages := []string{
		"This message won't be received by sub1",
		"But sub2 will still receive it",
	}

	for i, msg := range moreMessages {
		fmt.Printf("Publishing message %d: %q\n", i+4, msg)
		evt := event.New(MessagePayload{Content: msg})
		bus.Publish(evt)
	}

	fmt.Println()
	fmt.Println("Creating a new subscriber...")

	// Create a third subscriber
	sub3 := bus.Subscribe()
	sub3.On(event.Is("message.created"), func(evt event.Event) {
		if payload, ok := evt.Payload().(MessagePayload); ok {
			fmt.Printf("  [SUB3] Received: %q\n", payload.Content)
		}
	})
	sub3.ListenWithWorkers(1)
	defer sub3.Detach()

	fmt.Println()
	fmt.Println("Publishing final message...")

	// Publish one final message
	finalMsg := "Final message - all active subscribers will receive this"
	fmt.Printf("Publishing: %q\n", finalMsg)
	bus.Publish(event.New(MessagePayload{Content: finalMsg}))

	// Wait for all events to be processed
	time.Sleep(200 * time.Millisecond)

	fmt.Println()
	fmt.Println("=== Summary ===")
	fmt.Printf("Total events published: %d\n", activityNotifier.eventCount)

	// Note: metricsNotifier wasn't used because we can only have one notifier per bus
	// In a real application, you could create a composite notifier that delegates to multiple notifiers
	fmt.Println()
	fmt.Println("Example completed!")
}
