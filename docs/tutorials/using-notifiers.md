# Using Notifiers

Learn how to implement and use custom notifiers with the it-happened library to monitor event bus activity without subscribing to events directly.

## What You'll Learn

- How the Notifier interface works
- How to implement a custom notifier
- How to use NopNotifier as a base for partial implementations
- How to configure the in-memory bus with a custom notifier
- Practical use cases for notifiers

## Prerequisites

- Go 1.25+ installed
- Basic understanding of Go interfaces
- Completed the [Quick Start](../index.md) guide
- Familiarity with the [Basic Publish/Subscribe](basic-pubsub.md) tutorial

## Understanding Notifiers

The `event.Notifier` interface provides a way to receive notifications about event bus activity without being a regular subscriber. This is useful for:

- **Monitoring**: Track all published events for logging or metrics
- **Debugging**: Inspect event flow during development
- **Audit trails**: Record when subscribers join or leave the bus
- **Metrics collection**: Count events, measure throughput

## The Notifier Interface

The `event.Notifier` interface defines three methods:

```go
type Notifier interface {
    // Called when an event is published to the bus
    NotifyPublished(Event)
    
    // Called when a new subscriber is added to the bus
    NotifySubscribed(subscriber *Subscriber)
    
    // Called when a subscriber is removed
    NotifyUnsubscribed(subscriber *Subscriber)
}
```

## Using NopNotifier as a Base

The library provides a `NopNotifier` implementation that does nothing. This is useful as a base when you only want to implement some of the notifier methods.

```go
type NopNotifier struct{}

func (n NopNotifier) NotifyPublished(Event) {}
func (n NopNotifier) NotifySubscribed(*Subscriber) {}
func (n NopNotifier) NotifyUnsubscribed(*Subscriber) {}
```

## Step 1: Implement a Custom Notifier

Let's create a simple logging notifier that tracks all published events:

```go
type LoggingNotifier struct {
    event.NopNotifier // Embed to get no-op implementations of unused methods
    PublishedEvents []event.Event
    mu              sync.Mutex
}

// Override only the method we care about
func (n *LoggingNotifier) NotifyPublished(evt event.Event) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.PublishedEvents = append(n.PublishedEvents, evt)
    log.Printf("Event published: %s (ID: %s)", evt.Type(), evt.ID())
}
```

By embedding `event.NopNotifier`, we automatically get working implementations for `NotifySubscribed` and `NotifyUnsubscribed` that do nothing, so we only need to implement the methods we actually want to use.

## Step 2: Create a Complete Notifier

For more comprehensive monitoring, implement all three methods:

```go
type MonitoringNotifier struct {
    PublishedEvents   []event.Event
    SubscribedSubs    []*event.Subscriber
    UnsubscribedSubs  []*event.Subscriber
    mu               sync.Mutex
}

func (n *MonitoringNotifier) NotifyPublished(evt event.Event) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.PublishedEvents = append(n.PublishedEvents, evt)
}

func (n *MonitoringNotifier) NotifySubscribed(sub *event.Subscriber) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.SubscribedSubs = append(n.SubscribedSubs, sub)
}

func (n *MonitoringNotifier) NotifyUnsubscribed(sub *event.Subscriber) {
    n.mu.Lock()
    defer n.mu.Unlock()
    n.UnsubscribedSubs = append(n.UnsubscribedSubs, sub)
}
```

## Step 3: Configure the Bus with a Notifier

Use the `inmemory.WithNotifier()` option when creating your bus:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

// Create your custom notifier
notifier := &MonitoringNotifier{}

// Pass it to the bus constructor
bus := inmemory.NewBus(ctx, inmemory.WithNotifier(notifier))
```

## Step 4: Use the Notifier

Now your notifier will be called automatically:

```go
// Publish events as usual
bus.Publish(event.New(MyPayload{Data: "test"}))

// The notifier's NotifyPublished method will be called
// You can access the tracked data from your notifier instance
notifier.mu.Lock()
for _, evt := range notifier.PublishedEvents {
    fmt.Printf("Tracked event: %s\n", evt.Type())
}
notifier.mu.Unlock()
```

## Practical Use Cases

### 1. Metrics Collection

```go
type MetricsNotifier struct {
    PublishedCount int64
    SubscribeCount int64
    UnsubscribeCount int64
    mu sync.Mutex
}

func (m *MetricsNotifier) NotifyPublished(event.Event) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.PublishedCount++
    // Export to your metrics system
    metrics.Counter("events.published", 1)
}

func (m *MetricsNotifier) NotifySubscribed(*event.Subscriber) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.SubscribeCount++
    metrics.Counter("subscribers.joined", 1)
}

func (m *MetricsNotifier) NotifyUnsubscribed(*event.Subscriber) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.UnsubscribeCount++
    metrics.Counter("subscribers.left", 1)
}
```

### 2. Event Audit Trail

```go
type AuditNotifier struct {
    Events []struct {
        Timestamp time.Time
        Event    event.Event
    }
    mu sync.Mutex
}

func (a *AuditNotifier) NotifyPublished(evt event.Event) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.Events = append(a.Events, struct {
        Timestamp time.Time
        Event    event.Event
    }{
        Timestamp: time.Now(),
        Event:    evt,
    })
}
```

### 3. Selective Notifications

If you only care about certain event types, you can filter in your notifier:

```go
type SelectiveNotifier struct {
    event.NopNotifier
}

func (n *SelectiveNotifier) NotifyPublished(evt event.Event) {
    // Only log important events
    if evt.Type() == "order.placed" || evt.Type() == "payment.failed" {
        log.Printf("IMPORTANT: %s - %s", evt.Type(), evt.ID())
    }
}
```

## Key Concepts Used

- **Notifier**: Receives notifications about bus activity (see [Concepts](../concepts.md#notifier))
- **Bus**: The central communication hub (see [Concepts](../concepts.md#bus))
- **Subscriber**: Receives and processes events (see [Concepts](../concepts.md#subscriber))

## When to Use Notifiers vs Subscribers

| Use Notifiers When | Use Subscribers When |
|---|---|
| You need to monitor ALL events regardless of type | You need to receive specific event types |
| You need to track subscriber lifecycle (join/leave) | You need to process event payloads |
| You want to avoid subscribing to every event type | You want to use matchers to filter events |
| You're building monitoring/audit functionality | You're implementing business logic |

## Complete Example

See the complete, runnable example:

[examples/using-notifiers/main.go](https://github.com/thomas-marquis/it-happened/blob/main/examples/using-notifiers/main.go)

To run it:

```bash
cd examples/using-notifiers
go run main.go
```

## Next Steps

- [Event Chaining](event-chaining.md) - Learn how to create sequences of related events
- [Using Matchers](using-matchers.md) - Explore different matching strategies
- [Using Carriers](using-carriers.md) - Dispatch multiple events as a unit