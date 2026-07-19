# Using Carriers

Learn how to use carriers to dispatch multiple events as a single unit. Carriers are powerful tools for orchestrating complex workflows and managing batches of related events.

## What You'll Learn

- What carriers are and when to use them
- The difference between All, Sequence, and Pipeline carriers
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

The library provides three built-in carrier implementations:

### All Carrier

The `All` carrier dispatches all carried events in parallel (up to a maximum concurrency limit) and waits for all of them to be completed. This is ideal for batch processing where order doesn't matter.

### Sequence Carrier

The `Sequence` carrier dispatches carried events one at a time, waiting for each event to be completed before dispatching the next one. This is ideal for workflows where order matters.

### Pipeline Carrier

The `Pipeline` carrier executes a sequence of transformation functions, where each function takes the previous event's completion as input and returns the next event to be processed. This is ideal for event processing pipelines where each stage transforms or processes the event and passes it to the next stage. The pipeline can be interrupted early using a `PipelineStop` payload.

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

The done event factory is a function that creates the completion event when all carried events are processed.
This function's fist argument is the event carrier pointer, that can be used to chaining with the done event (it is optional).

```go
func doneFactory(evtCarrier event.Event, received []event.Event) event.Event {
    return evtCarrier.NewFollowup(BatchResultPayload{
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
4. If timeout occurs, cancel all remaining carried events and publish the timeout event

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
2. Wait for it to be completed (typically, when the corresponding followup event has been received)
3. Dispatch the second event
4. Continue until all events are processed
5. Publish the done event from the factory
6. If timeout occurs, cancel all remaining carried events and publish the timeout event

## Step 5: Create and Publish a Pipeline Carrier

The Pipeline carrier is different from All and Sequence as it uses functions to transform events through a pipeline. Each function receives the completion event from the previous stage and returns the next event to process.

```go
// Define the pipeline stages as functions
pipelineStages := []func(prev event.Event) event.Event{
    func(prev event.Event) event.Event {
        // Stage 1: Transform the input event
        return event.New(ProcessedPayload{Message: "Stage 1 processed: " + prev.Payload().(InputPayload).Data})
    },
    func(prev event.Event) event.Event {
        // Stage 2: Further transformation
        return event.New(ProcessedPayload{Message: prev.Payload().(ProcessedPayload).Message + " -> Stage 2"})
    },
    func(prev event.Event) event.Event {
        // Stage 3: Final transformation
        return event.New(FinalPayload{Result: prev.Payload().(ProcessedPayload).Message + " -> Stage 3"})
    },
}

// Create the initial event to start the pipeline
initEvent := event.New(InputPayload{Data: "initial data"})

// Create a timeout event
timeoutEvent := event.New(TimeoutPayload{Message: "Pipeline timed out"})

// Create the pipeline carrier
pipelineCarrier := carrier.NewPipeline(
    initEvent,
    pipelineStages,
    timeoutEvent,
    carrier.WithTimeout(5*time.Second),
)

bus.Publish(pipelineCarrier)
```

The `Pipeline` carrier will:
1. Publish the initEvent
2. Wait for it to be completed
3. Pass the completion event to the first pipeline function
4. Publish the returned event
5. Wait for completion and pass to the next function
6. Continue until all pipeline stages are executed
7. If timeout occurs, publish the timeout event
8. If a pipeline function returns a `PipelineStop` payload, the pipeline is interrupted and the wrapped event (if any) is published

## Complete Examples

Each carrier has its own dedicated example:

### All Carrier Example

📁 [examples/using-all-carrier/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/using-all-carrier/main.go)

To run it:

```bash
cd examples/using-all-carrier
go run main.go
```

### Sequence Carrier Example

📁 [examples/using-sequence-carrier/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/using-sequence-carrier/main.go)

To run it:

```bash
cd examples/using-sequence-carrier
go run main.go
```

### Pipeline Carrier Example

📁 [examples/using-pipeline-carrier/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/using-pipeline-carrier/main.go)

To run it:

```bash
cd examples/using-pipeline-carrier
go run main.go
```

## How Completion Works

Carriers use a `CompletionCondition` to determine when a dispatched event is considered complete.
By default, they wait for followup events that share the same ChainRef as the dispatched event.

When you publish a carrier:
1. The carrier dispatches its events to the bus (all at once for All, sequentially for Sequence and Pipeline)
2. For each dispatched event, it listens for followup events
3. When a followup event matches the completion condition, it's counted as complete
4. For All and Sequence carriers: When all carried events are complete (or timeout occurs), the carrier publishes the done event
5. For Pipeline carrier: When all pipeline stages are complete (or timeout occurs), the pipeline stops; if interrupted by PipelineStop, the wrapped event is published

## Carrier Configuration Options

All carriers support these configuration options:

- `carrier.WithTimeout(d)` - Set the timeout duration
- `carrier.WithCompletionCondition(cond)` - Set a custom completion condition

The All carrier additionally supports:

- `carrier.WithMaxConcurrency(n)` - Set the maximum number of concurrent operations (All carrier only)

## Real-World Use Cases

Carriers are useful for:

1. **Batch Processing**: Send multiple notifications, process multiple orders, etc. (All carrier)
2. **Workflow Orchestration**: Coordinate multi-step processes across services (Sequence carrier)
3. **Data Pipeline**: Process data through multiple transformation stages (Pipeline carrier)
4. **Fan-out/Fan-in**: Dispatch multiple parallel tasks and wait for all to complete (All carrier)
5. **Saga Pattern**: Manage distributed transactions with compensation logic (Sequence or Pipeline carriers)
6. **ETL Processes**: Extract, transform, and load data through a series of steps (Pipeline carrier)

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