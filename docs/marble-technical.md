# Marble Testing Implementation - Technical Analysis, Improvements, and Architecture

## Table of Contents
1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Identified Issues and Problems](#identified-issues-and-problems)
3. [Recommended Improvements](#recommended-improvements)
4. [Enhanced Recommendations](#enhanced-recommendations)
5. [Complete Architecture Scheme](#complete-architecture-scheme)

---

## Current Architecture Analysis

### Package Structure

```
eventest/
├── harness.go           # Main testing harness (INCOMPLETE)
├── matchers.go          # Gomock matchers for event testing
└── internal/
    ├── marble/
    │   ├── parser.go      # Lexer/parser for marble syntax
    │   ├── op.go          # Operation types (AST nodes)
    │   ├── semantic.go    # Validation rules
    │   ├── parser_test.go # Parser tests
    │   └── semantic_test.go # Validation tests
    └── runtime/
        ├── runtime.go     # Runtime execution engine
        ├── timeline.go    # Timeline construction
        ├── clock.go       # Virtual clock for time control
        ├── option.go       # Configuration options
        ├── interceptor.go # Event interception for verification
        └── *test.go files  # Unit tests
```

### Data Flow

```
Marble String
     ↓
Parser (marble.Parse)
     ↓
[]Op (AST)
     ↓
Semantic Validator (marble.Validate)
     ↓
Timeline Builder (runtime.NewTimeline)
     ↓
[]Tick (Execution Plan)
     ↓
Runtime (runtime.Run / RunAll)
     ↓
Event Bus (Publish Events)
```

### Key Components

1. **Parser**: Converts marble strings to operation AST
2. **Semantic Validator**: Enforces language rules
3. **Timeline**: Converts operations to time-based ticks
4. **Runtime**: Executes the timeline, publishing events
5. **Interceptor**: Captures and verifies published events

---

## Identified Issues and Problems

### Critical Issues

1. **Incomplete Harness Implementation** (`eventest/harness.go`)
   - `WithSideEffect`, `WithPayloads`, `WithMatchers` options are empty stubs
   - `Harness` struct is empty
   - `Run` method does nothing
   - This is the main entry point that's not functional

2. **Harness Disconnected from Runtime**
   - No integration between `Harness` and the `runtime` package
   - Runtime is internal but harness needs to use it

3. **Interceptor Incomplete** (`eventest/internal/runtime/interceptor.go`)
   - Lines 114, 124, 129: TODO comments about wrong approach
   - `failuresFromGroup` has incomplete implementation (lines 226-314)
   - `failuresFromOrderedGroup` and `failuresFromUnorderedGroup` are stubs
   - Duplicate code in `Failures()` method (strict vs lenient modes)

4. **Timeline Construction Issues**
   - `handleGroup` function has complex position tracking that's error-prone
   - Group handling for ordered vs unordered groups is mixed in one complex method
   - Position adjustments for nested groups are fragile

5. **Runtime Issues**
   - Start event (`^`) is parsed but not actually used in runtime
   - Clock management could be more robust
   - Error handling in `RunAll` swallows `SessionEnded` error

### Code Quality Issues

1. **Dead Code**
   - `eventest/internal/runtime/interceptor.go` lines 316-325: Commented out code
   - Lines 375-391: Unfinished utility functions
   - Lines 386-390: Commented out struct

2. **Code Duplication**
   - `failuresFromGroup` has both strict and lenient logic mixed
   - Event matching logic duplicated in interceptor

3. **Type Safety Issues**
   - Type assertions without proper error handling (e.g., line 77-83 in semantic.go)
   - Panics instead of proper error returns in several places

4. **Testing Gaps**
   - No tests for the main `Harness` type
   - Interceptor tests are minimal
   - No integration tests across packages

5. **Design Issues**
   - `EventWithFollowupOp` semantics unclear: `EventName` and `From` naming is confusing
   - Start event handling is inconsistent across different validation rules
   - Group position tracking is overly complex

---

## Recommended Improvements

### Phase 1: Fix Critical Implementation Gaps

#### 1. Complete the Harness Implementation

**Current State:**
```go
type Harness struct{}
func NewHarness(bus event.Bus, expected string, opts ...Option) *Harness
func (h *Harness) Run(t *testing.T, f func()) {}
```

**Recommended Implementation:**
```go
type Harness struct {
    runtime    *runtime.Runtime
    expected   string
    intercept  *runtime.Interceptor
}

func NewHarness(bus event.Bus, expected string, opts ...Option) *Harness {
    h := &Harness{}
    for _, opt := range opts {
        opt(h)
    }
    return h
}

func (h *Harness) Run(t *testing.T, f func()) {
    // Create interceptor wrapping the bus
    intercept := runtime.NewInterceptor(t, h.bus, h.clock)
    
    // Set up expected marble sequence
    recorder := intercept.EXPECT().FromMarble(h.expected)
    
    // Execute the test function
    f()
    
    // Finalize and check
    intercept.Finish()
}
```

#### 2. Fix Harness Options

Make options actually configure the harness:
```go
func WithPayloads(payloads map[string]event.Payload) Option {
    return func(h *Harness) {
        if h.runtime == nil {
            h.runtime = &runtime.Runtime{}
        }
        h.runtime.WithPayloadsMapping(payloads)
    }
}
```

#### 3. Complete Interceptor Implementation

Fix the group matching logic:
```go
func (r *InterceptorRecorder) failuresFromGroup(
    tick Tick, 
    activity []activityEntry, 
    ops []marble.Op, 
    ordered bool,
) []error {
    // Extract group operations (skip start/end markers)
    grpOps := extractGroupOps(ops)
    
    if ordered {
        return r.validateOrderedGroup(grpOps, activity)
    }
    return r.validateUnorderedGroup(grpOps, activity)
}

func extractGroupOps(ops []marble.Op) []marble.Op {
    // Skip first (start marker) and last (end marker)
    if len(ops) <= 2 {
        return nil
    }
    return ops[1 : len(ops)-1]
}

func (r *InterceptorRecorder) validateOrderedGroup(
    ops []marble.Op, 
    activity []activityEntry,
) []error {
    if len(ops) != len(activity) {
        return []error{fmt.Errorf("expected %d events, got %d", len(ops), len(activity))}
    }
    
    var errs []error
    for i, op := range ops {
        label := getOpLabel(op)
        if m := r.matchers[label]; !m.Match(activity[i].event) {
            errs = append(errs, fmt.Errorf("event %d: %s does not match %v", i, label, activity[i].event))
        }
    }
    return errs
}

func (r *InterceptorRecorder) validateUnorderedGroup(
    ops []marble.Op, 
    activity []activityEntry,
) []error {
    if len(ops) != len(activity) {
        return []error{fmt.Errorf("expected %d events, got %d", len(ops), len(activity))}
    }
    
    // Create a map of expected labels to matchers
    expected := make(map[string]event.Matcher)
    for _, op := range ops {
        label := getOpLabel(op)
        expected[label] = r.matchers[label]
    }
    
    // Check each actual event matches one of the expected
    matched := make(map[int]bool)
    var errs []error
    
    for _, act := range activity {
        found := false
        for i, op := range ops {
            if matched[i] {
                continue
            }
            label := getOpLabel(op)
            if r.matchers[label].Match(act.event) {
                matched[i] = true
                found = true
                break
            }
        }
        if !found {
            errs = append(errs, fmt.Errorf("unexpected event: %v", act.event))
        }
    }
    
    return errs
}
```

### Phase 2: Code Quality Improvements

#### 1. Remove Dead Code
- Delete commented out code in `interceptor.go` (lines 298-325, 375-391)
- Remove unused utility functions

#### 2. Improve Type Safety
Replace type assertions with type switches:
```go
// Current (problematic)
if o.Type() == marble.UnorderedGroupStartType {
    endPos = o.(marble.UnorderedGroupStartOp).EndPos
}

// Improved
switch o := op.(type) {
case marble.UnorderedGroupStartOp:
    endPos = o.EndPos
case marble.OrderedGroupStartOp:
    endPos = o.EndPos
}
```

#### 3. Simplify Group Position Tracking

The current position tracking in `Timeline.handleGroup` is overly complex. Simplify by:
- Using a stack-based approach for nested groups
- Separating concerns between parsing and execution
- Eliminating position adjustments

### Phase 3: Enhance Error Handling

#### 1. Better Error Messages
```go
// Current
return errors.Join(ErrMarbleSyntax, fmt.Errorf("unexpected character %q at %d", c, i))

// Improved
return fmt.Errorf("marble syntax error at position %d: unexpected character %q\n%s", 
    i, c, formatContext(marble, i, 10))

func formatContext(s string, pos, radius int) string {
    start := max(0, pos-radius)
    end := min(len(s), pos+radius+1)
    return fmt.Sprintf("...%s...", s[start:end])
}
```

#### 2. Use Proper Errors Instead of Panics
```go
// Current (in timeline.go line 29-31)
if err := marble.Validate(rowOps, marble.WaitlessGroupsRule{}); err != nil {
    panic(err)
}

// Improved
func NewTimeline(rowOps []marble.Op, opts ...TimelineOption) (*Timeline, error) {
    if err := marble.Validate(rowOps, marble.WaitlessGroupsRule{}); err != nil {
        return nil, fmt.Errorf("invalid operations for timeline: %w", err)
    }
    // ...
}
```

### Phase 4: API Improvements

#### 1. Rename EventWithFollowupOp Fields
The current naming is confusing:
```go
type EventWithFollowupOp struct {
    EventName string  // The NEW event being created
    From      string  // The event it's a followup OF
}
```

Should be:
```go
type EventWithFollowupOp struct {
    NewEvent string  // The new event
    OfEvent  string  // The event it follows up
}
```

#### 2. Add Builder Pattern for Marble Construction
```go
type MarbleBuilder struct {
    ops []marble.Op
}

func NewMarbleBuilder() *MarbleBuilder {
    return &MarbleBuilder{}
}

func (b *MarbleBuilder) Event(name string) *MarbleBuilder {
    b.ops = append(b.ops, marble.EventOp{Name: name})
    return b
}

func (b *MarbleBuilder) Wait() *MarbleBuilder {
    b.ops = append(b.ops, marble.WaitOp{})
    return b
}

func (b *MarbleBuilder) OrderedGroup(fn func(*MarbleBuilder)) *MarbleBuilder {
    // ...
}

func (b *MarbleBuilder) UnorderedGroup(fn func(*MarbleBuilder)) *MarbleBuilder {
    // ...
}

func (b *MarbleBuilder) String() (string, error) {
    // Convert ops back to marble string (need reverse parser)
}
```

#### 3. Add Marble String Formatter
```go
func Format(ops []marble.Op) string {
    var b strings.Builder
    for i, op := range ops {
        switch o := op.(type) {
        case marble.WaitOp:
            b.WriteRune('-')
        case marble.EventOp:
            b.WriteString(o.Name)
        case marble.StartEventOp:
            b.WriteRune('^')
        case marble.EventWithFollowupOp:
            b.WriteString(fmt.Sprintf("%s<-s", o.EventName, o.From))
        // Handle groups...
        }
        if i < len(ops)-1 && !isGroupBoundary(op, ops[i+1]) {
            // Add separator if needed
        }
    }
    return b.String()
}
```

---

## Enhanced Recommendations

### After Deeper Analysis

#### 1. Semantic Validation Should Be More Extensible

**Problem:** Current validation uses a list of rules, but adding new rules requires modifying the validation logic.

**Solution:** Make the rule system more pluggable:
```go
type ValidationContext struct {
    ops     []marble.Op
    current int
    errors  []error
}

type Rule interface {
    Validate(ctx *ValidationContext) error
}

type CompositeRule struct {
    rules []Rule
}

func (r CompositeRule) Validate(ctx *ValidationContext) error {
    for _, rule := range r.rules {
        if err := rule.Validate(ctx); err != nil {
            return err
        }
    }
    return nil
}

// Allow chaining rules
func And(rules ...Rule) Rule {
    return CompositeRule{rules: rules}
}

// Allow optional rules
func Or(rules ...Rule) Rule {
    return OptionalRule{rules: rules}
}
```

#### 2. Timeline Construction Should Handle Edge Cases

**Problem:** Complex nested groups with mixed ordered/unordered groups can cause position tracking issues.

**Solution:** Redesign timeline construction:
```go
type TimelineBuilder struct {
    ops    []marble.Op
    ticks  []Tick
    stack  []groupContext
}

type groupContext struct {
    startPos   int
    ops        []marble.Op
    ordered    bool
    parentPos  int
}

func (b *TimelineBuilder) Build() ([]Tick, error) {
    for i, op := range b.ops {
        switch o := op.(type) {
        case marble.OrderedGroupStartOp, marble.UnorderedGroupStartOp:
            b.pushGroup(i, o.Type() == marble.OrderedGroupStartType)
        case marble.OrderedGroupEndOp, marble.UnorderedGroupEndOp:
            tick, err := b.popGroup(i, o.Type() == marble.OrderedGroupEndType)
            if err != nil {
                return nil, err
            }
            b.ticks = append(b.ticks, tick)
        case marble.WaitOp:
            b.ticks = append(b.ticks, Tick{Duration: b.tickDuration, Ops: []marble.Op{op}})
        default:
            b.ticks = append(b.ticks, Tick{Duration: b.tickDuration, Ops: []marble.Op{op}})
        }
    }
    return b.ticks, nil
}
```

#### 3. Event Resolution Should Be More Flexible

**Problem:** Current approach of trying payload map, then event map, then default is inflexible.

**Solution:** Use a resolution strategy pattern:
```go
type EventResolver interface {
    Resolve(label string) (event.Event, bool)
}

type ChainResolver struct {
    resolvers []EventResolver
}

func (r ChainResolver) Resolve(label string) (event.Event, bool) {
    for _, resolver := range r.resolvers {
        if evt, ok := resolver.Resolve(label); ok {
            return evt, true
        }
    }
    return event.Event{}, false
}

type PayloadMapResolver struct {
    payloads map[string]event.Payload
}

func (r PayloadMapResolver) Resolve(label string) (event.Event, bool) {
    if pl, ok := r.payloads[label]; ok {
        return event.New(pl), true
    }
    return event.Event{}, false
}

type EventMapResolver struct {
    events map[string]event.Event
}

func (r EventMapResolver) Resolve(label string) (event.Event, bool) {
    evt, ok := r.events[label]
    return evt, ok
}

type DefaultResolver struct{}

func (r DefaultResolver) Resolve(label string) (event.Event, bool) {
    return event.New(DefaultPayload(label)), true
}
```

#### 4. Interceptor Should Support Multiple Verification Modes

**Problem:** Current implementation mixes strict and lenient modes.

**Solution:** Make it explicit:
```go
type VerificationMode int

const (
    StrictMode VerificationMode = iota
    LenientMode
    OrderedOnlyMode
)

type InterceptorConfig struct {
    Mode           VerificationMode
    Timeout        time.Duration
    FailFast       bool
    MatchAllEvents bool  // Whether to verify all events or just expected ones
}

func (r *InterceptorRecorder) WithConfig(cfg InterceptorConfig) *InterceptorRecorder {
    r.config = cfg
    return r
}
```

#### 5. Add Support for Time Constraints

**Problem:** Current implementation only checks event ordering, not timing.

**Solution:** Add time-based assertions:
```go
type TimeConstraint struct {
    MinDuration time.Duration
    MaxDuration time.Duration
}

type InterceptorRecorder struct {
    // ... existing fields
    timeConstraints map[string]TimeConstraint
}

func (r *InterceptorRecorder) Within(d time.Duration) *InterceptorRecorder {
    // Apply time constraint to next event
}

func (r *InterceptorRecorder) Between(min, max time.Duration) *InterceptorRecorder {
    // Apply time range constraint
}

func (r *InterceptorRecorder) Failures() []error {
    // ... existing logic
    
    // Add time constraint validation
    for i, tick := range r.timeline.Ticks() {
        if constraint, ok := r.timeConstraints[getOpLabel(tick.Ops[0])]; ok {
            actualDuration := getActualDuration(i)
            if actualDuration < constraint.MinDuration || actualDuration > constraint.MaxDuration {
                errs = append(errs, fmt.Errorf("event at position %d: duration %v not in range [%v, %v]", 
                    i, actualDuration, constraint.MinDuration, constraint.MaxDuration))
            }
        }
    }
    return errs
}
```

#### 6. Add Support for Conditional Events

**Problem:** Current marble language doesn't support conditional or optional events.

**Solution:** Extend the language:
```
// Optional event (may or may not occur)
a?

// Alternative events (one of these must occur)
a|b|c

// Repeated events (exactly n times)
a*3

// Repeated events (at least n times)
a+3

// Repeated events (0 or more times)
a*
```

This would require:
1. Extending the parser
2. Adding new Op types
3. Updating the validation rules
4. Modifying the runtime execution

#### 7. Improve Testability

**Problem:** Complex internal state makes testing difficult.

**Solution:** 
- Expose more internal state for testing
- Add String() methods to Op types for debugging
- Add pretty-printing for timelines
- Create test helpers

```go
func (o EventOp) String() string {
    return o.Name
}

func (o WaitOp) String() string {
    return "-"
}

func (o StartEventOp) String() string {
    return "^"
}

func (o EventWithFollowupOp) String() string {
    return fmt.Sprintf("%s<-s", o.EventName, o.From)
}

func DebugTimeline(tl *Timeline) string {
    var b strings.Builder
    for i, tick := range tl.Ticks() {
        b.WriteString(fmt.Sprintf("Tick %d (%v): ", i, tick.Duration))
        for _, op := range tick.Ops {
            b.WriteString(fmt.Sprintf("%s ", op))
        }
        b.WriteRune('\n')
    }
    return b.String()
}
```

#### 8. Add Performance Optimization

**Problem:** Parsing and validation happens every time a marble sequence is used.

**Solution:** Cache parsed sequences:
```go
type MarbleCache struct {
    cache map[string]*ParsedMarble
    mu    sync.RWMutex
}

type ParsedMarble struct {
    ops      []marble.Op
    timeline *Timeline
    hash     string
}

func (c *MarbleCache) Get(marble string) (*ParsedMarble, error) {
    c.mu.RLock()
    if pm, ok := c.cache[marble]; ok {
        c.mu.RUnlock()
        return pm, nil
    }
    c.mu.RUnlock()
    
    ops, err := marble.Parse(marble)
    if err != nil {
        return nil, err
    }
    
    tl := NewTimeline(ops)
    pm := &ParsedMarble{
        ops:      ops,
        timeline: tl,
        hash:     hashMarble(marble),
    }
    
    c.mu.Lock()
    c.cache[marble] = pm
    c.mu.Unlock()
    
    return pm, nil
}
```

#### 9. Add Better Integration with Testing Frameworks

**Problem:** Current integration with testing is minimal.

**Solution:** Add more testing helpers:
```go
// Assert that a function produces the expected marble sequence
func AssertMarble(t *testing.T, bus event.Bus, expected string, fn func()) {
    harness := NewHarness(bus, expected)
    harness.Run(t, fn)
}

// Assert with custom matchers
func AssertMarbleWithMatchers(t *testing.T, bus event.Bus, expected string, 
    matchers map[string]event.Matcher, fn func()) {
    harness := NewHarness(bus, expected, WithMatchers(matchers))
    harness.Run(t, fn)
}

// Assert that NO events matching the marble sequence occur
func AssertNoMarble(t *testing.T, bus event.Bus, unexpected string, fn func()) {
    // Create an interceptor that verifies no matching events occur
}
```

#### 10. Document the Architecture Better

**Problem:** Code structure and relationships are not well documented.

**Solution:** Add architecture documentation (this document) and code-level documentation:
- Package-level comments explaining purpose and relationships
- Function-level comments explaining behavior and contracts
- Examples in doc comments

---

## Complete Architecture Scheme

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                           eventest Package                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    │
│  │    Harness       │───▶│   Interceptor    │───▶│     Runtime      │    │
│  │  (User API)      │    │   (Verification) │    │   (Execution)    │    │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘    │
│           │                    │                      │                │
│           │ Configuration      │ Expected Events      │ Event Publishing│
│           ▼                    ▼                      ▼                │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    │
│  │  Matcher Helpers │    │  Activity Log    │    │    Timeline     │    │
│  │  (PayloadEq, etc)│    │   (Event Timing)│    │   (Tick Plan)    │    │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘    │
│                                                                         │
└─────────────────────────────────────────────────────────────────────┘
        │                           │                            │
        ▼                           ▼                            ▼
┌─────────────────┐   ┌─────────────────┐            ┌─────────────────┐
│   User's Test    │   │   event.Bus      │            │ event.Payload  │
│   Function       │   │   (External)     │            │   (External)   │
└─────────────────┘   └─────────────────┘            └─────────────────┘
```

### Internal Package Structure

```
eventest/internal/
├── marble/                          # Language Processing
│   ├── parser.go                    # String → AST
│   ├── op.go                        # AST Node Types
│   ├── semantic.go                  # Validation Rules
│   └── format.go (NEW)              # AST → String
│
└── runtime/                         # Execution Engine
    ├── runtime.go                   # Main Runtime
    ├── timeline.go                  # Timeline Construction
    ├── clock.go                     # Virtual Clock
    ├── interceptor.go               # Event Verification
    └── option.go                    # Configuration
```

### Data Flow Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                            DATA FLOW DIAGRAM                              │
└──────────────────────────────────────────────────────────────────────┘

Marble String
    "^a-(bc)[de]f"
         │
         ▼
┌─────────────────┐
│   Parser         │  ← marble.Parse()
│   Lexer + AST    │
└────────┬────────┘
         │ []Op (AST)
         ▼
┌─────────────────┐
│ Semantic Check  │  ← marble.Validate()
│ Rules Engine    │
└────────┬────────┘
         │ Validated []Op
         ▼
┌─────────────────┐
│  Timeline Builder│  ← runtime.NewTimeline()
│  Tick Generator  │
└────────┬────────┘
         │ []Tick (Execution Plan)
         ▼
┌─────────────────┐
│    Runtime       │  ← runtime.Run()
│  Execution Engine │
└────────┬────────┘
         │ Executes Ticks
         ▼
┌─────────────────┐
│   Event Bus      │  ← bus.Publish()
│   (External)     │
└────────┬────────┘
         │ Published Events
         ▼
┌─────────────────┐
│  Interceptor     │  ← Capture Events
│  (Verification)   │
└────────┬────────┘
         │ Comparison with Expected
         ▼
    Test Result (PASS/FAIL)
```

### Class/Component Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                           COMPONENT RELATIONSHIPS                         │
└──────────────────────────────────────────────────────────────────────┘

┌─────────────────────┐       ┌─────────────────────┐
│      Op (Interface)  │◄──────│      WaitOp          │
└─────────┬───────────┘       └─────────────────────┘
          ▲                     ┌─────────────────────┐
          │                     │     EventOp          │
          │                     └─────────────────────┘
          │                     ┌─────────────────────┐
          │                     │  StartEventOp        │
          │                     └─────────────────────┘
          │                     ┌─────────────────────┐
          │                     │ EventWithFollowupOp │
          │                     └─────────────────────┘
          │                     ┌─────────────────────┐
          └─────────────────────┤  OrderedGroupStartOp │
                                └─────────┬───────────┘
                                          ▲
┌─────────────────────┐               │
│   Parser             │───────────────┘
└─────────────────────┘
         ▲
         │
┌─────────────────────┐
│   Semantic Validator │
└──────────┬───────────┘
           ▲
           │
┌─────────────────────┐
│    Rule (Interface)  │
└─────────┬───────────┘
          ▲
┌─────────┴─────────┐
│  NotEmptyRule       │
├─────────────────────┤
│ StartEventRule*     │
├─────────────────────┤
│ WaitlessGroupsRule  │
└─────────────────────┘

┌─────────────────────┐       ┌─────────────────────┐
│    Runtime           │◄──────│    Timeline          │
└──────────┬───────────┘       └──────────┬───────────┘
           ▲                             ▲
           │                             │
┌──────────┴───────────┐     ┌───────────┴───────────┐
│    RunningSession     │     │       Tick            │
└─────────────────────┘     └─────────────────────┘
           ▲
           │
┌──────────┴───────────┐
│    Clock             │
└─────────────────────┘

┌─────────────────────┐       ┌─────────────────────┐
│   Interceptor        │◄──────│  InterceptorRecorder │
└──────────┬───────────┘       └─────────────────────┘
           ▲
           │ Implements
           ▼
┌─────────────────────┐
│    event.Bus         │
└─────────────────────┘
```

### Sequence Diagram (Test Execution)

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   Test    │     │ Harness  │     │Runtime   │     │ Bus      │
└─────┬─────┘     └─────┬─────┘     └─────┬─────┘     └─────┬─────┘
      │                │                │               │
      │ NewHarness()   │                │               │
      │──────────────▶│                │               │
      │                │                │               │
      │ harness.Run()  │                │               │
      │──────────────▶│                │               │
      │                │ Run()           │               │
      │                │───────────────▶│               │
      │                │                │ Parse()       │
      │                │                │───            │
      │                │                │ Validate()    │
      │                │                │───            │
      │                │                │ NewTimeline() │
      │                │                │───            │
      │                │                │ Run Session   │
      │                │                │──────────────▶│
      │                │                │               │ Publish()
      │                │                │◀──────────────│
      │                │                │ Next()        │
      │                │                │───            │
      │                │ Intercept      │               │
      │                │───────────────▶│               │
      │                │                │◀──────────────│
      │                │ Compare        │               │
      │                │ Results        │               │
      │                │───────────────▶│               │
      │                │                │               │
      │    Return       │                │               │
      │◀───────────────│                │               │
      │                │                │               │
```

### State Machines

#### Parser State Machine

```
                    ┌─────────────┐
                    │   Start     │
                    └──────┬──────┘
                           │
          ┌────────────────┼────────────────┐
          │                 │                 │
          ▼                 ▼                 ▼
    ┌─────────┐       ┌─────────┐       ┌─────────┐
    │ Event   │       │ Wait    │       │ Group   │
    │ (a-z)   │       │ (-, _)  │       │ (, [    │
    └────┬────┘       └────┬────┘       └────┬────┘
         │                 │                 │
         ▼                 ▼                 ▼
    ┌─────────────────────────────────────────┐
    │           Append Op to result            │
    └─────────────────────────────────────────┘
```

#### Timeline Builder State Machine

```
                     ┌─────────────┐
                     │   Idle      │
                     └──────┬──────┘
                            │
           ┌────────────────┼────────────────┐
           │                 │                 │
           ▼                 ▼                 ▼
    ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
    │ In Event    │   │ In Wait     │   │ In Group    │
    │             │   │             │   │             │
    └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
           │                 │                 │
           │ Add to Tick     │ Add to Tick     │ Push Context
           │                 │                 │
           ▼                 ▼                 ▼
    ┌─────────────────────────────────────────┐
    │           Next Tick / Continue            │
    └─────────────────────────────────────────┘
```

### Layered Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                            │
│  (User's test code using eventest.Harness)                         │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       Domain Layer (eventest)                        │
│  - Harness: Main API for users                                      │
│  - Matchers: Event matching utilities                                │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   Internal - Marble Language                          │
│  - Parser: String parsing                                            │
│  - Semantic: Validation                                              │
│  - Op: AST definitions                                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   Internal - Runtime Engine                           │
│  - Runtime: Execution engine                                        │
│  - Timeline: Tick generation                                         │
│  - Clock: Time management                                            │
│  - Interceptor: Event verification                                   │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                   External Dependencies                               │
│  - event: Event bus and types                                        │
│  - testing: Go testing framework                                     │
│  - gomock: Mocking framework                                          │
└─────────────────────────────────────────────────────────────────┘
```

### Recommended Refactored Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                        REFACTORED ARCHITECTURE                           │
└──────────────────────────────────────────────────────────────────────┘

eventest/
├── harness.go           # Complete harness implementation
├── matchers.go          # Matchers (keep)
├── marble.go            # Public marble API (NEW)
│
├── internal/
│   ├── marble/
│   │   ├── parser.go     # Parser
│   │   ├── op.go         # Op types
│   │   ├── semantic.go   # Validation
│   │   └── format.go     # Formatter (NEW)
│   │
│   └── engine/
│       ├── runtime/
│       │   └── runtime.go    # Runtime
│       ├── timeline/
│       │   ├── builder.go # Timeline builder (REFACTORED)
│       │   └── tick.go    # Tick type (EXISTING)
│       ├── clock/      # Clock
│       │   └── clock.go      # Clock
│       ├── interceptor/
│       │   ├── interceptor.go # Interceptor
│       │   ├── recorder.go    # Recorder (NEW - split from interceptor)
│       │   └── validator.go   # Event validator (NEW)
│       └── resolver/        # (NEW)
│           ├── chain.go     # Chain resolver
│           ├── payload.go   # Payload resolver
│           └── event.go      # Event resolver
├── assert.go        # Test assertions
└── helpers.go       # Test helpers
```

### Key Design Patterns Applied

1. **Interpreter Pattern**: Parser + AST + Execution
2. **Strategy Pattern**: Different validation rules
3. **Builder Pattern**: Timeline construction
4. **Composite Pattern**: Group operations
5. **Observer Pattern**: Event interception
6. **Chain of Responsibility**: Event resolution
7. **Template Method**: Timeline building process

### Error Handling Strategy

```
┌──────────────────────────────────────────────────────────────────┐
│                        ERROR HANDLING HIERARCHY                       │
└──────────────────────────────────────────────────────────────────┘

                    ┌─────────────────┐
                    │   Error         │ (root)
                    └────────┬────────┘
                             │
          ┌──────────────────┼──────────────────┐
          │                  │                  │
          ▼                  ▼                  ▼
   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐
   │ ErrMarble    │   │ ErrSemantic  │   │ ErrRuntime   │
   │ Syntax Error │   │ Validation   │   │ Execution    │
   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘
          │                  │                  │
          │ Errors.Join()    │ Errors.Join()    │ Errors.Join()
          ▼                  ▼                  ▼
   Specific error  Specific error  Specific error
   messages       messages       messages
```

### Performance Characteristics

| Operation | Time Complexity | Space Complexity |
|-----------|----------------|------------------|
| Parse | O(n) | O(n) |
| Validate | O(n) | O(1) |
| Build Timeline | O(n) | O(n) |
| Execute Tick | O(1) | O(1) |
| Intercept Event | O(1) | O(n) |
| Verify Sequence | O(n) | O(n) |

Where n = length of marble string / number of operations.

### Thread Safety Considerations

Current implementation:
- **Parser**: Thread-safe (pure function)
- **Validator**: Thread-safe (pure function)
- **Timeline**: Thread-safe after construction
- **Runtime**: NOT thread-safe (shared state)
- **Interceptor**: NOT thread-safe (captures events)

Recommended improvements:
- Make Runtime thread-safe with mutex
- Or document that Runtime is single-threaded
- Use channel-based communication for thread safety

---

## Summary of Recommendations

### Priority 1 (Critical - Must Fix)
1. ✅ Complete `Harness` implementation
2. ✅ Fix `Harness` options to actually configure behavior
3. ✅ Complete `Interceptor` group validation logic
4. ✅ Remove dead code and commented code
5. ✅ Fix runtime integration with harness

### Priority 2 (High - Should Fix)
1. ✅ Improve error messages with context
2. ✅ Replace panics with proper error returns
3. ✅ Use type switches instead of type assertions
4. ✅ Add String() methods to Op types for debugging
5. ✅ Add proper documentation

### Priority 3 (Medium - Nice to Have)
1. ✅ Rename `EventWithFollowupOp` fields for clarity
2. ✅ Add marble formatter (AST → string)
3. ✅ Add builder pattern for marble construction
4. ✅ Make event resolution more flexible
5. ✅ Add time-based constraints to interceptor

### Priority 4 (Low - Future Enhancements)
1. ✅ Add support for optional/conditional events
2. ✅ Add caching for parsed marble strings
3. ✅ Add more testing helpers
4. ✅ Add conditional events to language
5. ✅ Add support for event metadata in marble

---

## Conclusion

The marble language and its implementation in the `eventest` package provide a powerful and expressive way to describe and test event sequences. However, the current implementation has several critical gaps, particularly in the harness and interceptor components, that need to be addressed before it can be used effectively.

The architecture is fundamentally sound, following good separation of concerns between parsing, validation, timeline construction, and execution. With the recommended improvements, the system can become more robust, flexible, and maintainable.

The key areas to focus on are:
1. **Completing the core implementation** (harness, interceptor)
2. **Improving code quality** (error handling, type safety, removing dead code)
3. **Enhancing the API** (builder pattern, better error messages)
4. **Adding advanced features** (time constraints, conditional events, caching)

The refactored architecture maintains the current structure while addressing the identified issues and providing a clearer separation of concerns.
