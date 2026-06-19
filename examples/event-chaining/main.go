package main

import (
	"fmt"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// OrderCreatedPayload represents an order creation event.
type OrderCreatedPayload struct {
	OrderID string
	Amount  float64
}

func (p OrderCreatedPayload) EventType() event.Type {
	return "order.created"
}

// OrderProcessedPayload represents an order processing event.
type OrderProcessedPayload struct {
	OrderID string
	Status  string
}

func (p OrderProcessedPayload) EventType() event.Type {
	return "order.processed"
}

// OrderCompletedPayload represents an order completion event.
type OrderCompletedPayload struct {
	OrderID string
	Total   float64
}

func (p OrderCompletedPayload) EventType() event.Type {
	return "order.completed"
}

func main() {
	done := make(chan struct{})
	defer close(done)

	bus := inmemory.NewBus(done, nil)

	// Subscribe to order events
	sub := bus.Subscribe()

	// Handle order created events and create followup processing events
	sub.On(event.Is("order.created"), func(e event.Event) {
		if payload, ok := e.Payload().(OrderCreatedPayload); ok {
			fmt.Printf("Order created: %s for $%.2f\n", payload.OrderID, payload.Amount)

			// Create a followup event for processing
			// The event returned by event.New() is a ChainableEvent
			if chainableEvt, ok := e.(event.ChainableEvent); ok {
				processingEvt := chainableEvt.NewFollowup(
					OrderProcessedPayload{
						OrderID: payload.OrderID,
						Status:  "processing",
					},
				)
				bus.Publish(processingEvt)
			}
		}
	})

	// Handle order processed events and create followup completion events
	sub.On(event.Is("order.processed"), func(e event.Event) {
		if payload, ok := e.Payload().(OrderProcessedPayload); ok {
			fmt.Printf("Order processed: %s with status: %s\n", payload.OrderID, payload.Status)

			// Create a followup event for completion
			if chainableEvt, ok := e.(event.ChainableEvent); ok {
				completionEvt := chainableEvt.NewFollowup(
					OrderCompletedPayload{
						OrderID: payload.OrderID,
						Total:   100.0, // In a real app, this would come from the processing
					},
				)
				bus.Publish(completionEvt)
			}
		}
	})

	// Handle order completed events
	sub.On(event.Is("order.completed"), func(e event.Event) {
		if payload, ok := e.Payload().(OrderCompletedPayload); ok {
			fmt.Printf("Order completed: %s for total $%.2f\n", payload.OrderID, payload.Total)
		}
	})

	sub.ListenWithWorkers(1)

	// Publish an order created event (starts the chain)
	orderEvt := event.New(OrderCreatedPayload{
		OrderID: "ORD-12345",
		Amount:  99.99,
	})

	fmt.Printf("Starting order workflow for: %s\n", orderEvt.ID())
	bus.Publish(orderEvt)

	// Wait for the chain to complete
	time.Sleep(200 * time.Millisecond)

	fmt.Println("Order workflow completed!")
}
