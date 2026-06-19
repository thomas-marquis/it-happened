# Testing Contract: Test Coverage Improvement

**Feature**: Test Coverage Improvement  
**Date**: 2026-06-19  
**Spec**: [spec.md](../spec.md)  
**Plan**: [plan.md](../plan.md)

## Overview

This document defines the testing contracts for the it-happened library. Since this feature focuses on improving test coverage rather than adding new public APIs, the "contracts" are the testing conventions, patterns, and expectations that all tests must follow.

## Public API Testing Contracts

### 1. Event Package Contract

The `event` package exposes the following interfaces and types that must be tested:

#### Bus Interface Contract
```go
// Contract: event.Bus
type Bus interface {
    // Publish MUST deliver the event to all matching subscribers
    // Testing Requirement: Verify event is published to all subscribers
    // Testing Requirement: Verify event matching works correctly
    Publish(evt Event)
    
    // Subscribe MUST return a new Subscriber instance
    // Testing Requirement: Verify Subscriber is properly initialized
    // Testing Requirement: Verify Subscriber can register handlers
    Subscribe() *Subscriber
}
```

**Test Coverage Requirements**:
- [ ] Publish delivers events to all subscribers
- [ ] Publish handles concurrent calls safely
- [ ] Subscribe returns a valid Subscriber
- [ ] Multiple subscribers receive all matching events

#### Subscriber Contract
```go
// Contract: event.Subscriber
type Subscriber struct {
    // Register MUST add the handler for matching events
    // Testing Requirement: Verify handler is invoked for matching events
    // Testing Requirement: Verify handler is NOT invoked for non-matching events
    Register(matcher Matcher, handler Handler)
    
    // Unregister MUST remove the handler
    // Testing Requirement: Verify handler is no longer invoked after unregister
    Unregister(matcher Matcher)
}
```

**Test Coverage Requirements**:
- [ ] Register adds handler successfully
- [ ] Handler is invoked for matching events
- [ ] Handler is not invoked for non-matching events
- [ ] Unregister removes handler successfully
- [ ] Unregister is idempotent (calling twice doesn't error)

#### Notifier Contract
```go
// Contract: event.Notifier (internal to event package)
// Testing Requirement: Verify all registered callbacks are invoked
// Testing Requirement: Verify callback order is preserved
```

**Test Coverage Requirements**:
- [ ] All registered callbacks are invoked
- [ ] Callbacks are invoked in registration order
- [ ] Empty notifier doesn't panic

#### Option Contract
```go
// Contract: event.Option
// Testing Requirement: Verify all options correctly configure events
// Testing Requirement: Verify options can be composed
```

**Test Coverage Requirements**:
- [ ] Each option type correctly sets its property
- [ ] Multiple options can be applied to the same event
- [ ] Options are applied in order

### 2. Carrier Package Contract

The `carrier` package exposes the following interfaces and types:

#### Carrier Interface Contract
```go
// Contract: carrier.Carrier
type Carrier interface {
    event.Payload
    
    // Dispatch MUST publish all events in the carrier to the bus
    // Testing Requirement: Verify all events are dispatched
    // Testing Requirement: Verify dispatch respects carrier configuration
    Dispatch(bus event.Bus)
}
```

**Test Coverage Requirements**:
- [ ] Dispatch publishes all events to the bus
- [ ] Dispatch respects timeout configuration
- [ ] Dispatch respects concurrency limits
- [ ] Dispatch respects completion condition

#### All Carrier Contract
```go
// Contract: carrier.All
// Behavior: Dispatches all events concurrently
// Testing Requirement: Verify events are dispatched in parallel
// Testing Requirement: Verify all events are dispatched even if some fail
```

**Test Coverage Requirements**:
- [ ] All events are dispatched to the bus
- [ ] Events are dispatched concurrently (not sequentially)
- [ ] Individual event failures don't prevent other events from being dispatched
- [ ] Configuration options (timeout, concurrency) are respected

#### Sequence Carrier Contract
```go
// Contract: carrier.Sequence
// Behavior: Dispatches events sequentially
// Testing Requirement: Verify events are dispatched in order
// Testing Requirement: Verify each event completes before the next starts
```

**Test Coverage Requirements**:
- [ ] All events are dispatched to the bus
- [ ] Events are dispatched in the order they were added
- [ ] Each event completes before the next event is dispatched
- [ ] If an event fails, subsequent events are not dispatched (depending on configuration)
- [ ] Configuration options (timeout, completion condition) are respected

#### CompletionCondition Contract
```go
// Contract: carrier.CompletionCondition
// type CompletionCondition func(sent, received event.Event) bool
// Testing Requirement: Verify default completion condition works
// Testing Requirement: Verify custom completion conditions can be set
```

**Test Coverage Requirements**:
- [ ] Default completion condition returns expected results
- [ ] Custom completion conditions can be provided
- [ ] Completion condition is applied to all dispatched events

### 3. Inmemory Package Contract

The `inmemory` package provides the concrete bus implementation:

#### Inmemory Bus Contract
```go
// Contract: inmemory.Bus (implements event.Bus)
type Bus struct {
    // Implementation of event.Bus interface
}
```

**Test Coverage Requirements**:
- [ ] Publish delivers events to all registered subscribers
- [ ] Subscribe returns a properly configured Subscriber
- [ ] Concurrent Publish and Subscribe operations are thread-safe
- [ ] Event matching works correctly with various matcher types
- [ ] Subscriber registration and unregistration work correctly
- [ ] Multiple subscribers receive all matching events

## Testing Framework Contracts

### 1. Test Structure Contract

All tests MUST follow this structure:

```go
func Test<Component>_<Behavior>(t *testing.T) {
    t.Run("<description of scenario>", func(t *testing.T) {
        // Given: <setup initial state>
        <setup code>
        
        // When: <perform action>
        <action code>
        
        // Then: <verify outcome>
        <assertion code using testify/assert>
    })
}
```

**Contract Requirements**:
- [ ] Test function names follow Go conventions (TestXxx)
- [ ] Use t.Run() for each test scenario
- [ ] Include Given/When/Then comments
- [ ] Use testify/assert for assertions
- [ ] Use testify/require for preconditions

### 2. Mocking Contract

Mocks MUST be generated and used according to this contract:

```go
// In gen.go
//go:generate mockgen -source=<source> -destination=<destination> -package=mocks

// Usage in tests
mockCtrl := gomock.NewController(t)
// defer mockCtrl.Finish() => this is no longer necessary, gomock already manages it

mockBus := mocks.NewMockBus(mockCtrl)
// Setup expectations
mockBus.EXPECT().Publish(gomock.Any()).Times(1)

// Test code that uses the mock
mockBus.Publish(event)
```

**Contract Requirements**:
- [ ] Mocks generated using mockgen (gomock)
- [ ] Mocks stored in internal/mocks/ directory
- [ ] Mocks NOT edited manually
- [ ] Mocks regenerated using `go generate ./...`
- [ ] Mock expectations set before test execution
- [ ] mockCtrl.Finish() called at end of test

### 3. Coverage Contract

Coverage MUST be computed and validated according to this contract:

```bash
# Local coverage computation
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out

# With race detector
go test -race -cover ./... -coverprofile=coverage.out

# Makefile target (to be added)
make coverage
```

**Contract Requirements**:
- [ ] Coverage computed using `go test -cover`
- [ ] Each package must achieve >= 80% coverage
- [ ] Race detector must pass for all tests
- [ ] Coverage badge must be up-to-date in README.md

## Test Quality Contracts

### Readability Contract (Carrier Tests)

Carrier tests MUST be highly readable:

```go
// GOOD: Readable test with clear structure
func TestAllCarrier_Dispatch(t *testing.T) {
    t.Run("dispatches all events in parallel", func(t *testing.T) {
        // Given: a carrier with 3 events and a bus with tracking subscriber
        carrier := carrier.NewAll(event1, event2, event3)
        bus := inmemory.NewBus()
        received := make([]event.Event, 0)
        
        bus.Subscribe().Register(matcher.Any(), func(evt event.Event) {
            received = append(received, evt)
        })
        
        // When: the carrier dispatches events
        carrier.Dispatch(bus)
        
        // Then: all events are published to the bus
        assert.Len(t, received, 3)
        assert.Contains(t, received, event1)
        assert.Contains(t, received, event2)
        assert.Contains(t, received, event3)
    })
    
    t.Run("handles event dispatch failures gracefully", func(t *testing.T) {
        // Given, When, Then...
    })
}

// BAD: Less readable test (avoid)
func TestAllCarrier(t *testing.T) {
    c := carrier.NewAll(e1, e2, e3)
    b := inmemory.NewBus()
    var r []event.Event
    b.Subscribe().Register(matcher.Any(), func(e event.Event) { r = append(r, e) })
    c.Dispatch(b)
    assert.Len(t, r, 3)
}
```

**Readability Requirements**:
- [ ] Use descriptive test and subtest names
- [ ] Use clear variable names
- [ ] Add comments for complex setups
- [ ] Keep test logic focused on one behavior
- [ ] Use helper functions for repeated setups
- [ ] Avoid excessive nesting

### Asynchronous Testing Contract

Tests for asynchronous components MUST follow these patterns:

**Pattern 1: Using sync.WaitGroup**
```go
func TestConcurrentPublish(t *testing.T) {
    var wg sync.WaitGroup
    bus := inmemory.NewBus()
    numSubscribers := 10
    
    for i := 0; i < numSubscribers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            subscriber := bus.Subscribe()
            subscriber.Register(matcher.Any(), func(evt event.Event) {
                // Process event
            })
        }()
    }
    
    wg.Wait()
    // All subscribers registered, now publish
    bus.Publish(event)
    
    // Verify all subscribers received the event
}
```

**Pattern 2: Using Channels**
```go
func TestEventDelivery(t *testing.T) {
    bus := inmemory.NewBus()
    deliveryChan := make(chan event.Event, 1)
    
    bus.Subscribe().Register(matcher.Any(), func(evt event.Event) {
        deliveryChan <- evt
    })
    
    bus.Publish(testEvent)
    
    select {
    case received := <-deliveryChan:
        assert.Equal(t, testEvent, received)
    case <-time.After(100 * time.Millisecond):
        t.Fatal("Event not delivered within timeout")
    }
}
```

**Asynchronous Testing Requirements**:
- [ ] Use synchronization primitives (WaitGroup, channels, mutexes)
- [ ] Avoid timing-based assertions when possible
- [ ] Use reasonable timeouts if timing is necessary
- [ ] Always test with `-race` flag to detect race conditions
- [ ] Verify thread-safety under concurrent operations

## Compliance Matrix

| Requirement | Contract Location | Status |
|-------------|------------------|--------|
| Test-First Development | Constitution Principle II | Mandatory |
| Use testify assertions | Quality Standards | Mandatory |
| Use gomock for mocking | Quality Standards | Mandatory |
| Use t.Run() for subtests | Quality Standards | Mandatory |
| Use Given/When/Then comments | Quality Standards | Mandatory |
| Mocks in internal/mocks/ | Development Workflow | Mandatory |
| Mocks generated via gen.go | Development Workflow | Mandatory |
| Coverage >= 80% | Success Criteria SC-001 | Mandatory |
| README.md has coverage badge | Success Criteria SC-009 | Mandatory |
| CONTRIBUTE.md has Testing Strategy | Success Criteria SC-006 | Mandatory |