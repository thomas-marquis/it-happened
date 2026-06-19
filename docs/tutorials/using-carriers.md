# Using Carriers

Learn how to use carriers to dispatch multiple events as a single unit. Carriers are powerful tools for orchestrating complex workflows and managing batches of related events.

## What You'll Learn

- What carriers are and when to use them
- The difference between All and Sequence carriers
- How to create and use carriers
- How carriers handle completion and timeouts
- Practical use cases for carriers

## Prerequisites

- Go 1.25+ installed
- Completed the [Quick Start](../index.md) guide
- Understand basic publish/subscribe (see [Basic Pub/Sub](basic-pubsub.md))
- Familiar with the [Carrier](concepts.md#carrier) and [CompletionCondition](concepts.md#completioncondition) concepts

## What is a Carrier?

A Carrier is a special type of event payload that can dispatch multiple other events to the bus. It acts as an orchestrator, managing a group of events and their lifecycle.

Carriers are useful when you need to:
- Publish multiple related events as a single unit
- Ensure all events in a batch are processed
- Handle completion or timeout for a group of events
- Coordinate complex workflows

## Types of Carriers

The library provides two built-in carrier implementations:

### All Carrier

The `All` carrier dispatches all carried events in parallel (up to a maximum concurrency limit) and waits for all of them to be completed. This is ideal for batch processing where order doesn't matter.

### Sequence Carrier

The `Sequence` carrier dispatches carried events one at a time, waiting for each event to be completed before dispatching the next one. This is ideal for workflows where order matters.

## Step 1: Create Events to Carry

First, create the events you want the carrier to dispatch:

```go
// Define your payload type
type NotificationPayload struct {
    Message string
}

func (p NotificationPayload) EventType() event.Type {
    return "notification.sent"
}

// Create the events
events := []event.Event{
    event.New(NotificationPayload{Message: "Welcome email"}),
    event.New(NotificationPayload{Message: "Password reset"}),
    event.New(NotificationPayload{Message: "Account verified"}),
}
```

## Step 2: Create a Done Event Factory

The done event factory is a function that creates the completion event when all carried events are processed:

```go
func doneFactory(received []event.Event) event.Event {
    return event.New(BatchResultPayload{
        Processed: len(received),
        Total:     len(events),
    })
}
```

## Step 3: Create and Publish an All Carrier

```go
allCarrier := carrier.NewAll(
    events,
    doneFactory,
    event.New(TimeoutPayload{Message: "Carrier timed out"}),
    carrier.WithMaxConcurrency(2),  // Process 2 at a time
    carrier.WithTimeout(5*time.Second),
)

bus.Publish(allCarrier)
```

The `All` carrier will:
1. Dispatch all carried events in parallel (up to 2 at a time)
2. Wait for all events to be completed
3. Publish the done event from the factory
4. If timeout occurs, publish the timeout event

## Step 4: Create and Publish a Sequence Carrier

```go
sequenceCarrier := carrier.NewSequence(
    events,
    doneFactory,
    event.New(TimeoutPayload{Message: "Sequence timed out"}),
    carrier.WithTimeout(5*time.Second),
)

bus.Publish(sequenceCarrier)
```

The `Sequence` carrier will:
1. Dispatch the first carried event
2. Wait for it to be completed
3. Dispatch the second event
4. Continue until all events are processed
5. Publish the done event from the factory
6. If timeout occurs, publish the timeout event

## Complete Example

See the complete, runnable example:

📁 [examples/using-carriers/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/using-carriers/main.go)

To run it:

```bash
cd examples/using-carriers
go run main.go
```

Expected output:

```
=== All Carrier (Parallel Dispatch) ===
Publishing carrier with 3 events...
  Received: Event 1
  Received: Event 2
  Received: Event 3
Done: processed 3 events

=== Sequence Carrier (Sequential Dispatch) ===
Publishing carrier with 3 events...
  Received: Event 1
  Received: Event 2
  Received: Event 3
Done: processed 3 events

=== Demo Complete ===
Note: Carriers dispatch multiple events as a single unit.
      All carrier: parallel dispatch
      Sequence carrier: sequential dispatch
```

## How Completion Works

Carriers use a `CompletionCondition` to determine when a carried event is considered complete. By default, they wait for followup events that share the same ChainRef as the carried event.

When you publish a carrier:
1. The carrier dispatches all carried events to the bus
2. For each carried event, it listens for followup events
3. When a followup event matches the completion condition, it's counted as complete
4. When all carried events are complete (or timeout occurs), the carrier publishes the done event

## Carrier Configuration Options

Both carriers support these configuration options:

- `carrier.WithMaxConcurrency(n)` - Set the maximum number of concurrent operations (All carrier only)
- `carrier.WithTimeout(d)` - Set the timeout duration
- `carrier.WithCompletionCondition(cond)` - Set a custom completion condition

## Real-World Use Cases

Carriers are useful for:

1. **Batch Processing**: Send multiple notifications, process multiple orders, etc.
2. **Workflow Orchestration**: Coordinate multi-step processes across services
3. **Data Pipeline**: Process data through multiple stages
4. **Fan-out/Fan-in**: Dispatch multiple parallel tasks and wait for all to complete
5. **Saga Pattern**: Manage distributed transactions with compensation logic

## Key Concepts Used

- **Carrier**: A special event that dispatches multiple events (see [Concepts](../concepts.md#carrier))
- **CompletionCondition**: Determines when a carried event is complete (see [Concepts](../concepts.md#completioncondition))
- **ChainRef**: Links related events together (see [Concepts](../concepts.md#chainref))

## Next Steps

- [Basic Publish/Subscribe](basic-pubsub.md) - Review the fundamentals
- [Event Chaining](event-chaining.md) - Learn how to create sequences of related events
- [Using Matchers](using-matchers.md) - Learn advanced event filtering

## Additional Resources

- [Carrier package documentation](https://pkg.go.dev/github.com/thomas-marquis/it-happened/carrier)