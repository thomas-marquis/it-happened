# Basic Publish/Subscribe

Learn how to publish events and subscribe to them using the it-happened library. This tutorial covers the fundamental pattern of event-driven architecture: publish-subscribe.

## What You'll Learn

- How to define event payloads
- How to create an event bus
- How to subscribe to specific event types
- How to publish events to the bus
- How events are matched to subscribers

## Prerequisites

- Go 1.25+ installed
- Basic understanding of Go interfaces
- Completed the [Quick Start](../index.md) guide

## Step 1: Define Your Event Payload

First, create a type that implements the `event.Payload` interface. The payload contains your event data.

```go
type MessagePayload struct {
    Content string
}

func (p MessagePayload) EventType() event.Type {
    return "message.created"
}
```

Every payload must implement `EventType()` which returns a `event.Type` - this is how events are categorized.

## Step 2: Create the Event Bus

The bus is the central hub for all event communication.

```go
done := make(chan struct{})
defer close(done)

bus := inmemory.NewBus(done, nil)
```

The `done` channel controls the bus lifetime. When closed, the bus will shut down gracefully.

## Step 3: Create a Subscriber

Subscribers receive events from the bus and can register callbacks for specific event types.

```go
sub := bus.Subscribe()
```

## Step 4: Register a Callback

Use matchers to filter which events your callback receives. Here, we use `event.Is()` to match a specific event type.

```go
sub.On(event.Is("message.created"), func(e event.Event) {
    if payload, ok := e.Payload().(MessagePayload); ok {
        fmt.Printf("Received: %s\n", payload.Content)
    }
})
```

The callback receives the full `event.Event` interface, from which you can extract the payload.

## Step 5: Start Listening

Before publishing events, start the subscriber's workers:

```go
sub.ListenWithWorkers(1)
```

This starts 1 worker goroutine to process incoming events. Use more workers for higher throughput.

## Step 6: Publish Events

Now you can publish events to the bus:

```go
evt := event.New(MessagePayload{Content: "Hello, World!"})
bus.Publish(evt)
```

The event will be delivered to all subscribers that have matching callbacks.

## Complete Example

See the complete, runnable example:

📁 [examples/basic-pubsub/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/basic-pubsub/main.go)

To run it:

```bash
cd examples/basic-pubsub
go run main.go
```

## Key Concepts Used

- **Event**: The fundamental unit of information (see [Concepts](../concepts.md#event))
- **Payload**: The data carried by an event (see [Concepts](../concepts.md#payload))
- **Bus**: The central communication hub (see [Concepts](../concepts.md#bus))
- **Subscriber**: Receives and processes events (see [Concepts](../concepts.md#subscriber))
- **Matcher**: Filters events for subscribers (see [Concepts](../concepts.md#matcher))
- **Type**: Categorizes events (see [Concepts](../concepts.md#type))

## Next Steps

- [Event Chaining](event-chaining.md) - Learn how to create sequences of related events
- [Using Matchers](using-matchers.md) - Explore different matching strategies
- [Using Carriers](using-carriers.md) - Dispatch multiple events as a unit