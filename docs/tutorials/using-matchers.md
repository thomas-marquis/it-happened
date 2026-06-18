# Using Matchers

Learn how to use matchers to filter and route events to the appropriate handlers. Matchers provide powerful filtering capabilities that allow subscribers to receive only the events they're interested in.

## What You'll Learn

- The different types of matchers available
- How to match events by type
- How to match multiple event types with a single matcher
- How to match all events
- How matchers work with subscribers

## Prerequisites

- Go 1.25+ installed
- Completed the [Quick Start](../index.md) guide
- Understand basic publish/subscribe (see [Basic Pub/Sub](basic-pubsub.md))
- Familiar with the [Matcher](concepts.md#matcher) concept

## The Matcher Interface

All matchers implement the `event.Matcher` interface:

```go
type Matcher interface {
    Match(event Event) bool
}
```

The `Match` method receives an event and returns `true` if the event matches the matcher's criteria.

## Built-in Matchers

The library provides several built-in matchers for common use cases.

### Is: Match a Specific Event Type

The `Is` matcher matches events of a specific type:

```go
sub.On(event.Is("user.created"), func(e event.Event) {
    // This callback is only invoked for events of type "user.created"
    if payload, ok := e.Payload().(UserCreatedPayload); ok {
        fmt.Printf("User created: %s\n", payload.Username)
    }
})
```

This is the most commonly used matcher for simple event filtering.

### IsOneOf: Match Multiple Event Types

The `IsOneOf` matcher matches any of the specified event types:

```go
sub.On(event.IsOneOf("user.created", "user.updated", "user.deleted"), func(e event.Event) {
    // This callback is invoked for any of the three user event types
    switch payload := e.Payload().(type) {
    case UserCreatedPayload:
        fmt.Println("User was created")
    case UserUpdatedPayload:
        fmt.Println("User was updated")
    case UserDeletedPayload:
        fmt.Println("User was deleted")
    }
})
```

This is useful when you want to handle multiple related event types with the same logic.

### IsAny: Match All Events

The `IsAny` matcher matches all events:

```go
sub.On(event.IsAny(), func(e event.Event) {
    // This callback is invoked for every event published to the bus
    fmt.Printf("Event of type %s received\n", e.Type())
})
```

Use this sparingly, as it will receive all events. It's most useful for logging, monitoring, or debugging.

### IsFollowupOf: Match Followup Events

The `IsFollowupOf` matcher matches events that are followups of specific parent events:

```go
sub.On(event.IsFollowupOf(parentEvent), func(e event.Event) {
    // This callback is invoked for events that are followups of parentEvent
    fmt.Println("Received a followup event")
})
```

This is useful for tracking specific chains of events.

## Complete Example

See the complete, runnable example:

📁 [examples/using-matchers/main.go](../../../examples/using-matchers/main.go)

To run it:

```bash
cd examples/using-matchers
go run main.go
```

Expected output:

```
[IsAny] Event received: user.created
[Is] User created: alice (alice@example.com)
[IsAny] Event received: user.updated
[IsOneOf] User updated: alice.smith (alice.smith@example.com)
[IsAny] Event received: user.deleted
[IsOneOf] User deleted: user-002
[IsAny] Event received: system.health_check
Matcher demonstration complete!
```

Notice how:
- The `Is("user.created")` matcher only receives the user creation event
- The `IsOneOf("user.updated", "user.deleted")` matcher receives both update and delete events
- The `IsAny()` matcher receives all events

## Custom Matchers

You can create custom matchers by implementing the `Matcher` interface:

```go
type HighPriorityMatcher struct{}

func (m HighPriorityMatcher) Match(e event.Event) bool {
    // Implement your custom matching logic
    if payload, ok := e.Payload().(MyPayload); ok {
        return payload.Priority == "high"
    }
    return false
}

// Use the custom matcher
sub.On(HighPriorityMatcher{}, func(e event.Event) {
    fmt.Println("High priority event received")
})
```

Custom matchers are useful when you need matching logic that's not covered by the built-in matchers.

## Matcher Composition

While you can't directly compose matchers with logical operators, you can achieve similar results by:

1. Using `IsOneOf` for OR logic (match any of multiple types)
2. Creating custom matchers for AND logic (all conditions must be true)
3. Using multiple `On()` calls for different matchers

## Best Practices

1. **Be specific**: Use the most specific matcher possible to avoid unnecessary event processing
2. **Group related types**: Use `IsOneOf` for event types that share common handling logic
3. **Avoid IsAny**: Only use `IsAny` when truly necessary (logging, debugging, monitoring)
4. **Type assertions**: Always check the payload type before accessing it in your callback
5. **Error handling**: Handle type assertion failures gracefully

## Key Concepts Used

- **Matcher**: A filter for events (see [Concepts](../concepts.md#matcher))
- **Subscriber**: Receives and processes events (see [Concepts](../concepts.md#subscriber))
- **Event**: The fundamental unit of information (see [Concepts](../concepts.md#event))
- **Type**: Categorizes events (see [Concepts](../concepts.md#type))

## Next Steps

- [Basic Publish/Subscribe](basic-pubsub.md) - Review the fundamentals
- [Event Chaining](event-chaining.md) - Learn how to create sequences of related events
- [Using Carriers](using-carriers.md) - Dispatch multiple events as a unit