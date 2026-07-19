# it-happened

[![Go Reference](https://pkg.go.dev/badge/github.com/thomas-marquis/it-happened.svg)](https://pkg.go.dev/github.com/thomas-marquis/it-happened)
[![CI](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml/badge.svg)](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml)
[![Coverage](https://img.shields.io/badge/Coverage-66%25-yellow)](https://github.com/thomas-marquis/it-happened/actions/workflows/coverage-badge.yml)
[![License](https://img.shields.io/github/license/thomas-marquis/it-happened)](LICENSE)

Event management library written in Go simplifying event driven application development

<p align="center">
  <img src="docs/assets/images/logo-tr.png" width="200" alt="it-happened logo">
</p>

## ✨ Features

- **Asynchronous Event Bus**: Decouple your components with a robust pub-sub system
- **Event Chaining**: Track related events across workflows using ChainRef and ChainPosition
- **Powerful Matchers**: Subscribe to events using precise criteria (by type, followup relationship, etc.)
- **Event Carriers**: Orchestrate complex workflows with All (parallel), Sequence (sequential), and Pipeline (function-based transformation) carriers
- **Automated Lifecycle**: Carriers handle timeouts, concurrency, and completion tracking

## 📦 Installation

### Requirements

- Go 1.25 or higher

### Installation process

```bash
go get github.com/thomas-marquis/it-happened
```

## 📚 Documentation

- [Project Documentation](https://thomas-marquis.github.io/it-happened/) - Complete documentation with:
  - [Quick Start](https://thomas-marquis.github.io/it-happened/) - Get started in 10 minutes
  - [Concepts](https://thomas-marquis.github.io/it-happened/concepts/) - Core library abstractions
  - [Tutorials](https://thomas-marquis.github.io/it-happened/tutorials/) - Practical examples
  - [API References](https://thomas-marquis.github.io/it-happened/references/) - Package documentation links
- [Go Package Documentation](https://pkg.go.dev/github.com/thomas-marquis/it-happened)

## 💻 Usage

See the [Quick Start](https://thomas-marquis.github.io/it-happened/) guide to get started, or explore the [examples](examples/) directory for runnable code samples.

### Basic Usage Examples

#### Publish and Subscribe

```go
import (
    "context"
    "fmt"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/inmemory"
)

// Define your event payload
type UserCreatedPayload struct {
    Username string
}

func (p UserCreatedPayload) EventType() event.Type {
    return "user.created"
}

// Create an event bus
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
bus := inmemory.NewBus(ctx)

// Create a subscriber
sub := bus.Subscribe()

// Register a callback for specific event types
sub.On(event.Is("user.created"), func(e event.Event) {
    if payload, ok := e.Payload().(UserCreatedPayload); ok {
        fmt.Printf("User created: %s\n", payload.Username)
    }
})

// Start listening with workers
sub.ListenWithWorkers(1)
defer sub.Detach()

// Publish an event
evt := event.New(UserCreatedPayload{Username: "john_doe"})
bus.Publish(evt)
```

#### Using Notifiers for Monitoring

```go
import (
    "log"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/inmemory"
)

// Implement a custom notifier
type LoggingNotifier struct {
    event.NopNotifier // Use NopNotifier as base
}

func (n *LoggingNotifier) NotifyPublished(evt event.Event) {
    log.Printf("Event published: %s\n", evt.Type())
}

// Create bus with custom notifier (assuming ctx is available)
bus := inmemory.NewBus(ctx, inmemory.WithNotifier(&LoggingNotifier{}))

// Now all published events will be logged
type MyPayload struct{}
func (MyPayload) EventType() event.Type { return "my.event" }
bus.Publish(event.New(MyPayload{}))
```

#### Event Chaining

```go
import "fmt"

// Define payload types
type OrderPayload struct { OrderID string }
func (p OrderPayload) EventType() event.Type { return "order.created" }

type OrderCompletedPayload struct { OrderID string }
func (p OrderCompletedPayload) EventType() event.Type { return "order.completed" }

// Create an initial event
initialEvent := event.New(OrderPayload{OrderID: "123"})

// Create a followup event in the same chain
followupEvent := initialEvent.NewFollowup(OrderCompletedPayload{OrderID: "123"})

// Both events share the same ChainRef
fmt.Printf("ChainRef: %s\n", initialEvent.ChainRef())
fmt.Printf("ChainRef: %s\n", followupEvent.ChainRef())  // Same as above
fmt.Printf("ChainPosition: %d\n", initialEvent.ChainPosition())    // 0
fmt.Printf("ChainPosition: %d\n", followupEvent.ChainPosition())  // 1
```

#### Using Carriers

```go
import "github.com/thomas-marquis/it-happened/carrier"

// Define payload types
type SendEmailPayload struct { To string }
func (p SendEmailPayload) EventType() event.Type { return "email.sent" }

type UpdateDatabasePayload struct { ID string }
func (p UpdateDatabasePayload) EventType() event.Type { return "db.updated" }

// Create events to dispatch in parallel
events := []event.Event{
    event.New(SendEmailPayload{To: "user@example.com"}),
    event.New(UpdateDatabasePayload{ID: "123"}),
}

// Create an All carrier that dispatches events concurrently
carrierEvent := carrier.NewAll(events, nil, nil)

// Publish the carrier
bus.Publish(carrierEvent)
```

## 🤝 Contribute

All contributions are welcome! Feel free to open an issue or submit a PR. ✨

Check out [CONTRIBUTE.md](CONTRIBUTE.md) for more details.
