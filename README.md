# it-happened

[![Go Reference](https://pkg.go.dev/badge/github.com/thomas-marquis/it-happened.svg)](https://pkg.go.dev/github.com/thomas-marquis/it-happened)
[![CI](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml/badge.svg)](https://github.com/thomas-marquis/it-happened/actions/workflows/ci.yaml)
[![License](https://img.shields.io/github/license/thomas-marquis/it-happened)](LICENSE)

Event management library written in Go simplifying event driven application development

<p align="center">
  <img src="docs/assets/images/logo-tr.png" width="200" alt="it-happened logo">
</p>

## ✨ Features

### Core Event System
- **Asynchronous Event Bus**: Decouple your components with a robust pub-sub system
- **Strongly Linked Events**: Use the `Ref` system to effortlessly track related events across the bus
- **Powerful Matchers**: Subscribe to events using precise criteria like Type, ID, or Followup relationship
- **Event Carriers**: Orchestrate complex workflows by grouping events into `All` or `Sequence` carriers
- **Automated Lifecycle**: Carriers handle timeouts, concurrency, and completion tracking for you

### Testing Framework (eventest)
- **Marble Language**: Declarative syntax for describing event sequences and timelines
- **Harness API**: Simple testing interface for verifying event-driven behavior
- **Interceptor Pattern**: Verify actual events match expected marble sequences
- **Flexible Matching**: Use custom matchers, payloads, and event mappings

## 📦 Installation

### Requirements

- Go 1.25 or higher

### Installation process

```bash
go get github.com/thomas-marquis/it-happened
```

## 📚 Documentation

- [Project Documentation](https://thomas-marquis.github.io/it-happened/)
- [Go Package Documentation](https://pkg.go.dev/github.com/thomas-marquis/it-happened)
- [Marble Language Specification](docs/marble.md)

## 💻 Usage

### Basic Event Bus Usage

```go
import (
    "fmt"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/event/inmemory"
)

type MyPayload struct {
    Message string
}

func (p MyPayload) Type() event.Type {
    return "example.happened"
}

func main() {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)

    // Subscribe to events
    bus.Subscribe().
        On(event.Is("example.happened"), func(e event.Event) {
            payload := e.Payload.(MyPayload)
            fmt.Printf("Something happened: %s\n", payload.Message)
        }).
        ListenWithWorkers(1)

    // Publish an event
    bus.Publish(event.New(MyPayload{Message: "Hello, World!"}))
}
```

### Testing with eventest

```go
import (
    "testing"
    "time"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/event/inmemory"
    "github.com/thomas-marquis/it-happened/eventest"
)

func TestEventSequence(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)

    // Create a harness that expects events "a", "b", "c" in order
    harness := eventest.NewHarness(bus, "abc")

    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        // Your test code that publishes events
        bus.Publish(event.New(eventest.DefaultPayload("a")))
        bus.Publish(event.New(eventest.DefaultPayload("b")))
        bus.Publish(event.New(eventest.DefaultPayload("c")))
    })
}

func TestWithGroups(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)

    // Use marble syntax: [ab] means ordered group, (ab) means unordered group
    // This expects events a and b in order, then c
    harness := eventest.NewHarness(bus, "[ab]c")

    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(eventest.DefaultPayload("a")))
        bus.Publish(event.New(eventest.DefaultPayload("b")))
        bus.Publish(event.New(eventest.DefaultPayload("c")))
    })
}

func TestWithWaits(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)

    // Use - for wait ticks: a-b means a, wait, then b
    harness := eventest.NewHarness(bus, "a-b-c")

    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(eventest.DefaultPayload("a")))
        // Wait for next tick (default is 10ms)
        clock.Forward(10 * time.Millisecond)
        bus.Publish(event.New(eventest.DefaultPayload("b")))
        clock.Forward(10 * time.Millisecond)
        bus.Publish(event.New(eventest.DefaultPayload("c")))
    })
}
```

## 🤝 Contribute

All contributions are welcome! Feel free to open an issue or submit a PR. ✨

Check out [CONTRIBUTE.md](CONTRIBUTE.md) for more details.
