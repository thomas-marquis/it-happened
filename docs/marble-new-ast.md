# Marble AST Redesign: Hierarchical AST with Visitor Pattern

## Executive Summary

The current Marble language implementation uses a **flat operation list with position markers** to represent the Abstract Syntax Tree (AST). This approach causes significant complexity in timeline building, runtime execution, and interceptor validation due to manual position tracking for groups.

**Recommended Solution:** Adopt a **hierarchical AST structure with the Visitor pattern** to eliminate position tracking, simplify all consumers, and improve maintainability.

---

## Problem Analysis

### Current AST Structure

The existing implementation represents marble sequences as flat lists of operations with position markers:

```go
type Op interface {
    Type() OpType
}

// Example: "[ab]"
[]Op{
    OrderedGroupStartOp{EndPos: 3},  // index 0
    EventOp{Name: "a"},               // index 1
    EventOp{Name: "b"},               // index 2
    OrderedGroupEndOp{StartPos: 0},   // index 3
}
```

### Key Problems

#### 1. Position Tracking Complexity
- Every group operation must track `StartPos` and `EndPos`
- Nested groups require position adjustments relative to parent groups
- Timeline builder must reconstruct group boundaries from positions

#### 2. Timeline Building State Machine
The `Timeline.buildTicks()` and `handleGroup()` functions contain complex logic:
```go
// From timeline.go lines 54-186
func (t *Timeline) buildTicks(rowOps []marble.Op) []Tick {
    var (
        current    int
        ticks      []Tick
        grpExitPos int
        pos        int
    )
    // Complex position-based logic follows...
}

func (t *Timeline) handleGroup(ops []marble.Op, pos *int, tickStartPos int) []marble.Op {
    var (
        exitPos       int
        parts         [][]marble.Op
        startPos      = *pos
        shuffleNeeded bool
    )
    // More complex position tracking...
}
```

#### 3. Interceptor Validation Issues
The `InterceptorRecorder.failuresFromGroup()` method (lines 226-288) contains:
- Incomplete implementation
- Complex position-based group boundary reconstruction
- TODO comments indicating known problems
- Duplicate logic for strict vs lenient modes

#### 4. Nested Group Fragility
With deeply nested groups, position tracking becomes:
- Error-prone (off-by-one errors)
- Hard to debug
- Difficult to extend

#### 5. Code Duplication
Similar position-based logic appears in:
- `timeline.go`: `buildTicks()`, `handleGroup()`
- `interceptor.go`: `failuresFromGroup()`, `Failures()`
- `semantic.go`: `isStartEvent()`

---

## Recommended Solution: Hierarchical AST

### Core Concept

Replace the flat operation list with a **tree-structured AST** where groups are actual composite nodes containing their children.

### New AST Types

#### Base Interface

```go
// Node represents any element in the marble AST
type Node interface {
    // Accept allows the Visitor pattern to traverse the AST
    Accept(Visitor)
    
    // Position provides source location (optional, for error reporting)
    Position() Position
}

// Position represents a source location
type Position struct {
    Line   int
    Column int
    Offset int
}
```

#### Leaf Nodes (Terminal)

```go
// EventNode represents a named event (e.g., "a", "/eventName")
type EventNode struct {
    Name string
    pos  Position
}

func (n *EventNode) Accept(v Visitor) {
    v.VisitEvent(n)
}

func (n *EventNode) Position() Position {
    return n.pos
}

// WaitNode represents a wait operation ("-" or "_")
type WaitNode struct {
    pos Position
}

func (n *WaitNode) Accept(v Visitor) {
    v.VisitWait(n)
}

func (n *WaitNode) Position() Position {
    return n.pos
}

// StartNode represents the start event ("^")
type StartNode struct {
    pos Position
}

func (n *StartNode) Accept(v Visitor) {
    v.VisitStart(n)
}

func (n *StartNode) Position() Position {
    return n.pos
}

// FollowupNode represents a followup relationship (e.g., "a<-b")
type FollowupNode struct {
    NewEvent string  // The new event being created
    OfEvent  string  // The event it's a followup of
    pos      Position
}

func (n *FollowupNode) Accept(v Visitor) {
    v.VisitFollowup(n)
}

func (n *FollowupNode) Position() Position {
    return n.pos
}
```

#### Composite Nodes (Non-Terminal)

```go
// SequenceNode represents a sequence of nodes
type SequenceNode struct {
    Children []Node
    pos      Position
}

func (n *SequenceNode) Accept(v Visitor) {
    v.VisitSequence(n)
}

func (n *SequenceNode) Position() Position {
    return n.pos
}

// GroupNode represents a grouped sequence (either ordered or unordered)
type GroupNode struct {
    Children []Node
    Ordered  bool  // true = ordered [ ], false = unordered ( )
    pos      Position
}

func (n *GroupNode) Accept(v Visitor) {
    v.VisitGroup(n)
}

func (n *GroupNode) Position() Position {
    return n.pos
}
```

### AST Examples

#### Example 1: Simple Sequence "abc"

**Current (Flat):**
```go
[]Op{
    EventOp{Name: "a"},
    EventOp{Name: "b"},
    EventOp{Name: "c"},
}
```

**New (Hierarchical):**
```go
SequenceNode{
    Children: []Node{
        EventNode{Name: "a"},
        EventNode{Name: "b"},
        EventNode{Name: "c"},
    },
}
```

#### Example 2: Ordered Group "[ab]"

**Current (Flat):**
```go
[]Op{
    OrderedGroupStartOp{EndPos: 3},
    EventOp{Name: "a"},
    EventOp{Name: "b"},
    OrderedGroupEndOp{StartPos: 0},
}
```

**New (Hierarchical):**
```go
SequenceNode{
    Children: []Node{
        GroupNode{
            Children: []Node{
                EventNode{Name: "a"},
                EventNode{Name: "b"},
            },
            Ordered: true,
        },
    },
}
```

#### Example 3: Complex Nested "a-[b(cd)e]f"

**Current (Flat):**
```go
[]Op{
    EventOp{Name: "a"},
    WaitOp{},
    OrderedGroupStartOp{EndPos: 8},
    EventOp{Name: "b"},
    UnorderedGroupStartOp{EndPos: 6},
    EventOp{Name: "c"},
    EventOp{Name: "d"},
    UnorderedGroupEndOp{StartPos: 4},
    EventOp{Name: "e"},
    OrderedGroupEndOp{StartPos: 2},
    EventOp{Name: "f"},
}
```

**New (Hierarchical):**
```go
SequenceNode{
    Children: []Node{
        EventNode{Name: "a"},
        WaitNode{},
        GroupNode{
            Ordered: true,
            Children: []Node{
                EventNode{Name: "b"},
                GroupNode{
                    Ordered: false,
                    Children: []Node{
                        EventNode{Name: "c"},
                        EventNode{Name: "d"},
                    },
                },
                EventNode{Name: "e"},
            },
        },
        EventNode{Name: "f"},
    },
}
```

---

## Visitor Pattern Implementation

### Visitor Interface

```go
// Visitor defines the interface for AST traversal
type Visitor interface {
    // Leaf nodes
    VisitEvent(*EventNode)
    VisitWait(*WaitNode)
    VisitStart(*StartNode)
    VisitFollowup(*FollowupNode)
    
    // Composite nodes
    VisitSequence(*SequenceNode)
    VisitGroup(*GroupNode)
}
```

### Benefits of Visitor Pattern

1. **Separation of Concerns**: Each operation on the AST is a separate visitor
2. **Easy Extension**: Add new operations without modifying existing code
3. **Single Responsibility**: Each visitor has one clear purpose
4. **Type Safety**: Compiler enforces all node types are handled

### Example Visitors

#### 1. Timeline Builder Visitor

```go
// TimelineBuilder builds a timeline from the AST
type TimelineBuilder struct {
    ticks    []Tick
    current  []Op
    depth    int
    err      error
}

func BuildTimeline(root Node) ([]Tick, error) {
    builder := &TimelineBuilder{}
    root.Accept(builder)
    return builder.ticks, builder.err
}

func (b *TimelineBuilder) VisitSequence(n *SequenceNode) {
    for _, child := range n.Children {
        child.Accept(b)
    }
}

func (b *TimelineBuilder) VisitGroup(n *GroupNode) {
    // Create group start op
    var startOp Op
    if n.Ordered {
        startOp = OrderedGroupStartOp{}
    } else {
        startOp = UnorderedGroupStartOp{}
    }
    
    // Create group end op
    var endOp Op
    if n.Ordered {
        endOp = OrderedGroupEndOp{}
    } else {
        endOp = UnorderedGroupEndOp{}
    }
    
    // Build the tick for this group
    tickOps := []Op{startOp}
    
    // Process children (they add to tickOps)
    oldCurrent := b.current
    b.current = tickOps
    
    for _, child := range n.Children {
        child.Accept(b)
    }
    
    // Finalize the tick
    b.current = append(b.current, endOp)
    b.ticks = append(b.ticks, Tick{
        Duration: DefaultTickDuration,
        Ops:     b.current,
    })
    b.current = oldCurrent
}

func (b *TimelineBuilder) VisitEvent(n *EventNode) {
    if b.current == nil {
        // Top-level event = new tick
        b.ticks = append(b.ticks, Tick{
            Duration: DefaultTickDuration,
            Ops:     []Op{EventOp{Name: n.Name}},
        })
    } else {
        // Inside group = add to current tick
        b.current = append(b.current, EventOp{Name: n.Name})
    }
}

func (b *TimelineBuilder) VisitWait(n *WaitNode) {
    if b.current == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: DefaultTickDuration,
            Ops:     []Op{WaitOp{}},
        })
    } else {
        b.current = append(b.current, WaitOp{})
    }
}

func (b *TimelineBuilder) VisitStart(n *StartNode) {
    b.ticks = append(b.ticks, Tick{
        Duration: DefaultTickDuration,
        Ops:     []Op{StartEventOp{}},
    })
}

func (b *TimelineBuilder) VisitFollowup(n *FollowupNode) {
    op := EventWithFollowupOp{
        EventName: n.NewEvent,
        From:      n.OfEvent,
    }
    if b.current == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: DefaultTickDuration,
            Ops:     []Op{op},
        })
    } else {
        b.current = append(b.current, op)
    }
}
```

#### 2. Semantic Validator Visitor

```go
type SemanticValidator struct {
    errors     []error
    context    []string  // Stack of contexts for error reporting
    startCount int
    inGroup    bool
}

func ValidateSemantics(root Node, rules ...Rule) error {
    validator := &SemanticValidator{}
    
    // Apply rules
    for _, rule := range rules {
        if err := applyRule(validator, root, rule); err != nil {
            validator.errors = append(validator.errors, err)
        }
    }
    
    if len(validator.errors) > 0 {
        return errors.Join(validator.errors...)
    }
    return nil
}

func (v *SemanticValidator) VisitSequence(n *SequenceNode) {
    v.context = append(v.context, "sequence")
    for _, child := range n.Children {
        child.Accept(v)
    }
    v.context = v.context[:len(v.context)-1]
}

func (v *SemanticValidator) VisitGroup(n *GroupNode) {
    oldInGroup := v.inGroup
    v.inGroup = true
    v.context = append(v.context, fmt.Sprintf("group(%v)", n.Ordered))
    
    // Check for wait in group
    v.checkWaitInGroup(n)
    
    for _, child := range n.Children {
        child.Accept(v)
    }
    
    v.context = v.context[:len(v.context)-1]
    v.inGroup = oldInGroup
}

func (v *SemanticValidator) checkWaitInGroup(n *GroupNode) {
    for _, child := range n.Children {
        switch c := child.(type) {
        case *WaitNode:
            v.errors = append(v.errors, 
                fmt.Errorf("wait not allowed in %v group at %v", 
                    map[bool]string{true: "ordered", false: "unordered"}[n.Ordered],
                    c.Position()))
        case *GroupNode:
            v.checkWaitInGroup(c)  // Recursively check nested groups
        }
    }
}

func (v *SemanticValidator) VisitStart(n *StartNode) {
    v.startCount++
    if v.startCount > 1 {
        v.errors = append(v.errors, 
            fmt.Errorf("multiple start events at %v", n.Position()))
    }
}

func (v *SemanticValidator) VisitEvent(n *EventNode) {}
func (v *SemanticValidator) VisitWait(n *WaitNode) {}
func (v *SemanticValidator) VisitFollowup(n *FollowupNode) {}
```

#### 3. Interceptor Validation Visitor

```go
type InterceptorValidator struct {
    timeline    *Timeline
    activity    []activityEntry
    matchers    map[string]event.Matcher
    errors      []error
    currentTick int
}

func (v *InterceptorValidator) VisitSequence(n *SequenceNode) {
    for i, child := range n.Children {
        v.currentTick = i
        child.Accept(v)
    }
}

func (v *InterceptorValidator) VisitGroup(n *GroupNode) {
    // Get the tick corresponding to this group
    if v.currentTick >= len(v.timeline.Ticks()) {
        v.errors = append(v.errors, 
            fmt.Errorf("group at tick %d: no corresponding tick", v.currentTick))
        return
    }
    
    tick := v.timeline.Ticks()[v.currentTick]
    tickActivity := selectActivityEntriesByRange(
        v.activity, 
        tick.Duration*time.Duration(v.currentTick), 
        tick.Duration*time.Duration(v.currentTick+1))
    
    // Validate based on group type
    if n.Ordered {
        v.validateOrderedGroup(n, tickActivity)
    } else {
        v.validateUnorderedGroup(n, tickActivity)
    }
}

func (v *InterceptorValidator) validateOrderedGroup(n *GroupNode, activity []activityEntry) {
    if len(n.Children) != len(activity) {
        v.errors = append(v.errors,
            fmt.Errorf("ordered group: expected %d events, got %d", 
                len(n.Children), len(activity)))
        return
    }
    
    for i, child := range n.Children {
        if eventNode, ok := child.(*EventNode); ok {
            if i >= len(activity) {
                break
            }
            if m := v.matchers[eventNode.Name]; !m.Match(activity[i].event) {
                v.errors = append(v.errors,
                    fmt.Errorf("event %d: expected %s, got %v", i, eventNode.Name, activity[i].event))
            }
        }
    }
}

func (v *InterceptorValidator) validateUnorderedGroup(n *GroupNode, activity []activityEntry) {
    if len(n.Children) != len(activity) {
        v.errors = append(v.errors,
            fmt.Errorf("unordered group: expected %d events, got %d", 
                len(n.Children), len(activity)))
        return
    }
    
    // Create map of expected matchers
    expected := make(map[string]event.Matcher)
    for _, child := range n.Children {
        if eventNode, ok := child.(*EventNode); ok {
            expected[eventNode.Name] = v.matchers[eventNode.Name]
        }
    }
    
    // Match actual events to expected
    matched := make(map[string]bool)
    for _, act := range activity {
        found := false
        for name, matcher := range expected {
            if !matched[name] && matcher.Match(act.event) {
                matched[name] = true
                found = true
                break
            }
        }
        if !found {
            v.errors = append(v.errors,
                fmt.Errorf("unexpected event in unordered group: %v", act.event))
        }
    }
}

func (v *InterceptorValidator) VisitEvent(n *EventNode) {
    // Top-level events are validated by tick position
}
func (v *InterceptorValidator) VisitWait(n *WaitNode) {}
func (v *InterceptorValidator) VisitStart(n *StartNode) {}
func (v *InterceptorValidator) VisitFollowup(n *FollowupNode) {}
```

---

## Parser Adaptation

### Modified Parser

```go
// Parse parses a marble string and returns the root AST node
func Parse(marble string) (Node, error) {
    var pos int
    return parse(marble, &pos)
}

func parse(marble string, pos *int) (Node, error) {
    var children []Node
    
    for *pos < len(marble) {
        c := marble[*pos]
        
        switch {
        case c == ' ', c == '\t', c == '\n', c == '\r':
            *pos++
            continue
            
        case c == '^':
            if *pos != 0 {
                return nil, fmt.Errorf("unexpected ^ at position %d", *pos)
            }
            children = append(children, &StartNode{pos: Position{Offset: *pos}})
            *pos++
            
        case c == '-':
            children = append(children, &WaitNode{pos: Position{Offset: *pos}})
            *pos++
            
        case c == '_':
            // Consume all consecutive underscores
            start := *pos
            for *pos < len(marble) && marble[*pos] == '_' {
                *pos++
            }
            children = append(children, &WaitNode{pos: Position{Offset: start}})
            
        case c == '(' || c == '[':
            group, err := parseGroup(marble, pos, c == '[')
            if err != nil {
                return nil, err
            }
            children = append(children, group)
            
        case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
            children = append(children, &EventNode{
                Name: string(c),
                pos:  Position{Offset: *pos},
            })
            *pos++
            
        case c == '/':
            node, err := parseNamedEvent(marble, pos)
            if err != nil {
                return nil, err
            }
            children = append(children, node)
            
        default:
            return nil, fmt.Errorf("unexpected character %q at position %d", c, *pos)
        }
    }
    
    return &SequenceNode{Children: children}, nil
}

func parseGroup(marble string, pos *int, ordered bool) (*GroupNode, error) {
    openChar := marble[*pos]
    closeChar := matchingClose(openChar)
    *pos++ // skip open
    
    var children []Node
    
    for *pos < len(marble) {
        c := marble[*pos]
        
        if c == closeChar {
            *pos++ // skip close
            return &GroupNode{
                Children: children,
                Ordered:  ordered,
                pos:      Position{Offset: *pos - 1},
            }, nil
        }
        
        // Parse child node
        node, err := parse(marble, pos)
        if err != nil {
            return nil, err
        }
        if seq, ok := node.(*SequenceNode); ok {
            // Flatten nested sequences
            children = append(children, seq.Children...)
        } else {
            children = append(children, node)
        }
    }
    
    return nil, fmt.Errorf("unclosed %c group at position %d", openChar, *pos)
}

func parseNamedEvent(marble string, pos *int) (*EventNode, error) {
    start := *pos
    *pos++ // skip '/'
    
    var name strings.Builder
    for *pos < len(marble) {
        c := marble[*pos]
        if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
            name.WriteByte(c)
            *pos++
        } else {
            break
        }
    }
    
    if name.Len() == 0 {
        return nil, fmt.Errorf("expected event name after '/' at position %d", start)
    }
    
    return &EventNode{
        Name: name.String(),
        pos:  Position{Offset: start},
    }, nil
}

func matchingClose(open rune) rune {
    switch open {
    case '(': return ')'
    case '[': return ']'
    default: return 0
    }
}

// ParseFollowup is a special case - needs lookahead
func parseWithFollowup(marble string, pos *int) (Node, error) {
    // First parse the event
    start := *pos
    var name string
    
    c := marble[*pos]
    if c == '/' {
        node, err := parseNamedEvent(marble, pos)
        if err != nil {
            return nil, err
        }
        name = node.(*EventNode).Name
    } else if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
        name = string(c)
        *pos++
    } else {
        return nil, fmt.Errorf("expected event at position %d", *pos)
    }
    
    // Check for followup operator
    if *pos+1 < len(marble) && marble[*pos:*pos+2] == "<-<" {
        *pos += 2
        ofEvent, err := parseNamedEvent(marble, pos)
        if err != nil {
            return nil, err
        }
        return &FollowupNode{
            NewEvent: name,
            OfEvent:  ofEvent.Name,
            pos:      Position{Offset: start},
        }, nil
    }
    
    // Just a regular event
    return &EventNode{Name: name, pos: Position{Offset: start}}, nil
}
```

---

## Comparison: Current vs New Approach

### Complexity Comparison

| Aspect | Current Implementation | New Hierarchical AST | Improvement |
|--------|---------------------|---------------------|-------------|
| **Group Representation** | Start/End markers with positions | Composite node with children | ✅ Natural structure |
| **Timeline Building** | ~200 lines, complex state machine | ~80 lines, recursive visitor | ✅ 60% reduction |
| **Interceptor Validation** | ~200 lines, position reconstruction | ~100 lines, tree traversal | ✅ 50% reduction |
| **Parser** | ~160 lines, position tracking | ~200 lines, tree construction | ⚠️ 25% increase (worth it) |
| **Nested Groups** | Manual position adjustment | Automatic via recursion | ✅ Huge simplification |
| **Code Duplication** | High (similar logic in multiple places) | Low (each visitor is unique) | ✅ Eliminated |
| **Extensibility** | Difficult (must update all consumers) | Easy (add new Node types) | ✅ Much better |
| **Debugging** | Hard (position-based errors) | Easy (tree structure visible) | ✅ Significantly better |
| **Testing** | Complex (must set up positions) | Simple (natural tree structure) | ✅ Much easier |

### Code Size Estimation

| Component | Current Lines | New Lines | Change |
|-----------|---------------|-----------|--------|
| `marble/parser.go` | ~160 | ~220 | +60 |
| `marble/op.go` | ~80 | ~120 (new node.go) | +40 |
| `runtime/timeline.go` | ~200 | ~80 | -120 |
| `runtime/interceptor.go` | ~390 | ~150 | -240 |
| **Total** | **~830** | **~570** | **-260** |

**Net result: ~30% code reduction** despite adding the new AST structure.

---

## Migration Strategy

### Phase 1: Dual AST (1-2 days)

Create new AST types without breaking existing code:

```go
// In eventest/internal/marble/node.go (NEW FILE)
package marble

type Node interface {
    Accept(Visitor)
    ToOp() Op  // Convert to old format for compatibility
}

// Add ToNode() method to existing Op types
type Op interface {
    Type() OpType
    ToNode() Node  // NEW: Convert to new format
}

// Implement ToNode() for each Op type
func (o EventOp) ToNode() Node {
    return &EventNode{Name: o.Name}
}

func (o WaitOp) ToNode() Node {
    return &WaitNode{}
}

// etc.
```

Update parser to support both outputs:
```go
// Parse returns []Op (old format)
func Parse(marble string) ([]Op, error) {
    node, err := ParseAsNode(marble)
    if err != nil {
        return nil, err
    }
    return node.ToOpList(), nil
}

// ParseAsNode returns Node (new format)
func ParseAsNode(marble string) (Node, error) {
    // New implementation
}
```

### Phase 2: Gradual Consumer Migration (2-3 days)

1. **Create new Timeline builder** using Visitor pattern
   - Keep old `NewTimeline([]Op)` for compatibility
   - Add new `NewTimelineFromNode(Node)` 

2. **Create new Interceptor validator** using Visitor pattern
   - Keep old validation logic
   - Add new validation using Node

3. **Update Runtime** to use new timeline builder

### Phase 3: Full Cutover (1 day)

1. Update all call sites to use `ParseAsNode()`
2. Remove old Op-based timeline building
3. Remove old Op-based interceptor validation
4. Clean up dual AST code

### Phase 4: Remove Old AST (Optional)

Once all consumers are migrated, the Op interface can be:
- Deprecated
- Removed (breaking change)
- Kept as a thin wrapper

---

## Benefits Summary

### Immediate Benefits

1. **Simplified Timeline Building**
   - No more position tracking
   - Natural recursive structure
   - Easy to understand and maintain

2. **Simplified Interceptor**
   - Direct tree traversal
   - No position reconstruction
   - Clean separation of concerns

3. **Better Error Messages**
   - Position information preserved in nodes
   - Context-aware error reporting

4. **Easier Extension**
   - Add new marble features by adding Node types
   - No need to update all consumers

### Long-Term Benefits

1. **Better Testability**
   - Easy to create test ASTs
   - Each visitor can be tested independently

2. **Improved Performance**
   - Less position calculation overhead
   - More efficient tree traversal

3. **Enhanced Debugging**
   - Tree structure is intuitive
   - Easy to visualize AST

4. **Future-Proof**
   - Supports new language features easily
   - Clean architecture for growth

---

## Implementation Checklist

- [ ] Create `marble/node.go` with Node types
- [ ] Create `marble/visitor.go` with Visitor interface
- [ ] Update `marble/parser.go` to build hierarchical AST
- [ ] Add `ToNode()` and `ToOp()` conversion methods
- [ ] Create `runtime/timeline_builder.go` with Visitor implementation
- [ ] Create `runtime/validator.go` with Visitor implementation
- [ ] Create `runtime/interceptor_validator.go` with Visitor implementation
- [ ] Add comprehensive tests for new AST
- [ ] Update existing tests to use new AST
- [ ] Migrate Timeline construction to use Visitor
- [ ] Migrate Interceptor to use Visitor
- [ ] Remove old position-based code
- [ ] Update documentation

---

## Conclusion

The hierarchical AST with Visitor pattern is the **optimal solution** for the Marble language's AST management problems. It provides:

1. **Massive simplification** of timeline building and interceptor logic
2. **Elimination** of error-prone position tracking
3. **Significant code reduction** (~30% overall)
4. **Better maintainability** and extensibility
5. **Clean architecture** following proven design patterns

The migration is **low-risk** with a dual-AST approach that maintains backward compatibility while enabling gradual adoption of the new structure.

**Recommendation: Start implementation immediately with Phase 1 (Dual AST)**
