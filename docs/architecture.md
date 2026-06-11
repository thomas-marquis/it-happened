# Architecture Overview

This document describes the high-level architecture of the it-happened library, focusing on both the core event system and the eventest testing framework.

## Table of Contents

- [Core Event System Architecture](#core-event-system-architecture)
- [Eventest Testing Framework Architecture](#eventest-testing-framework-architecture)
- [Component Relationships](#component-relationships)
- [Data Flow](#data-flow)
- [Design Decisions](#design-decisions)

---

## Core Event System Architecture

The core event system is organized into the following packages:

```
┌─────────────────────────────────────────────────────────────────┐
│                         it-happened                                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   event     │    │   event/    │    │  event/carrier       │  │
│  │             │    │   inmemory  │    │                      │  │
│  │  - Event   │    │             │    │  - all.go            │  │
│  │  - Payload │    │  - bus.go   │    │  - sequence.go        │  │
│  │  - Type    │    │             │    │  - carrier.go         │  │
│  │  - Ref     │    │             │    │                      │  │
│  │  - Matcher │    │             │    │                      │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│                                                                     │
└─────────────────────────────────────────────────────────────────┘
```

### event Package

The `event` package contains the fundamental types and interfaces:

- **Event**: The core event type with ID, Type, Payload, Context, and Ref fields
- **Payload**: Interface that all payloads must implement (Type() method)
- **Type**: String type for categorizing events
- **Ref**: Reference identifier for linking related events
- **Matcher**: Interface for filtering events (Match() method)
- **Built-in Matchers**: `Is()`, `HasPayload()`, `IsFollowupOf()`, etc.

### event/inmemory Package

The `inmemory` package provides an in-memory implementation of the event bus:

- **Bus**: Thread-safe event bus implementation
- **Subscriber**: Subscription management with filtering
- Uses channels for asynchronous event delivery
- Supports multiple workers for parallel event processing

### event/carrier Package

The `carrier` package provides event carriers for orchestrating complex workflows:

- **All**: Executes multiple events in parallel
- **Sequence**: Executes events sequentially
- **Outcome Factory**: Creates summary events from carrier results

---

## Eventest Testing Framework Architecture

The eventest package provides a testing framework for event-driven systems:

```
┌─────────────────────────────────────────────────────────────────┐
│                      eventest                                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                    internal/                                    │  │
│  │                                                                  │  │
│  │  ┌─────────────┐    ┌─────────────┐    ┌─────────────────┐  │  │
│  │  │   marble/   │    │  engine/     │    │  gomockevent/    │  │  │
│  │  │             │    │             │    │                 │  │  │
│  │  │  - node.go │    │  - clock/   │    │  - matchers.go    │  │  │
│  │  │  - op.go   │    │  │           │    │                 │  │  │
│  │  │  - parser. │    │  │  - clock. │    │                 │  │  │
│  │  │  - visitor.│    │  │  - go     │    │                 │  │  │
│  │  │  - node_  │    │  │           │    │                 │  │  │
│  │  │    helpers│    │  ├───────────┤    │                 │  │  │
│  │  │  - ...    │    │  │           │    │                 │  │  │
│  │  │           │    │  │  - runtime/│    │                 │  │  │
│  │  │           │    │  │  │         │    │                 │  │  │
│  │  │           │    │  │  ├─────┐  │    │                 │  │  │
│  │  │           │    │  │  │     │  │    │                 │  │  │
│  │  │           │    │  │  │ runtime│  │    │                 │  │  │
│  │  │           │    │  │  │     .go│    │                 │  │  │
│  │  │           │    │  │  │ option │  │    │                 │  │  │
│  │  │           │    │  │  │    .go │    │                 │  │  │
│  │  │           │    │  │  └─────┘  │    │                 │  │  │
│  │  │           │    │  │           │    │                 │  │  │
│  │  │           │    │  ├───────────┤    │                 │  │  │
│  │  │           │    │  │           │    │                 │  │  │
│  │  │           │    │  │ timeline/ │    │                 │  │  │
│  │  │           │    │  │           │    │                 │  │  │
│  │  │           │    │  ├───────────┤    │                 │  │  │
│  │  │           │    │  │interceptor/│    │                 │  │  │
│  │  │           │    │  │           │    │                 │  │  │
│  │  └─────────────┘    └─────────────┘    └─────────────────┘  │
│  │                                                                   │
│  └─────────────────────────────────────────────────────────────┘
│                                                                     │
│  ┌─────────────┐    ┌─────────────┐                              │
│  │  harness.go │    │ harness_test │                              │
│  │             │    │        .go   │                              │
│  └─────────────┘    └─────────────┘                              │
│                                                                     │
└─────────────────────────────────────────────────────────────────┘
```

### marble Package

The `marble` package implements the Marble language for describing event sequences:

- **Node Types**: Hierarchical AST nodes (EventNode, WaitNode, StartNode, FollowupNode, SequenceNode, GroupNode)
- **Visitor Pattern**: Interface for traversing the AST (Visitor, BaseVisitor)
- **Parser**: Converts marble strings to AST nodes (ParseAsNode, Parse)
- **Conversions**: Converts between AST nodes and Op lists (ToOpList, SequenceNodeFromOps)
- **Validation**: Semantic validation rules (Rule interface, WaitlessGroupsRule, StartEventAnywhereRule, etc.)
- **Helpers**: String representation, debugging utilities

#### AST Node Hierarchy

```
Node (interface)
├── EventNode (leaf)
│   └── Name: string
│
├── WaitNode (leaf)
│
├── StartNode (leaf)
│
├── FollowupNode (leaf)
│   ├── NewEvent: string
│   └── OfEvent: string
│
├── SequenceNode (composite)
│   └── Children: []Node
│
└── GroupNode (composite)
    ├── Children: []Node
    └── Ordered: bool
```

#### Visitor Pattern

The Visitor pattern allows for flexible traversal and processing of the AST:

```go
type Visitor interface {
    VisitEvent(*EventNode)
    VisitWait(*WaitNode)
    VisitStart(*StartNode)
    VisitFollowup(*FollowupNode)
    VisitSequence(*SequenceNode)
    VisitGroup(*GroupNode)
}
```

This pattern is used by:
- **TimelineBuilder**: Converts AST to timeline ticks
- **InterceptorValidator**: Validates actual events against expected AST

### engine Packages

The `engine` packages provide the runtime infrastructure for eventest:

#### clock Package

- **Clock**: Interface for controlling time in tests (Start, Stop, Forward, Elapsed, Started)
- **Implementation**: Thread-safe clock with timing control

#### runtime Package

- **Runtime**: Manages execution of marble sequences
- **RunningSession**: Represents an active execution session (Next, HasNext, CurrentTick, Clock)
- **Options**: Configuration options (WithClock, WithPayloadsMapping, WithEventsMapping, WithBaseTickDuration)

#### timeline Package

- **Timeline**: Represents a sequence of ticks
- **Tick**: Represents a single time unit with duration and operations
- **TimelineBuilder**: Builds timelines from AST nodes using Visitor pattern
- **Options**: Timeline configuration (WithTickDuration)

#### interceptor Package

- **Interceptor**: Wraps an event bus to intercept and record published events
- **InterceptorRecorder**: Records expected event sequences and validates them
- **InterceptorValidator**: Validates actual events against expected AST using Visitor pattern
- **activityEntry**: Internal type for recording event activity with timestamps

### harness Package (Public API)

- **Harness**: Main testing harness with fluent API
- **Option**: Functional options for configuring harness behavior
  - `WithPayloads`: Map event labels to payloads
  - `WithEvents`: Map event labels to specific event instances
  - `WithMatchers`: Custom matchers for event validation
  - `WithSideEffect`: Execute marble sequences as side effects before main test
  - `WithTickDuration`: Configure tick duration

---

## Component Relationships

```
┌─────────────────────────────────────────────────────────────────────┐
│                                                                         │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────────────┐    │
│  │  Test Code  │────▶│   Harness   │────▶│    Interceptor      │    │
│  └─────────────┘     └─────────────┘     │   (wraps Bus)       │    │
│                                          │                      │    │
│  ┌─────────────┐     ┌─────────────┐     │  ┌─────────────────┐ │    │
│  │  Marble     │────▶│   Parser    │────▶│  │  Interceptor     │ │    │
│  │  String     │     │             │     │  │  Recorder        │ │    │
│  └─────────────┘     └─────────────┘     │  └─────────────────┘ │    │
│                                          │                      │    │
│                                          │  ┌─────────────────┐ │    │
│                                          │  │   Timeline       │ │    │
│                                          │  │                 │ │    │
│                                          │  └─────────────────┘ │    │
│                                          │                      │    │
│                                          │  ┌─────────────────┐ │    │
│                                          │  │   Validator       │ │    │
│                                          │  │  (Visitor)        │ │    │
│                                          │  └─────────────────┘ │    │
│                                          └──────────────────────┘    │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────┐    │
│  │                        Actual Event Bus                       │    │
│  │                     (e.g., inmemory.Bus)                        │    │
│  └─────────────────────────────────────────────────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────┘
```

### Data Flow

1. **Test Setup**:
   - Test code creates a Harness with a marble string
   - Harness creates an Interceptor that wraps the actual event bus
   - Marble string is parsed into AST nodes
   - AST is validated using semantic rules
   - AST is converted to a Timeline

2. **Test Execution**:
   - Test code executes, publishing events to the bus
   - Interceptor records all published events with timestamps
   - Test code can use side effects to set up initial state

3. **Test Verification**:
   - Harness finishes, triggering validation
   - InterceptorValidator traverses the AST using Visitor pattern
   - Actual events are matched against expected events
   - Errors are collected and reported

### Key Interfaces

#### event Package

```go
type Event interface {
    ID() string
    Type() event.Type
    Payload() event.Payload
    Context() context.Context
    Ref() event.Ref
}

type Payload interface {
    Type() event.Type
}

type Matcher interface {
    Match(event.Event) bool
}

type Bus interface {
    Publish(event.Event)
    Subscribe() *Subscriber
}
```

#### marble Package

```go
type Node interface {
    Accept(Visitor)
    Position() Position
}

type Visitor interface {
    VisitEvent(*EventNode)
    VisitWait(*WaitNode)
    VisitStart(*StartNode)
    VisitFollowup(*FollowupNode)
    VisitSequence(*SequenceNode)
    VisitGroup(*GroupNode)
}

type Rule interface {
    Validate(node Node) error
}
```

#### eventest Package

```go
type Harness struct {
    // Configurable via Options
    bus          event.Bus
    expected     string
    sideEffect   string
    payloadMap   map[string]event.Payload
    eventMap     map[string]event.Event
    matchers     map[string]event.Matcher
    tickDuration time.Duration
}

type Option func(*Harness)
```

---

## Design Decisions

### Why Hierarchical AST?

The original implementation used a flat list of operations with position markers to track group boundaries. This approach had several drawbacks:

1. **Complex Position Tracking**: Required maintaining start/end positions for nested groups
2. **Error-Prone**: Easy to get position calculations wrong, especially with nested groups
3. **Hard to Extend**: Adding new features required complex position manipulation
4. **Difficult to Validate**: Semantic validation was scattered and hard to maintain

The hierarchical AST with Visitor pattern solves these issues:

1. **Natural Representation**: Groups are naturally represented as tree nodes with children
2. **Simpler Traversal**: Visitor pattern provides clean, type-safe traversal
3. **Easier Extension**: New node types can be added without modifying existing code
4. **Better Validation**: Validation rules can be implemented as independent visitors

### Why Visitor Pattern?

The Visitor pattern was chosen for AST traversal because:

1. **Separation of Concerns**: Processing logic (building timelines, validation) is separated from the AST structure
2. **Open/Closed Principle**: New operations can be added without modifying node types
3. **Type Safety**: Each Visit method receives the exact node type, no type assertions needed
4. **Flexibility**: Multiple visitors can process the same AST in different ways

### Why Two Representations (Node and Op)?

The system maintains both hierarchical Node representation and flat Op list representation for:

1. **Backward Compatibility**: Existing code using Op lists continues to work
2. **Gradual Migration**: Components can be migrated to use Nodes incrementally
3. **Different Use Cases**: Some operations are easier with flat lists (e.g., serialization), others with trees (e.g., validation)
4. **Testing**: Easy to verify equivalence between representations

The conversion functions (`ToOpList`, `SequenceNodeFromOps`) ensure both representations stay in sync.

### Error Handling Strategy

The library uses a consistent error handling approach:

1. **Parse Errors**: Returned from parser for syntax errors
2. **Semantic Errors**: Returned from validation rules for logical errors
3. **Runtime Errors**: Returned from runtime for execution errors
4. **Panic for Programming Errors**: Used for internal inconsistencies (e.g., nil node)

This approach provides clear error messages while maintaining robustness.

---

## Performance Considerations

### Memory Usage

- AST nodes are allocated during parsing and kept for the duration of test execution
- Timeline ticks are pre-computed for efficiency
- Interceptor records all events, so long-running tests may use significant memory

### Time Complexity

- **Parsing**: O(n) where n is the length of the marble string
- **Validation**: O(n) where n is the number of nodes (each node visited once)
- **Timeline Building**: O(n) where n is the number of nodes
- **Event Matching**: O(n*m) where n is expected events and m is actual events

### Optimization Opportunities

1. **Node Pooling**: Reuse node allocations for common patterns
2. **Lazy Validation**: Validate only when needed
3. **Incremental Matching**: Match events as they occur rather than at the end
4. **Caching**: Cache parsed marbles and built timelines

---

## Extensibility

The architecture is designed to be extensible:

1. **New Node Types**: Add new node types by implementing the Node interface and adding Visit methods to the Visitor interface
2. **New Validation Rules**: Implement the Rule interface and add to validation chain
3. **New Matchers**: Implement the Matcher interface for custom event matching
4. **New Clock Implementations**: Implement the Clock interface for different timing strategies
5. **New Bus Implementations**: Implement the Bus interface for different transport mechanisms
