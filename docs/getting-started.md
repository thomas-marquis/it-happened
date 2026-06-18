# Getting Started with it-happened

This guide will walk you through the basics of using the it-happened library, from setting up your first event-driven application to testing it with the eventest framework.

## Prerequisites

- Go 1.25 or higher
- Basic understanding of Go programming
- Familiarity with event-driven architecture concepts (helpful but not required)

## Installation

```bash
go get github.com/thomas-marquis/it-happened
```

This will install the core library and all its dependencies.

---

## Part 1: Core Event System

### Step 1: Define Your Event Payload

Every event in it-happened carries a payload that implements the `event.Payload` interface:

```go
package myapp

import "github.com/thomas-marquis/it-happened/event"

// Define a custom payload type
type UserCreated struct {
    UserID   string
    Username string
    Email    string
}

// Implement the Payload interface
func (p UserCreated) Type() event.Type {
    return "user.created"
}
```

The `Type()` method returns an `event.Type` (which is just a `string`) that categorizes your event.

### Step 2: Create an Event Bus

The library provides an in-memory event bus implementation:

```go
package myapp

import (
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/event/inmemory"
)

func main() {
    // Create a done channel for graceful shutdown
    done := make(chan struct{})
    
    // Create the bus
    bus := inmemory.NewBus(done, nil)
    
    // Don't forget to close the done channel when shutting down
    defer close(done)
}
```

### Step 3: Subscribe to Events

Subscribe to events using matchers:

```go
// Subscribe to all user.created events
bus.Subscribe().
    On(event.Is("user.created"), func(e event.Event) {
        // Type assertion to get your payload
        payload := e.Payload.(UserCreated)
        fmt.Printf("User created: %s (%s)\n", payload.Username, payload.Email)
    }).
    ListenWithWorkers(1)
```

The `ListenWithWorkers(n)` method starts `n` goroutines to process events concurrently.

### Step 4: Publish Events

```go
// Create and publish an event
userCreatedEvent := event.New(UserCreated{
    UserID:   "12345",
    Username: "john_doe",
    Email:    "john@example.com",
})

bus.Publish(userCreatedEvent)
```

That's it! You now have a basic event-driven system.

---

## Part 2: Event Relationships

### Followup Events

Followup events are used to represent events that are logically connected to previous events:

```go
// In a handler, create a followup event
bus.Subscribe().
    On(event.Is("user.created"), func(e event.Event) {
        // Create a followup event
        followup := event.NewFollowup(
            e, // The original event
            UserWelcomeEmailSent{
                UserID: e.ID, // Same Ref as the original event
            },
        )
        bus.Publish(followup)
    }).
    ListenWithWorkers(1)
```

Followup events share the same `Ref` as their parent event, allowing you to track related events across your system.

### Using Matchers for Followups

You can match followup events:

```go
// Match followup events of a specific type
bus.Subscribe().
    On(event.IsFollowupOfType("user.created"), func(e event.Event) {
        fmt.Println("Followup event received")
    }).
    ListenWithWorkers(1)
```

---

## Part 3: Testing with eventest

The `eventest` package provides a powerful testing framework for event-driven systems using the Marble language.

### Step 1: Import eventest

```go
import (
    "testing"
    "github.com/thomas-marquis/it-happened/eventest"
)
```

### Step 2: Create Your First Test

```go
func TestUserCreationFlow(t *testing.T) {
    // Create an in-memory bus for testing
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Create a harness that expects a specific event sequence
    // "^abc" means: initEvent (^), then event a, then event b, then event c
    // The initEvent (^) is mandatory and marks the start of the timeline
    harness := eventest.NewHarness(bus, "^abc")
    
    // Run the test - no need to pass initEvent anymore
    harness.RunAndWait(t)
}
```

The harness will automatically validate the expected sequence against actual events published to the bus.

### Step 3: Understanding the Test Result

If your events match the expected sequence, the test passes. If not, you'll get clear error messages about what went wrong.

---

## Part 4: Marble Language Basics

The Marble language is a declarative syntax for describing event sequences. Here are the key elements:

### Simple Events

```
^abc
```
initEvent (^), then three events (a, b, c) each in their own time tick. Note: expectations MUST start with initEvent (^).

### Waits

```
^-b-c
```
initEvent (^), event a, wait one tick, event b, wait one tick, event c.

You can also use underscores for waits:
```
^a___b
```
initEvent (^), event a, wait three ticks (each `_` is treated as one tick), event b.

### Groups

Groups allow multiple events to occur within a single time tick.

**Ordered Group** (events must occur in order):
```
^[ab]c
```
initEvent (^), then events a and b in order within one tick, then c in the next tick.

**Unordered Group** (events can occur in any order):
```
^(ab)c
```
initEvent (^), then events a and b in any order within one tick, then c in the next tick.

### Nested Groups

```
^[(ab)c]d
```
initEvent (^), then ordered group containing an unordered group (a and b in any order) followed by c, all within one tick, then d.

```
^[ a (bc) d ]
```
initEvent (^), then ordered group with nested unordered group: a, then b and c in any order, then d, all in one tick.

### initEvent

```
^abc
```
initEvent (^), then a, b, c. The initEvent marks the beginning of the timeline and is mandatory for expectations.

### Followup Events in Marble

```
a<-b
```
Event a is a followup of event b.

---

## Part 5: Using Options

The Harness API supports several options for customizing test behavior:

### WithPayloads

Map event labels to specific payloads:

```go
type MyPayload struct {
    Data string
}

func (p MyPayload) Type() event.Type {
    return "my.payload"
}

func TestWithCustomPayloads(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Expectation MUST start with initEvent (^)
    harness := eventest.NewHarness(
        bus, 
        "^abc",
        eventest.WithPayloads(map[string]event.Payload{
            "a": MyPayload{Data: "test-a"},
            "b": MyPayload{Data: "test-b"},
            "c": MyPayload{Data: "test-c"},
        }),
    )
    
    // Publish the events - they will be matched against the expectation
    // No need to manually advance the clock - RunAndWait handles it
    bus.Publish(event.New(MyPayload{Data: "test-a"}))
    bus.Publish(event.New(MyPayload{Data: "test-b"}))
    bus.Publish(event.New(MyPayload{Data: "test-c"}))
    
    // Call RunAndWait to validate
    harness.RunAndWait(t)
}
```

### WithMatchers

Use custom matchers for more flexible event matching:

```go
func TestWithCustomMatchers(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Expectation MUST start with initEvent (^)
    harness := eventest.NewHarness(
        bus,
        "^a",
        eventest.WithMatchers(map[string]event.Matcher{
            "a": event.HasPayload(MyPayload{Data: "expected"}),
        }),
    )
    
    // Publish the event
    bus.Publish(event.New(MyPayload{Data: "expected"}))
    
    // Call RunAndWait to validate
    harness.RunAndWait(t)
}
```

### WithSideEffect

Execute a marble sequence as a side effect before the main test:

```go
func TestWithSideEffect(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Expectation MUST start with initEvent (^)
    // Side effect MUST NOT contain initEvent (^)
    // Use "-x" to have the side effect start with a wait (aligning with expectation's initEvent)
    harness := eventest.NewHarness(
        bus,
        "^a",
        eventest.WithSideEffect("-x"),
    )
    
    // The side effect "-x" will be executed automatically
    // It publishes event x at tick 1 (after the wait at tick 0)
    // The expectation expects event a at tick 1
    bus.Publish(event.New(eventest.DefaultPayload("a")))
    
    // Call RunAndWait to validate
    harness.RunAndWait(t)
}
```

### WithTickDuration

Configure the duration of each tick:

```go
import "time"

func TestWithCustomTickDuration(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Expectation MUST start with initEvent (^)
    harness := eventest.NewHarness(
        bus,
        "^-ab",
        eventest.WithTickDuration(100*time.Millisecond),
    )
    
    // Publish events - they will be matched against the expectation
    bus.Publish(event.New(eventest.DefaultPayload("a")))
    // Wait for tick duration
    time.Sleep(100 * time.Millisecond)
    bus.Publish(event.New(eventest.DefaultPayload("b")))
    
    // Call RunAndWait to validate
    harness.RunAndWait(t)
}
```

---

## Part 6: Complete Example

Here's a complete example that ties everything together:

```go
package myapp_test

import (
    "testing"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/event/inmemory"
    "github.com/thomas-marquis/it-happened/eventest"
)

// Define your payload types
type UserCreated struct {
    UserID string
}

func (p UserCreated) Type() event.Type {
    return "user.created"
}

type WelcomeEmailSent struct {
    UserID string
}

func (p WelcomeEmailSent) Type() event.Type {
    return "welcome.email.sent"
}

func TestUserCreationWithWelcomeEmail(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Set up the system under test
    // When a user is created, send a welcome email
    bus.Subscribe().
        On(event.Is("user.created"), func(e event.Event) {
            payload := e.Payload.(UserCreated)
            welcomeEvent := event.New(WelcomeEmailSent{
                UserID: payload.UserID,
            })
            bus.Publish(welcomeEvent)
        }).
        ListenWithWorkers(1)
    
    // Create a harness that expects: initEvent, then user.created, then welcome.email.sent
    // The expectation MUST start with initEvent (^)
    harness := eventest.NewHarness(
        bus,
        "^ab",
        eventest.WithPayloads(map[string]event.Payload{
            "a": UserCreated{UserID: "123"},
            "b": WelcomeEmailSent{UserID: "123"},
        }),
    )
    
    // Publish user created event - the welcome email will be sent automatically by the subscriber
    bus.Publish(event.New(UserCreated{UserID: "123"}))
    
    // Call RunAndWait to validate
    harness.RunAndWait(t)
}
```

---

## Next Steps

Now that you've learned the basics, check out:

- [Advanced Usage Guide](advanced.md) - For more complex scenarios
- [Marble Language Specification](marble.md) - Complete reference for Marble syntax
- [Architecture Overview](architecture.md) - Understand the library's design

## Troubleshooting

### Common Issues

**Problem**: Test fails with "no corresponding tick"

**Solution**: Make sure your test publishes events in the correct order and at the right times. Remember that each character in a marble string represents a separate tick unless grouped. Also ensure expectations start with initEvent (^).

**Problem**: Type assertion fails in subscriber

**Solution**: Ensure your payload implements the `Payload` interface and that you're using the correct type in your type assertion.

**Problem**: Events aren't being received

**Solution**: Check that:
1. You've called `ListenWithWorkers(n)` on your subscription
2. The event bus is the same instance for both publisher and subscriber
3. The event type matches what you're subscribing to

---

## Summary

You've now learned:

1. How to set up and use the core event system
2. How to define payloads and publish/subscribe to events
3. How to use followup events for tracking related events
4. How to test your event-driven code with eventest
5. The basics of the Marble language
6. How to use Harness options for customization

The it-happened library provides a powerful, flexible foundation for building event-driven applications in Go, with excellent support for testing through the eventest framework.
