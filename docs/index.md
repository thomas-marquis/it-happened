# Quick Start

<figure markdown="span">
  ![Logo](assets/images/logo-tr.png)
</figure>

Welcome to **it-happened**, a Go library that simplifies event-driven application development. This guide will get you up and running with basic event publishing and subscription in just a few minutes.

## Prerequisites

- Go 1.25 or higher
- Basic familiarity with Go syntax

## Installation

```bash
go get github.com/thomas-marquis/it-happened
```

## Your First Event

Let's create a simple application that publishes and receives events.

### Step 1: Define Your Event Payload

First, create a payload type that implements the `event.Payload` interface:

```go
package main

import (
    "fmt"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/inmemory"
)

type MyEventPayload struct {
    Message string
}

func (p MyEventPayload) EventType() event.Type {
    return "my.event"
}
```

### Step 2: Create the Event Bus

Initialize an in-memory event bus:

```go
func main() {
    // Create a done channel to control bus lifetime
    done := make(chan struct{})
    defer close(done)

    // Create the bus with the done channel
    bus := inmemory.NewBus(done, nil)
```

### Step 3: Subscribe to Events

Create a subscriber that listens for your event type:

```go
    // Create a subscriber
    sub := bus.Subscribe()

    // Register a callback for your event type
    sub.On(event.Is("my.event"), func(e event.Event) {
        // Extract the payload
        if payload, ok := e.Payload().(MyEventPayload); ok {
            fmt.Printf("Received event: %s\n", payload.Message)
        }
    })

    // Start listening with 1 worker
    sub.ListenWithWorkers(1)
```

### Step 4: Publish an Event

Publish your event to the bus:

```go
    // Create and publish an event
    evt := event.New(MyEventPayload{Message: "Hello, World!"})
    bus.Publish(evt)

    // Wait a moment for the event to be processed
    // In a real application, you would use proper synchronization
}
```

### Complete Example

Here's the complete code in one place:

```go
package main

import (
    "fmt"
    "time"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/inmemory"
)

type MyEventPayload struct {
    Message string
}

func (p MyEventPayload) EventType() event.Type {
    return "my.event"
}

func main() {
    done := make(chan struct{})
    defer close(done)

    bus := inmemory.NewBus(done, nil)

    sub := bus.Subscribe()
    sub.On(event.Is("my.event"), func(e event.Event) {
        if payload, ok := e.Payload().(MyEventPayload); ok {
            fmt.Printf("Received: %s\n", payload.Message)
        }
    })
    sub.ListenWithWorkers(1)

    evt := event.New(MyEventPayload{Message: "Hello, World!"})
    bus.Publish(evt)

    // Wait for event processing
    time.Sleep(100 * time.Millisecond)
}
```

### Run It

Save the code to a file (e.g., `main.go`) and run:

```bash
go run main.go
```

You should see output similar to:

```
Received: Hello, World!
```

## Next Steps

Now that you have the basics working, explore these next:

1. **[Concepts](concepts.md)** - Understand the core library abstractions
2. **[Tutorials](tutorials/)** - Practical examples for common use cases
3. **[References](references.md)** - API documentation

## Need Help?

- Check the [Concepts](concepts.md) page for detailed explanations
- Browse the [Tutorials](tutorials/) for practical examples
- Visit the [References](references.md) for API documentation
- Open an issue on [GitHub](https://github.com/thomas-marquis/it-happened)

---

*This Quick Start guide should take less than 10 minutes to complete.*