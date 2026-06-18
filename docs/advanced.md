# Advanced Usage Guide

This guide covers advanced topics and patterns for using the it-happened library and eventest framework effectively.

## Table of Contents

- [Advanced Marble Syntax](#advanced-marble-syntax)
- [Custom Matchers](#custom-matchers)
- [Complex Event Flows](#complex-event-flows)
- [Testing Asynchronous Systems](#testing-asynchronous-systems)
- [Multiple Harnesses](#multiple-harnesses)
- [Error Handling and Validation](#error-handling-and-validation)
- [Performance Considerations](#performance-considerations)
- [Integration with Existing Code](#integration-with-existing-code)
- [Best Practices](#best-practices)

---

## Advanced Marble Syntax

### Combining All Features

Marble supports combining all its features in complex sequences:

```
^a-(bc)[d(e<-f)g]h-i
```

This translates to:
1. initEvent (`^`) - marks the start of the timeline
2. Event `a`
3. Unordered group: `b` and `c` in any order (single tick)
4. Ordered group: `d`, then unordered group (`e` is followup of `f`), then `g` - all in order (single tick)
5. Event `h`
6. Wait tick (`-`)
7. Event `i`

### Named Events

Use `/` prefix for multi-character event names:

```
^/login /auth.success /dashboard.load
```

This is especially useful when event names would otherwise be ambiguous or when you need descriptive names.

### Complex Nesting

Deeply nested groups are supported:

```
^[ a ( b [ c d ] e ) f ]
```

This creates:
- Outer ordered group containing:
  - Event `a`
  - Unordered group containing:
    - Event `b`
    - Ordered group containing events `c` and `d`
    - Event `e`
  - Event `f`
- All within a single time tick

### Followup Events in Groups

Followup events can appear anywhere, including within groups:

```
^[ a<-b c<-d ]
```

In an ordered group, the followup relationships must still be respected within the ordering constraints. Note: expectations MUST start with initEvent (^).

---

## Custom Matchers

### Creating Custom Matchers

Implement the `event.Matcher` interface to create custom matching logic:

```go
type CustomMatcher struct {
    expectedValue string
}

func (m CustomMatcher) Match(e event.Event) bool {
    // Your custom matching logic
    if payload, ok := e.Payload.(MyPayload); ok {
        return payload.Value == m.expectedValue
    }
    return false
}

func (m CustomMatcher) String() string {
    return fmt.Sprintf("has value %q", m.expectedValue)
}

// Use it in a test
harness := eventest.NewHarness(
    bus,
    "a",
    eventest.WithMatchers(map[string]event.Matcher{
        "a": CustomMatcher{expectedValue: "expected"},
    }),
)
```

### Using Mock Matchers

The library provides gomock-based matchers for testing:

```go
import "github.com/thomas-marquis/it-happened/eventest/gomockevent"

// Match by payload equality
matcher := gomockevent.PayloadEq(MyPayload{Value: "test"})

// Match followup events
fromEvt := event.New(MyPayload{Value: "original"})
matcher := gomockevent.IsFollowupOf(fromEvt)
```

### Combining Matchers

Create composite matchers by combining multiple conditions:

```go
type AndMatcher struct {
    matchers []event.Matcher
}

func (m AndMatcher) Match(e event.Event) bool {
    for _, matcher := range m.matchers {
        if !matcher.Match(e) {
            return false
        }
    }
    return true
}

func (m AndMatcher) String() string {
    var parts []string
    for _, matcher := range m.matchers {
        parts = append(parts, matcher.String())
    }
    return "(" + strings.Join(parts, " and ") + ")"
}

// Use it
matcher := AndMatcher{
    matchers: []event.Matcher{
        event.Is("user.created"),
        event.HasPayload(UserCreated{Username: "admin"}),
    },
}
```

---

## Complex Event Flows

### Sequential Processing

Test systems with sequential event processing:

```go
func TestSequentialProcessing(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // System under test: process events in sequence
    bus.Subscribe().
        On(event.Is("step1"), func(e event.Event) {
            // Process step 1, then emit step 2
            bus.Publish(event.New(Step2Payload{}))
        }).
        On(event.Is("step2"), func(e event.Event) {
            // Process step 2, then emit step 3
            bus.Publish(event.New(Step3Payload{}))
        }).
        ListenWithWorkers(1)
    
    // Expect the complete sequence
    harness := eventest.NewHarness(bus, "abc")
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(Step1Payload{}))
    })
}
```

### Parallel Processing with Groups

Test parallel event processing:

```go
func TestParallelProcessing(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // System under test: emit multiple events in parallel
    bus.Subscribe().
        On(event.Is("start"), func(e event.Event) {
            // Emit events a, b, c in parallel
            go func() { bus.Publish(event.New(EventAPayload{})) }()
            go func() { bus.Publish(event.New(EventBPayload{})) }()
            go func() { bus.Publish(event.New(EventCPayload{})) }()
        }).
        ListenWithWorkers(3)
    
    // Use unordered group since events can arrive in any order
    harness := eventest.NewHarness(bus, "(abc)")
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(StartPayload{}))
    })
}
```

### Conditional Event Flows

Test conditional logic in your event handlers:

```go
func TestConditionalFlow(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // System under test: different paths based on payload
    bus.Subscribe().
        On(event.Is("request"), func(e event.Event) {
            payload := e.Payload.(RequestPayload)
            if payload.IsValid {
                bus.Publish(event.New(SuccessPayload{}))
            } else {
                bus.Publish(event.New(FailurePayload{}))
            }
        }).
        ListenWithWorkers(1)
    
    // Test success path
    harness := eventest.NewHarness(
        bus,
        "ab",
        eventest.WithPayloads(map[string]event.Payload{
            "a": RequestPayload{IsValid: true},
            "b": SuccessPayload{},
        }),
    )
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(RequestPayload{IsValid: true}))
    })
    
    // Test failure path
    harness2 := eventest.NewHarness(
        bus,
        "ac",
        eventest.WithPayloads(map[string]event.Payload{
            "a": RequestPayload{IsValid: false},
            "c": FailurePayload{},
        }),
    )
    
    harness2.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(RequestPayload{IsValid: false}))
    })
}
```

---

## Testing Asynchronous Systems

### Using Clock Control

The test clock allows precise control over time in your tests:

```go
func TestAsyncWithClock(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // System under test: delay event processing
    bus.Subscribe().
        On(event.Is("request"), func(e event.Event) {
            // Simulate async processing
            go func() {
                time.Sleep(50 * time.Millisecond)
                bus.Publish(event.New(ResponsePayload{}))
            }()
        }).
        ListenWithWorkers(1)
    
    // Expect request, wait 5 ticks, response
    // With default 10ms tick duration, 5 ticks = 50ms
    harness := eventest.NewHarness(
        bus,
        "a-----b",
        eventest.WithTickDuration(10*time.Millisecond),
    )
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(RequestPayload{}))
        // The clock automatically advances as events are published
    })
}
```

### Manual Clock Advancement

For more control, manually advance the clock:

```go
func TestManualClockAdvancement(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    harness := eventest.NewHarness(
        bus,
        "a-b-c",
        eventest.WithTickDuration(100*time.Millisecond),
    )
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(EventAPayload{}))
        
        // Manually advance to next tick
        clock.Forward(100 * time.Millisecond)
        
        bus.Publish(event.New(EventBPayload{}))
        
        // Advance again
        clock.Forward(100 * time.Millisecond)
        
        bus.Publish(event.New(EventCPayload{}))
    })
}
```

---

## Multiple Harnesses

### Testing Multiple Subscribers

Use multiple harnesses to verify different parts of your system:

```go
func TestMultipleSubscribers(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // First subscriber: transforms a to b
    bus.Subscribe().
        On(event.Is("a"), func(e event.Event) {
            bus.Publish(event.New(EventBPayload{}))
        }).
        ListenWithWorkers(1)
    
    // Second subscriber: transforms a to c
    bus.Subscribe().
        On(event.Is("a"), func(e event.Event) {
            bus.Publish(event.New(EventCPayload{}))
        }).
        ListenWithWorkers(1)
    
    // First harness: verify first subscriber
    harness1 := eventest.NewHarness(bus, "ab")
    
    // Second harness: verify second subscriber
    harness2 := eventest.NewHarness(bus, "ac")
    
    // Run both tests (they share the same bus)
    harness1.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(EventAPayload{}))
    })
    
    harness2.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(EventAPayload{}))
    })
}
```

### Isolating Tests

Each test should use its own bus to ensure isolation:

```go
func TestIsolated1(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    harness := eventest.NewHarness(bus, "a")
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(EventAPayload{}))
    })
}

func TestIsolated2(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    harness := eventest.NewHarness(bus, "b")
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(EventBPayload{}))
    })
}
```

---

## Error Handling and Validation

### Custom Validation Rules

Create custom validation rules for your marble sequences:

```go
// Implement the marble.Rule interface
type MaxEventsRule struct {
    maxEvents int
}

func (r MaxEventsRule) Validate(node marble.Node) error {
    counter := &eventCounter{}
    node.Accept(counter)
    
    if counter.count > r.maxEvents {
        return fmt.Errorf("sequence has %d events, maximum allowed is %d", 
            counter.count, r.maxEvents)
    }
    return nil
}

type eventCounter struct {
    marble.BaseVisitor
    count int
}

func (v *eventCounter) VisitEvent(*marble.EventNode) {
    v.count++
}

func (v *eventCounter) VisitFollowup(*marble.FollowupNode) {
    v.count++
}

// Use it in your code
node, _ := marble.ParseAsNode("abc")
if err := marble.Validate(node, MaxEventsRule{maxEvents: 2}); err != nil {
    // Handle validation error
}
```

### Testing Error Cases

Use Go's testing.T to verify error handling:

```go
func TestErrorCase(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    harness := eventest.NewHarness(bus, "abc")
    
    // This test is expected to fail
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        // Only publish a and b, missing c
        bus.Publish(event.New(eventest.DefaultPayload("a")))
        bus.Publish(event.New(eventest.DefaultPayload("b")))
    })
    
    // The test will fail with an error message about missing event c
}
```

---

## Performance Considerations

### Optimizing Large Test Suites

1. **Reuse Bus Instances**: For multiple tests in the same package, consider reusing bus instances
2. **Minimize Tick Duration**: Use shorter tick durations for faster tests
3. **Avoid Deep Nesting**: Deeply nested groups can be harder to debug and validate

### Memory Management

Long-running tests with many events can consume significant memory:

```go
// Process events in batches to reduce memory usage
func TestLargeSequence(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Generate 1000 events
    marbleStr := strings.Repeat("a", 1000)
    
    harness := eventest.NewHarness(bus, marbleStr)
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        for i := 0; i < 1000; i++ {
            bus.Publish(event.New(eventest.DefaultPayload("a")))
        }
    })
}
```

### Parallel Testing

Use Go's built-in parallel testing:

```go
func TestParallel1(t *testing.T) {
    t.Parallel()
    // Test code
}

func TestParallel2(t *testing.T) {
    t.Parallel()
    // Test code
}
```

---

## Integration with Existing Code

### Wrapping Existing Systems

Wrap existing event systems with the it-happened bus interface:

```go
type MyExistingBus struct {
    // Your existing bus implementation
}

func (b *MyExistingBus) Publish(e event.Event) {
    // Convert to your existing event format and publish
    existingEvent := convertToExisting(e)
    b.existingPublish(existingEvent)
}

func (b *MyExistingBus) Subscribe() *event.Subscriber {
    // Return a subscriber that adapts to your existing system
    return &adapterSubscriber{
        onEvent: func(e event.Event) {
            // Convert from existing format
            converted := convertFromExisting(e)
            // Call handler
        },
    }
}
```

### Using with HTTP Servers

Test HTTP handlers that publish events:

```go
func TestHTTPHandler(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Create handler that publishes events
    handler := &MyHandler{Bus: bus}
    
    // Create test server
    mux := http.NewServeMux()
    mux.HandleFunc("/users", handler.CreateUser)
    server := httptest.NewServer(mux)
    defer server.Close()
    
    // Set up harness
    harness := eventest.NewHarness(bus, "a")
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        // Make HTTP request
        resp, err := http.Post(
            server.URL+"/users",
            "application/json",
            strings.NewReader(`{"name": "test"}`),
        )
        if err != nil {
            t.Fatal(err)
        }
        defer resp.Body.Close()
        
        // The handler should have published event "a"
    })
}
```

---

## Best Practices

### Test Organization

1. **Group Related Tests**: Keep tests for the same functionality together
2. **Descriptive Names**: Use descriptive test names that explain the behavior
3. **Test Both Happy and Error Paths**: Test success cases and error handling
4. **Keep Tests Focused**: Each test should verify one specific behavior

### Marble String Design

1. **Use Groups for Simultaneous Events**: Use `[ab]` or `(ab)` when events occur in the same tick
2. **Use Waits for Timing**: Use `-` to represent explicit waits between events
3. **Start with Simple**: Start with simple marble strings and add complexity as needed
4. **Add Comments**: Use Go comments to explain complex marble strings

### Example: Well-Structured Test

```go
// TestUserRegistration tests the complete user registration flow
// Marble: ^ /user.registered (/email.sent /sms.sent) /user.activated
// This means:
//   1. initEvent
//   2. User registered
//   3. Unordered group: email and SMS can arrive in any order
//   4. User activated
func TestUserRegistration(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Set up system under test
    setupUserRegistrationHandler(bus)
    
    // Create harness
    harness := eventest.NewHarness(
        bus,
        "^/user.registered(/email.sent/sms.sent)/user.activated",
    )
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        // Trigger registration
        bus.Publish(event.New(UserRegisteredPayload{}))
    })
}
```

### Performance Tips

1. **Use t.Parallel()**: Run independent tests in parallel
2. **Minimize Clock Advancement**: Only advance the clock when necessary
3. **Reuse Bus**: For integration tests, reuse the same bus across multiple harnesses
4. **Clean Up**: Always close done channels to prevent goroutine leaks

### Debugging Tips

1. **Check Event Types**: Verify your payloads implement the Payload interface correctly
2. **Verify Matchers**: Ensure your matchers are matching events as expected
3. **Inspect Marble Strings**: Use the `marble.String()` function to convert nodes back to marble
4. **Check Clock Timing**: Verify clock advancement matches your expectations

---

## Advanced Patterns

### Event Aggregation

Test systems that aggregate multiple events:

```go
func TestEventAggregation(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // System under test: aggregate events
    var received []string
    bus.Subscribe().
        On(event.Is("event"), func(e event.Event) {
            payload := e.Payload.(NamedPayload)
            received = append(received, payload.Name)
            
            if len(received) == 3 {
                // Aggregate and publish result
                bus.Publish(event.New(AggregatedPayload{
                    Events: received,
                }))
            }
        }).
        ListenWithWorkers(1)
    
    // Expect 3 events, then aggregation
    harness := eventest.NewHarness(bus, "aaaab")
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(NamedPayload{Name: "first"}))
        bus.Publish(event.New(NamedPayload{Name: "second"}))
        bus.Publish(event.New(NamedPayload{Name: "third"}))
    })
}
```

### Conditional Validation

Use custom validation based on test context:

```go
func TestConditionalValidation(t *testing.T) {
    done := make(chan struct{})
    bus := inmemory.NewBus(done, nil)
    
    // Create harness with conditional validation
    harness := eventest.NewHarness(
        bus,
        "a",
        eventest.WithMatchers(map[string]event.Matcher{
            "a": &conditionalMatcher{
                condition: func(e event.Event) bool {
                    // Your condition logic
                    return e.Payload.(ConditionalPayload).ShouldMatch
                },
            },
        }),
    )
    
    harness.Run(t, func(bus event.Bus, clock eventest.Clock) {
        bus.Publish(event.New(ConditionalPayload{ShouldMatch: true}))
    })
}

type conditionalMatcher struct {
    condition func(event.Event) bool
}

func (m *conditionalMatcher) Match(e event.Event) bool {
    return m.condition(e)
}

func (m *conditionalMatcher) String() string {
    return "conditional matcher"
}
```

---

## Summary

This guide covered:

1. **Advanced Marble Syntax**: Complex combinations and nesting
2. **Custom Matchers**: Creating flexible matching logic
3. **Complex Event Flows**: Sequential, parallel, and conditional processing
4. **Testing Async Systems**: Clock control and timing
5. **Multiple Harnesses**: Testing multiple subscribers and isolation
6. **Error Handling**: Custom validation and error cases
7. **Performance**: Optimization tips for large test suites
8. **Integration**: Working with existing systems
9. **Best Practices**: Organizing and debugging tests
10. **Advanced Patterns**: Aggregation and conditional validation

For more information, check out:

- [Architecture Overview](architecture.md) - Understand the library's design
- [Marble Language Specification](marble.md) - Complete reference
- [Getting Started Guide](getting-started.md) - Basics and first steps
