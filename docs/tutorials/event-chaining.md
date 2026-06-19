# Event Chaining

Learn how to create sequences of related events using the chain mechanism. Event chaining allows you to track workflows and processes that span multiple steps across your application.

## What You'll Learn

- How event chaining works
- How to create followup events
- How ChainRef links events together
- How ChainPosition tracks progress through a chain
- Practical use cases for event chaining

## Prerequisites

- Go 1.25+ installed
- Completed the [Quick Start](../index.md) guide
- Understand basic publish/subscribe (see [Basic Pub/Sub](basic-pubsub.md))
- Familiar with the [Chain](concepts.md#chain), [ChainRef](concepts.md#chainref), [ChainPosition](concepts.md#chainposition), and [Followup](concepts.md#followup) concepts

## How Event Chaining Works

In it-happened, events can be linked together in chains. A chain is a sequence of related events that share a common reference (ChainRef). Each event in the chain has a position (ChainPosition) that indicates its place in the sequence.

The first event in a chain has ChainPosition 0. When you create a followup event from it, the followup shares the same ChainRef but has ChainPosition 1, and so on.

## Step 1: Create the Initial Event

Start by creating the first event in your chain:

```go
event.New(OrderCreatedPayload{
    OrderID: "ORD-12345",
    Amount:  99.99,
})
```

This creates an event with a unique ID. Since no ChainRef was specified, the event's ID becomes its ChainRef, and its ChainPosition is 0.

## Step 2: Create a Followup Event

When handling an event, you can create a followup that continues the chain:

```go
sub.On(event.Is("order.created"), func(e event.Event) {
    // Check if the event supports chaining
    if chainableEvt, ok := e.(event.ChainableEvent); ok {
        // Create a followup event
        followupEvt := chainableEvt.NewFollowup(
            OrderProcessedPayload{
                OrderID: payload.OrderID,
                Status:  "processing",
            },
        )
        bus.Publish(followupEvt)
    }
})
```

The `NewFollowup` method creates a new event that:
- Shares the same ChainRef as the parent event
- Has ChainPosition incremented by 1
- Carries the new payload you provide

## Step 3: Continue the Chain

You can continue creating followups to build longer chains:

```go
sub.On(event.Is("order.processed"), func(e event.Event) {
    if chainableEvt, ok := e.(event.ChainableEvent); ok {
        completionEvt := chainableEvt.NewFollowup(
            OrderCompletedPayload{
                OrderID: payload.OrderID,
                Total:   100.0,
            },
        )
        bus.Publish(completionEvt)
    }
})
```

## Step 4: Track Chain Progress

You can inspect the chain information on any event:

```go
sub.On(event.IsOneOf("order.created", "order.processed", "order.completed"), func(e event.Event) {
    if chainableEvt, ok := e.(event.ChainableEvent); ok {
        fmt.Printf("ChainRef: %s, Position: %d\n", 
            chainableEvt.ChainRef(), 
            chainableEvt.ChainPosition())
    }
})
```

This will show that all events in the chain share the same ChainRef, but each has a unique ChainPosition.

## Complete Example

See the complete, runnable example:

📁 [examples/event-chaining/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/event-chaining/main.go)

To run it:

```bash
cd examples/event-chaining
go run main.go
```

Expected output:

```
Starting order workflow for: <event-id>
Order created: ORD-12345 for $99.99
Order processed: ORD-12345 with status: processing
Order completed: ORD-12345 for total $100.00
Order workflow completed!
```

## Key Concepts Used

- **Chain**: A sequence of related events (see [Concepts](../concepts.md#chain))
- **ChainRef**: The unique identifier linking all events in a chain (see [Concepts](../concepts.md#chainref))
- **ChainPosition**: The position of an event within its chain (see [Concepts](../concepts.md#chainposition))
- **Followup**: A new event created from a parent event in a chain (see [Concepts](../concepts.md#followup))
- **ChainableEvent**: An event that supports chaining (see [Concepts](../concepts.md#chainableevent))

## Real-World Use Cases

Event chaining is useful for:

1. **Multi-step workflows**: Track the progress of a business process through multiple stages
2. **Request-response patterns**: Match responses to their original requests
3. **Saga pattern**: Manage distributed transactions across multiple services
4. **Audit trails**: Maintain a complete history of related actions
5. **State machines**: Model state transitions as a series of events

## Next Steps

- [Basic Publish/Subscribe](basic-pubsub.md) - Review the fundamentals
- [Using Matchers](using-matchers.md) - Learn advanced event filtering
- [Using Carriers](using-carriers.md) - Dispatch multiple events as a unit