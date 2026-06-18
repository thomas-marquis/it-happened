# Technical Specification: Marble Testing Feature Refactoring

## Overview

This document provides a detailed technical specification for refactoring the marble testing feature to address the requirements outlined in DRAFT.md. The goal is to create a cleaner, more consistent API where tests are driven by expectations and the `^` initEvent has a well-defined, unambiguous role.

## Current State Analysis

### Existing Architecture

The marble testing system consists of:

1. **Harness** (`eventest/harness.go`): Main entry point with `PublishAndWait` method
2. **Marble Parser** (`eventest/internal/marble/parser.go`): Parses marble strings into AST nodes
3. **Node Types** (`eventest/internal/marble/node.go`): AST nodes including `InitEventNode` (represents `^`)
4. **Semantic Rules** (`eventest/internal/marble/semantic.go`): Validation rules including initEvent rules
5. **Runtime** (`eventest/internal/engine/runtime/`): Executes side effect marble sequences
6. **Interceptor** (`eventest/internal/engine/interceptor/`): Intercepts and validates published events against expected sequences
7. **Timeline** (`eventest/internal/engine/timeline/`): Converts marble AST to executable ticks

### Current Terminology Confusion

The `^` character is currently referred to by multiple names:
- `PlaceholderNode` in the code (to be renamed to `InitEventNode`)
- initEvent in documentation (`marble.md` line 13)
- initEvent in error messages (`semantic.go` line 71)
- initEvent in DRAFT.md

### Current Flow

```
User calls: PublishAndWait(t, initEvents...)
    ↓
Creates Interceptor with expected marble sequence
    ↓
If sideEffect is set:
    Runtime.RunAll(sideEffect) → publishes side effect events
    Clock stops when side effect completes
Else:
    Clock starts
    User publishes initEvent event
    Clock stops when test completes
    ↓
Interceptor.Finish() → validates all recorded events against expected sequence
```

## Problem Statement

### Issue 1: initEvent Naming and Role Confusion
- The initEvent symbol (`^`) is currently called `PlaceholderNode`, `StartEvent`, or `initEvent` interchangeably
- Its role is ambiguous: it represents the initial event but is treated inconsistently
- It can appear in both expectation and side effect chains, causing confusion

### Issue 2: Test Duration Not Driven by Expectation
- Currently, test duration is determined by whichever completes first: side effect or user's test code
- The side effect chain can be longer than the expectation chain
- No explicit error when side effect exceeds expectation duration

### Issue 3: initEvent in Side Effect Chain
- The initEvent (`^`) can currently appear in side effect marble sequences
- This is semantically incorrect as it represents the user's initial event, not a side effect

## Proposed Solution

### Core Principles

1. **The initEvent**: Standardize terminology to "initEvent" throughout
2. **initEvent is Unique**: Exactly one initEvent (`^`) must exist in the expectation chain
3. **initEvent is Mandatory**: Every expectation chain must begin with initEvent (`^`), it mey be part of a group, but must be in the first tick
4. **initEvent is NOT in Side Effects**: Side effect chains MUST NEVER contain initEvent (`^`)
5. **Test Duration = Expectation Duration**: The test runs for exactly the duration of the expectation chain
6. **Side Effect Synchronization**: Side effect chain starts at the VERY SAME tick as expectation (tick 0) and must complete within or at the same time as expectation duration (side effect may terminate sooner, but never later)

### New Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Harness                                  │
├─────────────────────────────────────────────────────────────┤
│  expected:     "^abc"    (MUST start with initEvent, defines   │
│               duration)                                       │
│  sideEffect:   "-abc"    (NO initEvent, starts at SAME tick 0  │
│               as expectation, may terminate sooner)            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Interceptor                               │
│  - Validates actual events against expected sequence            │
│  - Expected sequence duration determines test end time         │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────────┼───────────────────┐
              ▼                   ▼                   ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   Runtime        │ │    Timeline       │ │    Validator      │
│  - Executes side │ │  - Converts marble │ │  - Validates     │
│    effect chain  │ │    to ticks       │ │    actual vs      │
│  - Starts at     │ │  - Calculates     │ │    expected       │
│    tick 0        │ │    durations      │ │  - Checks start   │
│  - Must complete │ │                   │ │    event presence │
│    within expect │ │                   │ │  - Checks side    │
│    duration      │ │                   │ │    effect ≤ expect│
└──────────────────┘ └──────────────────┘ └──────────────────┘
```

### Terminology Standardization

| Old Terms | New Standard Term |
|-----------|-------------------|
| PlaceholderNode | InitEventNode |
| placeholder | initEvent |
| start event | initEvent (MANDATORY) |

### API Changes

#### Harness Creation

**Before:**
```go
harness := eventest.NewHarness(bus, "^abc",
    eventest.WithSideEffect("-abc"))
harness.PublishAndWait(t, initEvent)
```

**After:**
```go
harness := eventest.NewHarness(bus, "^abc",
    eventest.WithSideEffect("abc"))
harness.RunAndWait(t)
```

Key changes:
- Expectation string MUST start with initEvent (`^`) (enforced at parse/validation time)
- Side effect string MUST NOT contain initEvent (`^`) (enforced at validation time)
- Side effect CAN start with a wait (e.g., "-abc") - both expectation and side effect still start at tick 0
- Method renamed from `PublishAndWait` to `RunAndWait`
- No initEvents passed by user (initEvent is implicit)

#### Renamed Types

```go
// Before
type PlaceholderNode struct{}

// After
type InitEventNode struct{}
```

#### Validation Rules

**New Semantic Rules:**

1. **MandatoryInitEventRule**: Expectation chain MUST start with exactly one initEvent (`^`)
   ```go
   type MandatoryInitEventRule struct{}
   func (r MandatoryInitEventRule) Validate(node Node) error {
       // Must have exactly one initEvent at position 0
   }
   ```

2. **NoInitEventInSideEffectRule**: Side effect chain MUST NEVER contain initEvent (`^`)
   ```go
   type NoInitEventInSideEffectRule struct{}
   func (r NoInitEventInSideEffectRule) Validate(node Node) error {
       // Return error if any InitEventNode exists in side effect
   }
   ```

3. **SideEffectDurationRule**: Side effect duration must not exceed expectation duration
   ```go
   type SideEffectDurationRule struct {
       expectedDuration int // in ticks
   }
   func (r SideEffectDurationRule) Validate(node Node) error {
       // Calculate side effect duration in ticks
       // Compare with expectedDuration
       // Return error if side effect > expected
       // Note: Both expectation and side effect start at tick 0 by definition
   }
   ```

### Execution Flow

```
1. User calls: harness.RunAndWait(t)
   
2. Harness:
   a. Parse expectation marble (MUST start with initEvent)
   b. Validate expectation with MandatoryInitEventRule
   c. Parse side effect marble (if provided)
   d. Validate side effect with NoInitEventInSideEffectRule (MUST NOT contain `^`)
   e. Calculate expectation duration from timeline
   f. Calculate side effect duration from timeline
   g. Validate side effect duration <= expectation duration
   h. Both start at the VERY SAME tick 0 (by marble definition)
   
3. Create Interceptor with expected sequence
   
4. Create Runtime for side effect (if provided)
   
5. Start clock
   
6. If side effect exists:
   a. Runtime executes side effect chain starting at tick 0
   b. Side effect events are published to bus
   
7. Clock advances through ticks
   
8. User's test code (via RunAndWait callback or separate goroutine) can publish additional events
   
9. Interceptor records all events
   
10. When clock reaches expectation duration:
    a. Clock stops
    b. Interceptor.Finish() validates all events
    c. If side effect not complete, error
    
11. Test completes
```

### Timeline Synchronization

**Time Model:**
- Each marble operator (event, wait, group) = 1 tick
- Default tick duration: 10ms (configurable)
- initEvent `^` occupies tick 0
- Side effect starts at the VERY SAME tick 0 as expectation

**Example:**
```
Expectation: "^abc"      (4 ticks: initEvent, a, b, c)
Side Effect: "-abc"     (4 ticks: wait, a, b, c)

Tick 0: initEvent (^) + Side effect wait (-)
Tick 1: a + a
Tick 2: b + b
Tick 3: c + c

Test duration: 4 ticks (40ms)
Both start at tick 0. Side effect has wait in first tick but still
starts at the same tick as expectation.

Another valid example:
Expectation: "(^abc)"   (1 tick: unordered group with initEvent)
Side Effect: "(ab)"     (1 tick: unordered group, same duration)

Tick 0: initEvent, a, b, c (expectation) + a, b (side effect)
```

**Error Case:**
```
Expectation: "^a-b"    (3 ticks)
Side Effect: "a-b-c"   (5 ticks)

Error: "side effect duration (5 ticks) exceeds expectation duration (3 ticks)"
```

### Implementation Changes

#### 1. Rename PlaceholderNode to InitEventNode

**Files affected:**
- `eventest/internal/marble/node.go`
- `eventest/internal/marble/parser.go`
- `eventest/internal/marble/op.go`
- `eventest/internal/marble/node_helpers.go`
- `eventest/internal/marble/semantic.go`
- `eventest/internal/marble/visitor.go`

**Change:**
```go
// Before
type PlaceholderNode struct { pos Position }
func (n *PlaceholderNode) Accept(v Visitor) { v.VisitPlaceholder(n) }

// After
type InitEventNode struct { pos Position }
func (n *InitEventNode) Accept(v Visitor) { v.VisitInitEvent(n) }
```

#### 2. Update Parser

**File:** `eventest/internal/marble/parser.go`

```go
// Before
case c == '^':
    children = append(children, &PlaceholderNode{pos: Position{Offset: *pos}})

// After
case c == '^':
    children = append(children, &InitEventNode{pos: Position{Offset: *pos}})
```

#### 3. Update Op Type

**File:** `eventest/internal/marble/op.go`

```go
// Before
const (
    WaitOpType OpType = iota
    EventOpType
    PlaceholderEventOpType
    EventWithFollowupOpType
    ...
)
type PlaceholderEventOp struct{}
func (o PlaceholderEventOp) Type() OpType { return PlaceholderEventOpType }
func (o PlaceholderEventOp) String() string { return "^" }

// After
const (
    WaitOpType OpType = iota
    EventOpType
    InitEventOpType
    EventWithFollowupOpType
    ...
)
type InitEventOp struct{}
func (o InitEventOp) Type() OpType { return InitEventOpType }
func (o InitEventOp) String() string { return "^" }
```

#### 4. Update Visitor Interface

**File:** `eventest/internal/marble/visitor.go`

```go
// Before
type Visitor interface {
    VisitEvent(*EventNode)
    VisitWait(*WaitNode)
    VisitPlaceholder(*PlaceholderNode)
    VisitFollowup(*FollowupNode)
    VisitSequence(*SequenceNode)
    VisitGroup(*GroupNode)
}

type BaseVisitor struct{}
func (v *BaseVisitor) VisitPlaceholder(*PlaceholderNode) {}

// After
type Visitor interface {
    VisitEvent(*EventNode)
    VisitWait(*WaitNode)
    VisitInitEvent(*InitEventNode)
    VisitFollowup(*FollowupNode)
    VisitSequence(*SequenceNode)
    VisitGroup(*GroupNode)
}

type BaseVisitor struct{}
func (v *BaseVisitor) VisitInitEvent(*InitEventNode) {}
```

#### 5. Update Semantic Rules

**File:** `eventest/internal/marble/semantic.go`

Replace existing start event rules with:

```go
// MandatoryInitEventRule: Expectation must have exactly one initEvent (^) at beginning
type MandatoryInitEventRule struct{}

func (r MandatoryInitEventRule) Validate(node Node) error {
    v := &initEventVisitor{}
    node.Accept(v)
    
    if v.count == 0 {
        return errors.Join(ErrSemantic, errors.New("expectation must have an initEvent (^) at the beginning"))
    }
    if v.count > 1 {
        return errors.Join(ErrSemantic, errors.New("expectation can have only one initEvent (^)"))
    }
    
    // Check it's at the beginning
    seq, ok := node.(*SequenceNode)
    if !ok || len(seq.Children) == 0 {
        return nil
    }
    if !isFirstNodeInitEvent(seq.Children[0]) {
        return errors.Join(ErrSemantic, errors.New("initEvent (^) must be at the beginning of expectation"))
    }
    
    return nil
}

func isFirstNodeInitEvent(n Node) bool {
    switch node := n.(type) {
    case *InitEventNode:
        return true
    case *GroupNode:
        if len(node.Children) > 0 {
            return isFirstNodeInitEvent(node.Children[0])
        }
    case *SequenceNode:
        if len(node.Children) > 0 {
            return isFirstNodeInitEvent(node.Children[0])
        }
    }
    return false
}

type initEventVisitor struct {
    BaseVisitor
    count int
}

func (v *initEventVisitor) VisitInitEvent(*InitEventNode) {
    v.count++
}

func (v *initEventVisitor) VisitSequence(n *SequenceNode) {
    for _, child := range n.Children {
        child.Accept(v)
    }
}

func (v *initEventVisitor) VisitGroup(n *GroupNode) {
    for _, child := range n.Children {
        child.Accept(v)
    }
}

// NoInitEventInSideEffectRule: Side effect must not contain initEvent (^)
type NoInitEventInSideEffectRule struct{}

func (r NoInitEventInSideEffectRule) Validate(node Node) error {
    v := &initEventVisitor{}
    node.Accept(v)
    
    if v.count > 0 {
        return errors.Join(ErrSemantic, errors.New("side effect must not contain initEvent (^)"))
    }
    
    return nil
}
```

Remove old rules:
- `StartEventAtBeginningRule`
- `StartEventAnywhereRule`
- `UniqueStartEventRule`

#### 6. Update Harness

**File:** `eventest/harness.go`

```go
// Before
type Harness struct {
    bus          event.Bus
    expected     string
    sideEffect   string
    ...
}

func (h *Harness) PublishAndWait(t *testing.T, placeholders ...event.Event) {
    // ...
}

// After
type Harness struct {
    bus          event.Bus
    expected     string
    sideEffect   string
    ...
}

func (h *Harness) RunAndWait(t *testing.T) {
    clk := clock.NewClock()
    intercept := interceptor.NewInterceptor(t, h.bus, clk)
    
    // Parse and validate expectation
    expectedNode, err := marble.ParseAsNode(h.expected)
    if err != nil {
        t.Fatalf("failed to parse expected marble: %v", err)
    }
    if err := marble.Validate(expectedNode, marble.MandatoryInitEventRule{}); err != nil {
        t.Fatalf("invalid expectation marble: %v", err)
    }
    
    recorder := intercept.EXPECT().FromMarble(h.expected)
    
    // Apply matchers from payload and event maps (same as before)
    // ...
    
    // Parse and validate side effect
    if h.sideEffect != "" {
        sideEffectNode, err := marble.ParseAsNode(h.sideEffect)
        if err != nil {
            t.Fatalf("failed to parse side effect marble: %v", err)
        }
        if err := marble.Validate(sideEffectNode, 
            marble.NoInitEventInSideEffectRule{},
            marble.WaitlessGroupsRule{}); err != nil {
            t.Fatalf("invalid side effect marble: %v", err)
        }
        
        // Calculate durations
        expectedTimeline := timeline.NewTimeline(expectedNode)
        sideEffectTimeline := timeline.NewTimeline(sideEffectNode)
        
        expectedDuration := calculateTotalDuration(expectedTimeline.Ticks())
        sideEffectDuration := calculateTotalDuration(sideEffectTimeline.Ticks())
        
        if sideEffectDuration > expectedDuration {
            t.Fatalf("side effect duration (%d ticks) exceeds expectation duration (%d ticks)",
                sideEffectDuration, expectedDuration)
        }
        
        // Run side effect
        rt := runtime.NewRuntime(intercept,
            runtime.WithClock(clk),
            runtime.WithPayloadsMapping(h.payloadMap),
            runtime.WithEventsMapping(h.eventMap),
            runtime.WithBaseTickDuration(h.tickDuration))
        
        if err := rt.RunAll(h.sideEffect); err != nil {
            t.Fatalf("failed to execute side effect: %v", err)
        }
        
        // Clock will be stopped by side effect completion, but we need to
        // ensure it runs for the full expectation duration
        // Actually, we should NOT stop the clock here - let it run to expectation end
        
    } else {
        clk.Start()
        defer clk.Stop()
    }
    
    // Wait for full expectation duration
    expectedTimeline := timeline.NewTimeline(expectedNode)
    totalDuration := calculateTotalDuration(expectedTimeline.Ticks())
    
    // If clock is already running (side effect case), advance to total duration
    // If clock is not running (no side effect), start it
    if !clk.Started() {
        clk.Start()
    }
    
    // Block until full duration
    time.Sleep(totalDuration * h.tickDuration)
    clk.Stop()
    
    intercept.Finish()
}

func calculateTotalDuration(ticks []timeline.Tick) int {
    return len(ticks)
}
```

#### 7. Update Runtime

**File:** `eventest/internal/engine/runtime/runtime.go`

The runtime currently uses `StartEventAnywhereRule` for validation. Update to not validate initEvent (since side effects won't have them):

```go
// Before
if err := marble.Validate(node,
    marble.StartEventAnywhereRule{},
    marble.WaitlessGroupsRule{},
); err != nil {
    return nil, err
}

// After
if err := marble.Validate(node,
    marble.WaitlessGroupsRule{},
); err != nil {
    return nil, err
}
```

Also, the runtime should NOT treat `InitEventOp` specially in side effects - it should error if encountered:

```go
// In Runtime.RunAllFromNode or RunFromNode
// Add validation that node contains no InitEventNode
if err := marble.Validate(node, marble.NoInitEventInSideEffectRule{}); err != nil {
    return err
}
```

#### 8. Update Interceptor

**File:** `eventest/internal/engine/interceptor/interceptor.go`

The interceptor's `FromMarble` method already validates with `WaitlessGroupsRule`. We need to also validate with `MandatoryInitEventRule`:

```go
// Before
if err := marble.Validate(node,
    marble.WaitlessGroupsRule{}); err != nil {
    panic(err)
}

// After
if err := marble.Validate(node,
    marble.MandatoryInitEventRule{},
    marble.WaitlessGroupsRule{}); err != nil {
    panic(err)
}
```

#### 9. Update Validator

**File:** `eventest/internal/engine/interceptor/validator.go`

The validator currently skips `PlaceholderNode`. Update to handle `InitEventNode`:

```go
// Before
func (v *InterceptorValidator) VisitPlaceholder(n *marble.PlaceholderNode) {
    // initEvent is usually not verified in the same way, but we can check if it was published
    // For now, we skip it as it's often used for initialization
}

// After
func (v *InterceptorValidator) VisitInitEvent(n *marble.InitEventNode) {
    // initEvent marks tick 0
    // It should be present in the actual events
    v.validateInitEvent(v.currentTick)
}

func (v *InterceptorValidator) validateInitEvent(tickIdx int) {
    if tickIdx != 0 {
        v.errors = append(v.errors, fmt.Errorf("initEvent must be at tick 0"))
        return
    }
    
    // The initEvent should have been published
    // This is validated by checking that there's activity in tick 0
    // The actual validation happens in the sequence validation
    // For initEvent specifically, we just ensure it's in the right position
}

// Also update VisitSequence to properly handle InitEventNode
func (v *InterceptorValidator) VisitSequence(n *marble.SequenceNode) {
    for i, child := range n.Children {
        v.currentTick = i
        child.Accept(v)
    }
}
```

#### 10. Update Timeline Builder

**File:** `eventest/internal/engine/timeline/timeline_builder.go`

Update to handle `InitEventNode`:

```go
// Before
func (b *TimelineBuilder) VisitPlaceholder(n *marble.PlaceholderNode) {
    op := marble.PlaceholderEventOp{}
    // ...
}

// After
func (b *TimelineBuilder) VisitInitEvent(n *marble.InitEventNode) {
    op := marble.InitEventOp{}
    if b.currentOps == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: b.tickDuration,
            Ops:      []marble.Op{op},
        })
    } else {
        b.currentOps = append(b.currentOps, op)
    }
}
```

#### 11. Update All References

Use `grep` to find all references to:
- `PlaceholderNode` → replace with `InitEventNode`
- `PlaceholderEventOp` → replace with `InitEventOp`
- `PlaceholderEventOpType` → replace with `InitEventOpType`
- `VisitPlaceholder` → replace with `VisitInitEvent`
- `placeholder` in error messages → replace with `initEvent`

### Testing Strategy

1. **Unit Tests for New Rules**: Add tests for `MandatoryInitEventRule` and `NoInitEventInSideEffectRule`
2. **Integration Tests**: Test the complete flow with new harness API
3. **Backward Compatibility**: Provide migration guide for existing tests
4. **Error Cases**: Test all error conditions (missing initEvent, initEvent in side effect, duration mismatch)

### Migration Guide

**For existing tests:**

```go
// Old style
func TestSomething(t *testing.T) {
    bus := inmemory.NewBus(nil, nil)
    harness := eventest.NewHarness(bus, "^abc",
        eventest.WithSideEffect("-abc"))
    harness.PublishAndWait(t, initEvent)
}

// New style
func TestSomething(t *testing.T) {
    bus := inmemory.NewBus(nil, nil)
    harness := eventest.NewHarness(bus, "^abc",
        eventest.WithSideEffect("abc"))  // Remove leading - or ^ from side effect
    harness.RunAndWait(t)
}
```

**Key migration steps:**
1. Ensure all expectation strings start with initEvent (`^`)
2. Remove initEvent (`^`) from all side effect strings
3. Remove `-` prefix from side effect strings (side effect starts at tick 0)
4. Replace `PublishAndWait(t, events...)` with `RunAndWait(t)`
5. Remove initEvent creation and passing

### Benefits

1. **Clear Semantics**: initEvent (`^`) has a single, well-defined meaning
2. **Consistent API**: Expectation always starts with initEvent, side effect never contains it
3. **Explicit Duration**: Test duration is clearly defined by expectation chain
4. **Better Error Messages**: Clear validation errors for all edge cases
5. **Simplified Code**: Removes multiple initEvent rule variants

### Potential Challenges

1. **Breaking Change**: This is a breaking change for existing users
2. **Test Migration**: All existing tests need to be updated
3. **Backward Compatibility**: Consider if backward compatibility layer is needed (not recommended - clean break is better)

## Recommended Implementation Order

1. Rename `PlaceholderNode` to `InitEventNode` throughout codebase
2. Rename `PlaceholderEventOp` to `InitEventOp` throughout codebase
3. Update visitor interface and all implementations
4. Add new semantic rules (`MandatoryInitEventRule`, `NoInitEventInSideEffectRule`)
5. Update harness to use `RunAndWait` method
6. Update runtime validation
7. Update interceptor validation
8. Update timeline builder
9. Update all tests to use new API
10. Add new tests for error cases

## Files to Modify

**Core marble package:**
- `eventest/internal/marble/node.go`
- `eventest/internal/marble/op.go`
- `eventest/internal/marble/parser.go`
- `eventest/internal/marble/visitor.go`
- `eventest/internal/marble/node_helpers.go`
- `eventest/internal/marble/semantic.go`
- `eventest/internal/marble/parser_test.go`
- `eventest/internal/marble/semantic_test.go`
- `eventest/internal/marble/node_test.go`

**Engine packages:**
- `eventest/internal/engine/timeline/timeline_builder.go`
- `eventest/internal/engine/timeline/timeline.go`
- `eventest/internal/engine/runtime/runtime.go`
- `eventest/internal/engine/interceptor/interceptor.go`
- `eventest/internal/engine/interceptor/validator.go`
- `eventest/internal/engine/interceptor/interceptor_test.go`
- `eventest/internal/engine/runtime/runtime_test.go`

**Public API:**
- `eventest/harness.go`
- `eventest/harness_test.go`

**Documentation:**
- `docs/marble.md` (update terminology and examples)
- `docs/getting-started.md` (update examples)
- `docs/advanced.md` (update examples)

## Validation

The implementation must satisfy these test cases:

```go
// Test 1: Valid expectation with initEvent
NewHarness(bus, "^abc").RunAndWait(t)  // PASS

// Test 2: Missing initEvent in expectation
NewHarness(bus, "abc").RunAndWait(t)   // FAIL: expectation must start with initEvent

// Test 3: initEvent in side effect
NewHarness(bus, "^abc", WithSideEffect("-ab")).RunAndWait(t)  // FAIL: side effect must not contain initEvent

// Test 4: Side effect longer than expectation
NewHarness(bus, "^ab", WithSideEffect("-ab")).RunAndWait(t)  // FAIL: side effect duration exceeds expectation

// Test 5: Valid side effect
NewHarness(bus, "^abc", WithSideEffect("-ab")).RunAndWait(t)  // PASS

// Test 6: Side effect same length as expectation
NewHarness(bus, "^abc", WithSideEffect("-abc")).RunAndWait(t)  // PASS
```

## Conclusion

This specification provides a comprehensive plan for refactoring the marble testing feature to address all requirements in DRAFT.md. The solution standardizes terminology to initEvent, enforces clear semantic rules, and creates a more intuitive API where the test duration is explicitly driven by the expectation chain.
