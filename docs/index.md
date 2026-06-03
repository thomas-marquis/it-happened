# it-happened

A library that simplify event-driven development.

<figure markdown="span">
  ![Logo](assets/images/logo-tr.png)
</figure>

## Requirements

- Go 1.25 or higher

## Installation

```bash
go get github.com/thomas-marquis/it-happened
```

## Features

**Basic features:**
- **Asynchronous Event Bus**: Decouple your components with a robust pub-sub system.
- **Strongly Linked Events**: Use the `Ref` system to effortlessly track related events across the bus.
- **Powerful Matchers**: Subscribe to events using precise criteria like Type, ID, or Followup relationship.

**What makes a difference:**
- **Event Carriers**: Orchestrate complex workflows by grouping events into `All` or `Sequence` carriers.
- **Automated Lifecycle**: Carriers handle timeouts, concurrency, and completion tracking for you.
- **Result Aggregation**: Use **Outcome Factories** to transform multiple event results into a single, meaningful completion event.

## Getting Started

To get started, define a payload that implements the `event.Payload` interface and initialize an in-memory bus:

```go
type MyPayload struct {
    Message string
}

func (p MyPayload) Type() event.Type {
    return "example.happened"
}

// Initialize the bus
done := make(chan struct{})
bus := inmemory.NewBus(done, nil)
```

## Basic Usage

### Publishing and Subscribing

```go
// Subscribe to your event
bus.Subscribe().
    On(event.Is("example.happened"), func(e event.Event) {
        payload := e.Payload.(MyPayload)
        fmt.Printf("Something happened: %s\n", payload.Message)
    }).
    ListenWithWorkers(1)

// Publish the event
bus.Publish(event.New(MyPayload{Message: "Hello, World!"}))
```

### Using Followup Events

```go
// In a handler, create a followup event
bus.Subscribe().
    On(event.Is("request"), func(e event.Event) {
        bus.Publish(event.NewFollowup(e, ResponsePayload{Data: "OK"}))
    }).
    ListenWithWorkers(1)
```

## Examples

You can find more detailed examples and advanced usage (like Carriers) in the `examples` folder and the [Concepts](concepts.md) page.

## Useful Links

- [GitHub Repository](https://github.com/thomas-marquis/it-happened)