# Marble AST Redesign Implementation Plan

## Document Information

- **Project**: it-happened/eventest
- **Feature**: Hierarchical AST with Visitor Pattern for Marble Language
- **Author**: Generated from analysis of eventest package
- **Date**: 2026-06-11
- **Status**: Draft
- **Priority**: High

---

## Executive Summary

This plan outlines the **step-by-step process** to migrate the Marble language implementation from a flat operation list with position markers to a hierarchical AST with the Visitor pattern. This migration will simplify timeline building, runtime execution, and interceptor validation by eliminating complex position tracking.

**Expected Outcome:** improved maintainability, easier extension, better error handling.


---

## Table of Contents

1. [Preparation Phase](#1-preparation-phase)
2. [Phase 1: Foundation - New AST Types](#2-phase-1-foundation---new-ast-types)
3. [Phase 2: Parser Migration](#3-phase-2-parser-migration)
4. [Phase 3: Dual AST Implementation](#4-phase-3-dual-ast-implementation)
5. [Phase 4: Timeline Builder Migration](#5-phase-4-timeline-builder-migration)
6. [Phase 5: Interceptor Migration](#6-phase-5-interceptor-migration)
7. [Phase 6: Runtime Integration](#7-phase-6-runtime-integration)
8. [Phase 7: Validation & Testing](#8-phase-7-validation--testing)
9. [Phase 8: Cleanup & Optimization](#9-phase-8-cleanup--optimization)
10. [Risk Management](#10-risk-management)
11. [Rollback Plan](#11-rollback-plan)
12. [Success Criteria](#12-success-criteria)

---

## 1. Preparation Phase

DONE

**Deliverables:**
- current document
- `docs/marble-new-ast.md` - Technical specification

---

## 2. Phase 1: Foundation - New AST Types

### Objective
Create the new hierarchical AST node types and Visitor interface without modifying existing code.


### Location
`eventest/internal/marble/`

### Tasks

#### 2.1 Create New File: `node.go`
```bash
# File: eventest/internal/marble/node.go
```

**Content:**
```go
package marble

// Position represents a source location for error reporting
type Position struct {
    Line   int
    Column int
    Offset int
}

// Node represents any element in the marble AST
type Node interface {
    // Accept allows the Visitor pattern to traverse the AST
    Accept(Visitor)
    
    // Position provides source location (optional, for error reporting)
    Position() Position
}

// Leaf Nodes (Terminal)

type EventNode struct {
    Name string
    pos  Position
}

func (n *EventNode) Accept(v Visitor) { v.VisitEvent(n) }
func (n *EventNode) Position() Position { return n.pos }

type WaitNode struct {
    pos Position
}

func (n *WaitNode) Accept(v Visitor) { v.VisitWait(n) }
func (n *WaitNode) Position() Position { return n.pos }

type StartNode struct {
    pos Position
}

func (n *StartNode) Accept(v Visitor) { v.VisitStart(n) }
func (n *StartNode) Position() Position { return n.pos }

type FollowupNode struct {
    NewEvent string
    OfEvent  string
    pos      Position
}

func (n *FollowupNode) Accept(v Visitor) { v.VisitFollowup(n) }
func (n *FollowupNode) Position() Position { return n.pos }

// Composite Nodes (Non-Terminal)

type SequenceNode struct {
    Children []Node
    pos      Position
}

func (n *SequenceNode) Accept(v Visitor) { v.VisitSequence(n) }
func (n *SequenceNode) Position() Position { return n.pos }

type GroupNode struct {
    Children []Node
    Ordered  bool
    pos      Position
}

func (n *GroupNode) Accept(v Visitor) { v.VisitGroup(n) }
func (n *GroupNode) Position() Position { return n.pos }
```

#### 2.2 Create New File: `visitor.go`
```bash
# File: eventest/internal/marble/visitor.go
```

**Content:**
```go
package marble

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

// BaseVisitor provides no-op implementations for all Visit methods
type BaseVisitor struct{}

func (v *BaseVisitor) VisitEvent(*EventNode) {}
func (v *BaseVisitor) VisitWait(*WaitNode) {}
func (v *BaseVisitor) VisitStart(*StartNode) {}
func (v *BaseVisitor) VisitFollowup(*FollowupNode) {}
func (v *BaseVisitor) VisitSequence(*SequenceNode) {}
func (v *BaseVisitor) VisitGroup(*GroupNode) {}
```

#### 2.3 Add Conversion Methods to Existing Op Types

**File:** `eventest/internal/marble/op.go`

**Changes:**
```go
// Add ToNode() method to each Op type

func (o EventOp) ToNode() Node {
    return &EventNode{Name: o.Name}
}

func (o WaitOp) ToNode() Node {
    return &WaitNode{}
}

func (o StartEventOp) ToNode() Node {
    return &StartNode{}
}

func (o EventWithFollowupOp) ToNode() Node {
    return &FollowupNode{
        NewEvent: o.EventName,
        OfEvent:  o.From,
    }
}

func (o OrderedGroupStartOp) ToNode() Node {
    // Note: This is a partial conversion - needs context
    // For now, return a placeholder
    return &GroupNode{Ordered: true}
}

func (o OrderedGroupEndOp) ToNode() Node {
    return &GroupNode{Ordered: true}
}

func (o UnorderedGroupStartOp) ToNode() Node {
    return &GroupNode{Ordered: false}
}

func (o UnorderedGroupEndOp) ToNode() Node {
    return &GroupNode{Ordered: false}
}
```

#### 2.4 Create Helper Functions

**File:** `eventest/internal/marble/node_helpers.go`

**Content:**
```go
package marble

// ToOpList converts a Node to []Op for backward compatibility
func ToOpList(node Node) []Op {
    builder := &opListBuilder{}
    node.Accept(builder)
    return builder.ops
}

type opListBuilder struct {
    ops []Op
}

func (b *opListBuilder) VisitEvent(n *EventNode) {
    b.ops = append(b.ops, EventOp{Name: n.Name})
}

func (b *opListBuilder) VisitWait(n *WaitNode) {
    b.ops = append(b.ops, WaitOp{})
}

func (b *opListBuilder) VisitStart(n *StartNode) {
    b.ops = append(b.ops, StartEventOp{})
}

func (b *opListBuilder) VisitFollowup(n *FollowupNode) {
    b.ops = append(b.ops, EventWithFollowupOp{
        EventName: n.NewEvent,
        From:      n.OfEvent,
    })
}

func (b *opListBuilder) VisitSequence(n *SequenceNode) {
    for _, child := range n.Children {
        child.Accept(b)
    }
}

func (b *opListBuilder) VisitGroup(n *GroupNode) {
    // Add start marker
    if n.Ordered {
        b.ops = append(b.ops, OrderedGroupStartOp{})
    } else {
        b.ops = append(b.ops, UnorderedGroupStartOp{})
    }
    
    // Add children
    for _, child := range n.Children {
        child.Accept(b)
    }
    
    // Add end marker
    if n.Ordered {
        b.ops = append(b.ops, OrderedGroupEndOp{})
    } else {
        b.ops = append(b.ops, UnorderedGroupEndOp{})
    }
}

// String returns a string representation of a Node (for debugging)
func String(node Node) string {
    builder := &stringBuilder{}
    node.Accept(builder)
    return builder.String()
}

type stringBuilder struct {
    strings.Builder
}

func (b *stringBuilder) VisitEvent(n *EventNode) {
    b.WriteString(n.Name)
}

func (b *stringBuilder) VisitWait(n *WaitNode) {
    b.WriteRune('-')
}

func (b *stringBuilder) VisitStart(n *StartNode) {
    b.WriteRune('^')
}

func (b *stringBuilder) VisitFollowup(n *FollowupNode) {
    b.WriteString(fmt.Sprintf("%s<-s", n.NewEvent, n.OfEvent))
}

func (b *stringBuilder) VisitSequence(n *SequenceNode) {
    for i, child := range n.Children {
        child.Accept(b)
        if i < len(n.Children)-1 {
            // Add separator if needed
        }
    }
}

func (b *stringBuilder) VisitGroup(n *GroupNode) {
    open := '['
    close := ']'
    if !n.Ordered {
        open = '('
        close = ')'
    }
    b.WriteRune(open)
    for _, child := range n.Children {
        child.Accept(b)
    }
    b.WriteRune(close)
}
```

#### 2.5 Add Tests for New AST Types

**File:** `eventest/internal/marble/node_test.go`

**Content:**
```go
package marble_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestNodeCreation(t *testing.T) {
    t.Run("should create EventNode", func(t *testing.T) {
        node := &marble.EventNode{Name: "test"}
        assert.NotNil(t, node)
        assert.Equal(t, "test", node.Name)
    })
    
    t.Run("should create GroupNode", func(t *testing.T) {
        node := &marble.GroupNode{
            Ordered: true,
            Children: []marble.Node{
                &marble.EventNode{Name: "a"},
                &marble.EventNode{Name: "b"},
            },
        }
        assert.NotNil(t, node)
        assert.True(t, node.Ordered)
        assert.Len(t, node.Children, 2)
    })
}

func TestNodeToOpConversion(t *testing.T) {
    t.Run("should convert EventNode to EventOp", func(t *testing.T) {
        node := &marble.EventNode{Name: "test"}
        op := node.ToNode().(*marble.EventNode)
        assert.Equal(t, "test", op.Name)
    })
}

func TestOpToNodeConversion(t *testing.T) {
    t.Run("should convert EventOp to EventNode", func(t *testing.T) {
        op := marble.EventOp{Name: "test"}
        node := op.ToNode()
        assert.IsType(t, &marble.EventNode{}, node)
    })
}

func TestToOpList(t *testing.T) {
    t.Run("should convert SequenceNode to Op list", func(t *testing.T) {
        node := &marble.SequenceNode{
            Children: []marble.Node{
                &marble.EventNode{Name: "a"},
                &marble.WaitNode{},
                &marble.EventNode{Name: "b"},
            },
        }
        ops := marble.ToOpList(node)
        assert.Len(t, ops, 3)
        assert.Equal(t, marble.EventOp{Name: "a"}, ops[0])
        assert.Equal(t, marble.WaitOp{}, ops[1])
        assert.Equal(t, marble.EventOp{Name: "b"}, ops[2])
    })
    
    t.Run("should convert GroupNode to Op list", func(t *testing.T) {
        node := &marble.GroupNode{
            Ordered: true,
            Children: []marble.Node{
                &marble.EventNode{Name: "a"},
                &marble.EventNode{Name: "b"},
            },
        }
        ops := marble.ToOpList(node)
        assert.Len(t, ops, 4)
        assert.Equal(t, marble.OrderedGroupStartOp{}, ops[0])
        assert.Equal(t, marble.EventOp{Name: "a"}, ops[1])
        assert.Equal(t, marble.EventOp{Name: "b"}, ops[2])
        assert.Equal(t, marble.OrderedGroupEndOp{}, ops[3])
    })
}
```

**Verifications:**
- [ ] All new types compile without errors
- [ ] All tests pass
- [ ] No breaking changes to existing code

---

## 3. Phase 2: Parser Migration

### Objective
Update the parser to build hierarchical AST while maintaining backward compatibility.

### Duration
- **Estimated**: 1 day

### Location
`eventest/internal/marble/parser.go`

### Tasks

#### 3.1 Update Parser to Build Hierarchical AST

**Changes to `parser.go`:**

```go
// Add new export function
func ParseAsNode(marble string) (Node, error) {
    var pos int
    return parse(marble, &pos)
}

// Modify existing Parse to use new parser
func Parse(marble string) ([]Op, error) {
    node, err := ParseAsNode(marble)
    if err != nil {
        return nil, err
    }
    return ToOpList(node), nil
}

// Update parse function to return Node
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
                return nil, errors.Join(
                    ErrMarbleSyntax,
                    fmt.Errorf("unexpected ^ at %d", *pos),
                )
            }
            children = append(children, &StartNode{pos: Position{Offset: *pos}})
            *pos++
            
        case c == '-':
            children = append(children, &WaitNode{pos: Position{Offset: *pos}})
            *pos++
            
        case c == '_':
            start := *pos
            for *pos < len(marble) && marble[*pos] == '_' {
                *pos++
            }
            children = append(children, &WaitNode{pos: Position{Offset: start}})
            
        case c == '(', c == '[':
            ordered := c == '['
            group, err := parseGroup(marble, pos, ordered)
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
            node, err := parseLabel(marble, pos)
            if err != nil {
                return nil, err
            }
            // Check for followup
            if *pos+1 < len(marble) && marble[*pos:*pos+2] == "<-<" {
                *pos += 2
                ofNode, err := parseLabel(marble, pos)
                if err != nil {
                    return nil, err
                }
                children = append(children, &FollowupNode{
                    NewEvent: node.(*EventNode).Name,
                    OfEvent:  ofNode.(*EventNode).Name,
                    pos:      Position{Offset: *pos},
                })
            } else {
                children = append(children, node)
            }
            
        default:
            return nil, errors.Join(
                ErrMarbleSyntax,
                fmt.Errorf("unexpected character %q at %d", c, *pos),
            )
        }
    }
    
    if len(children) == 0 {
        return nil, ErrEmptyMarble
    }
    
    return &SequenceNode{Children: children}, nil
}

// New helper for parsing groups
func parseGroup(marble string, pos *int, ordered bool) (*GroupNode, error) {
    openChar := marble[*pos]
    closeChar := matchingClose(openChar)
    startPos := *pos
    *pos++ // skip open
    
    var children []Node
    
    for *pos < len(marble) {
        c := marble[*pos]
        
        if c == closeChar {
            endPos := *pos
            *pos++ // skip close
            return &GroupNode{
                Children: children,
                Ordered:  ordered,
                pos:      Position{Offset: startPos},
            }, nil
        }
        
        // Parse child node
        node, err := parse(marble, pos)
        if err != nil {
            return nil, err
        }
        children = append(children, node)
    }
    
    return nil, errors.Join(
        ErrMarbleSyntax,
        fmt.Errorf("unclosed %c at %d", openChar, startPos),
    )
}

// Update parseLabel to return Node
func parseLabel(marble string, i *int) (Node, error) {
    c := marble[*i]
    var label string
    
    if c == '/' {
        lb := strings.Builder{}
        *i++
        for *i < len(marble) {
            c = marble[*i]
            if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
                lb.WriteByte(marble[*i])
                *i++
            } else {
                break
            }
        }
        label = lb.String()
    } else {
        label = string(c)
        *i++
    }
    
    return &EventNode{Name: label, pos: Position{Offset: *i - len(label)}}, nil
}

func matchingClose(open rune) rune {
    switch open {
    case '(': return ')'
    case '[': return ']'
    default: return 0
    }
}
```

#### 3.2 Update Parser Tests

**File:** `eventest/internal/marble/parser_test.go`

**Changes:**
- Add tests for `ParseAsNode`
- Verify AST structure for various marble strings
- Ensure backward compatibility with existing `Parse` tests

```go
func TestParseAsNode(t *testing.T) {
    t.Run("should parse simple sequence", func(t *testing.T) {
        node, err := marble.ParseAsNode("abc")
        assert.NoError(t, err)
        
        seq, ok := node.(*marble.SequenceNode)
        assert.True(t, ok)
        assert.Len(t, seq.Children, 3)
        
        assert.IsType(t, &marble.EventNode{}, seq.Children[0])
        assert.Equal(t, "a", seq.Children[0].(*marble.EventNode).Name)
    })
    
    t.Run("should parse ordered group", func(t *testing.T) {
        node, err := marble.ParseAsNode("[ab]")
        assert.NoError(t, err)
        
        seq, ok := node.(*marble.SequenceNode)
        assert.True(t, ok)
        assert.Len(t, seq.Children, 1)
        
        group, ok := seq.Children[0].(*marble.GroupNode)
        assert.True(t, ok)
        assert.True(t, group.Ordered)
        assert.Len(t, group.Children, 2)
    })
    
    t.Run("should parse nested groups", func(t *testing.T) {
        node, err := marble.ParseAsNode("[a(bc)d]")
        assert.NoError(t, err)
        
        seq := node.(*marble.SequenceNode)
        group := seq.Children[0].(*marble.GroupNode)
        
        assert.Len(t, group.Children, 3)
        innerGroup := group.Children[1].(*marble.GroupNode)
        assert.False(t, innerGroup.Ordered)
    })
    
    t.Run("should parse followup events", func(t *testing.T) {
        node, err := marble.ParseAsNode("a<-b")
        assert.NoError(t, err)
        
        seq := node.(*marble.SequenceNode)
        followup := seq.Children[0].(*marble.FollowupNode)
        
        assert.Equal(t, "a", followup.NewEvent)
        assert.Equal(t, "b", followup.OfEvent)
    })
}

func TestParseBackwardCompatibility(t *testing.T) {
    t.Run("Parse should still work", func(t *testing.T) {
        ops, err := marble.Parse("abc")
        assert.NoError(t, err)
        assert.Len(t, ops, 3)
    })
    
    t.Run("Parse and ParseAsNode should produce equivalent results", func(t *testing.T) {
        marbleStr := "a-[bc]-(de)"
        
        ops, err := marble.Parse(marbleStr)
        assert.NoError(t, err)
        
        node, err := marble.ParseAsNode(marbleStr)
        assert.NoError(t, err)
        
        convertedOps := marble.ToOpList(node)
        assert.Equal(t, ops, convertedOps)
    })
}
```

**Verifications:**
- [ ] All existing parser tests still pass
- [ ] New `ParseAsNode` tests pass
- [ ] Backward compatibility maintained

---

## 4. Phase 3: Dual AST Implementation

### Objective
Ensure both AST representations (Op and Node) work simultaneously, allowing gradual migration of consumers.

### Duration
- **Estimated**: 0.5 day

### Tasks

#### 4.1 Add Convenience Functions

**File:** `eventest/internal/marble/conversions.go`

```go
package marble

// ParseAndValidateAsNode parses and validates a marble string, returning a Node
func ParseAndValidateAsNode(marble string, rules ...Rule) (Node, error) {
    node, err := ParseAsNode(marble)
    if err != nil {
        return nil, err
    }
    
    // Convert to Op for validation (using existing validation)
    ops := ToOpList(node)
    if err := Validate(ops, rules...); err != nil {
        return nil, err
    }
    
    return node, nil
}

// ValidateNode validates a Node using the existing Rule system
func ValidateNode(node Node, rules ...Rule) error {
    ops := ToOpList(node)
    return Validate(ops, rules...)
}
```

#### 4.2 Update Runtime to Accept Both Formats

**File:** `eventest/internal/runtime/runtime.go`

**Changes:**
```go
// Add new method that accepts Node
func (r *Runtime) RunFromNode(marbleNode marble.Node) (*RunningSession, error) {
    ops := marble.ToOpList(marbleNode)
    return r.RunFromOps(ops)
}

// Rename existing Run to RunFromOps for clarity
func (r *Runtime) RunFromOps(marbleSeq string) (*RunningSession, error) {
    ops, err := marble.Parse(marbleSeq)
    if err != nil {
        return nil, err
    }
    
    if err := marble.Validate(ops,
        marble.StartEventAnywhereRule{},
        marble.WaitlessGroupsRule{},
    ); err != nil {
        return nil, err
    }
    
    tl := NewTimeline(ops, TimelineWithTickDuration(r.baseTickDuration))
    ticks := tl.Ticks()
    
    return &RunningSession{
        rt:         r,
        ticks:      ticks,
        clock:      r.clock,
        bus:        r.bus,
        payloadMap: r.payloadMap,
        eventMap:   r.eventMap,
    }, nil
}

// Update RunAll to use RunFromOps
func (r *Runtime) RunAll(marbleSeq string) error {
    sess, err := r.RunFromOps(marbleSeq)
    if err != nil {
        return err
    }
    
    for sess.HasNext() {
        if err := sess.Next(); err != nil {
            if errors.Is(err, SessionEnded) {
                err = nil
            }
            return err
        }
    }
    
    return nil
}
```

**Verifications:**
- [ ] Both `Run` and `RunFromNode` work correctly
- [ ] No breaking changes to existing API

---

## 5. Phase 4: Timeline Builder Migration

### Objective
Rewrite the Timeline builder to use the Visitor pattern with the new hierarchical AST.

### Duration
- **Estimated**: 1 day

### Location
`eventest/internal/runtime/`

### Tasks

#### 5.1 Create New Timeline Builder

**File:** `eventest/internal/runtime/timeline_builder.go` (NEW)

```go
package runtime

import (
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

// TimelineBuilder builds a timeline from a hierarchical AST node
type TimelineBuilder struct {
    ticks   []Tick
    current []marble.Op
    config  TimelineBuilderConfig
}

type TimelineBuilderConfig struct {
    TickDuration time.Duration
}

func NewTimelineBuilder(config TimelineBuilderConfig) *TimelineBuilder {
    return &TimelineBuilder{
        ticks:  make([]Tick, 0),
        config: config,
    }
}

// Build constructs a timeline from the AST root node
func (b *TimelineBuilder) Build(root marble.Node) ([]Tick, error) {
    root.Accept(b)
    return b.ticks, nil
}

// Implement Visitor interface
func (b *TimelineBuilder) VisitEvent(n *marble.EventNode) {
    if b.current == nil {
        // Top-level event = new tick
        b.ticks = append(b.ticks, Tick{
            Duration: b.config.TickDuration,
            Ops:     []marble.Op{marble.EventOp{Name: n.Name}},
        })
    } else {
        // Inside group = add to current tick
        b.current = append(b.current, marble.EventOp{Name: n.Name})
    }
}

func (b *TimelineBuilder) VisitWait(n *marble.WaitNode) {
    if b.current == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: b.config.TickDuration,
            Ops:     []marble.Op{marble.WaitOp{}},
        })
    } else {
        b.current = append(b.current, marble.WaitOp{})
    }
}

func (b *TimelineBuilder) VisitStart(n *marble.StartNode) {
    if b.current == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: b.config.TickDuration,
            Ops:     []marble.Op{marble.StartEventOp{}},
        })
    } else {
        b.current = append(b.current, marble.StartEventOp{})
    }
}

func (b *TimelineBuilder) VisitFollowup(n *marble.FollowupNode) {
    op := marble.EventWithFollowupOp{
        EventName: n.NewEvent,
        From:      n.OfEvent,
    }
    if b.current == nil {
        b.ticks = append(b.ticks, Tick{
            Duration: b.config.TickDuration,
            Ops:     []marble.Op{op},
        })
    } else {
        b.current = append(b.current, op)
    }
}

func (b *TimelineBuilder) VisitSequence(n *marble.SequenceNode) {
    // Process all children
    for _, child := range n.Children {
        child.Accept(b)
    }
}

func (b *TimelineBuilder) VisitGroup(n *marble.GroupNode) {
    // Create group start marker
    var startOp marble.Op
    if n.Ordered {
        startOp = marble.OrderedGroupStartOp{}
    } else {
        startOp = marble.UnorderedGroupStartOp{}
    }
    
    // Create group end marker
    var endOp marble.Op
    if n.Ordered {
        endOp = marble.OrderedGroupEndOp{}
    } else {
        endOp = marble.UnorderedGroupEndOp{}
    }
    
    // Save current state
    oldCurrent := b.current
    
    // Start new tick for this group
    b.current = []marble.Op{startOp}
    
    // Process all children (they add to current tick)
    for _, child := range n.Children {
        child.Accept(b)
    }
    
    // Finalize the tick
    b.current = append(b.current, endOp)
    b.ticks = append(b.ticks, Tick{
        Duration: b.config.TickDuration,
        Ops:     b.current,
    })
    
    // Restore previous state
    b.current = oldCurrent
}
```

#### 5.2 Create Timeline from Node

**File:** `eventest/internal/runtime/timeline.go` (UPDATE)

**Changes:**
```go
// Add new constructor
func NewTimelineFromNode(root marble.Node, opts ...TimelineOption) *Timeline {
    if err := marble.ValidateNode(root, marble.WaitlessGroupsRule{}); err != nil {
        panic(err)
    }
    
    t := &Timeline{
        events: make(map[string]event.Event),
        randGen: rand.New(
            rand.NewPCG(
                uint64(time.Now().UnixNano()), uint64(time.Now().UnixMilli()))),
        tickDuration: DefaultTickDuration,
    }
    
    for _, opt := range opts {
        opt(t)
    }
    
    // Use new builder
    builder := NewTimelineBuilder(TimelineBuilderConfig{
        TickDuration: t.tickDuration,
    })
    
    ticks, err := builder.Build(root)
    if err != nil {
        panic(err)
    }
    
    t.ticks = ticks
    return t
}

// Keep existing NewTimeline for backward compatibility
func NewTimeline(rowOps []marble.Op, opts ...TimelineOption) *Timeline {
    // Convert to node and use new implementation
    // Or keep old implementation temporarily
    node := marble.SequenceNodeFromOps(rowOps)
    return NewTimelineFromNode(&node, opts...)
}
```

#### 5.3 Add Helper to Convert Ops to Node

**File:** `eventest/internal/marble/conversions.go` (UPDATE)

```go
// SequenceNodeFromOps converts []Op to SequenceNode
func SequenceNodeFromOps(ops []Op) SequenceNode {
    var children []Node
    for _, op := range ops {
        children = append(children, op.ToNode())
    }
    return SequenceNode{Children: children}
}
```

#### 5.4 Update Timeline Tests

**File:** `eventest/internal/runtime/timeline_test.go`

**Changes:**
- Add tests for `NewTimelineFromNode`
- Verify equivalence with existing `NewTimeline`
- Test complex nested group scenarios

```go
func TestNewTimelineFromNode(t *testing.T) {
    t.Run("should build timeline from simple sequence node", func(t *testing.T) {
        node, err := marble.ParseAsNode("abc")
        assert.NoError(t, err)
        
        tl := runtime.NewTimelineFromNode(node)
        ticks := tl.Ticks()
        
        assert.Len(t, ticks, 3)
        assert.Equal(t, runtime.DefaultTickDuration, ticks[0].Duration)
    })
    
    t.Run("should build timeline from group node", func(t *testing.T) {
        node, err := marble.ParseAsNode("[ab]")
        assert.NoError(t, err)
        
        tl := runtime.NewTimelineFromNode(node)
        ticks := tl.Ticks()
        
        assert.Len(t, ticks, 1)
        assert.Len(t, ticks[0].Ops, 4) // start, a, b, end
    })
    
    t.Run("should produce same result as NewTimeline", func(t *testing.T) {
        marbleStr := "a-[bc]-(de)"
        
        // Old way
        ops, _ := marble.Parse(marbleStr)
        tl1 := runtime.NewTimeline(ops)
        
        // New way
        node, _ := marble.ParseAsNode(marbleStr)
        tl2 := runtime.NewTimelineFromNode(node)
        
        assert.Equal(t, tl1.Ticks(), tl2.Ticks())
    })
}
```

**Verifications:**
- [ ] New timeline builder works correctly
- [ ] Backward compatibility maintained
- [ ] All timeline tests pass

---

## 6. Phase 5: Interceptor Migration

### Objective
Rewrite the Interceptor validation logic to use the Visitor pattern with the new hierarchical AST.

### Duration
- **Estimated**: 1 day

### Location
`eventest/internal/runtime/interceptor.go`

### Tasks

#### 6.1 Create Interceptor Validator

**File:** `eventest/internal/runtime/interceptor_validator.go` (NEW)

```go
package runtime

import (
    "fmt"
    "time"
    
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

// InterceptorValidator validates actual events against expected AST
type InterceptorValidator struct {
    timeline    *Timeline
    activity    []activityEntry
    matchers    map[string]event.Matcher
    errors      []error
    currentTick int
    config      InterceptorConfig
}

type InterceptorConfig struct {
    Strict     bool
    FailFast   bool
}

func NewInterceptorValidator(
    timeline *Timeline,
    activity []activityEntry,
    matchers map[string]event.Matcher,
    config InterceptorConfig,
) *InterceptorValidator {
    return &InterceptorValidator{
        timeline: timeline,
        activity: activity,
        matchers: matchers,
        config:   config,
    }
}

func (v *InterceptorValidator) Validate(root marble.Node) []error {
    root.Accept(v)
    return v.errors
}

// Implement Visitor interface
func (v *InterceptorValidator) VisitSequence(n *marble.SequenceNode) {
    for i, child := range n.Children {
        v.currentTick = i
        child.Accept(v)
    }
}

func (v *InterceptorValidator) VisitGroup(n *marble.GroupNode) {
    if v.currentTick >= len(v.timeline.Ticks()) {
        v.errors = append(v.errors,
            fmt.Errorf("group at tick %d: no corresponding tick", v.currentTick))
        return
    }
    
    tick := v.timeline.Ticks()[v.currentTick]
    tickActivity := selectActivityEntriesByRange(
        v.activity,
        time.Duration(v.currentTick)*tick.Duration,
        time.Duration(v.currentTick+1)*tick.Duration)
    
    if n.Ordered {
        v.validateOrderedGroup(n, tickActivity)
    } else {
        v.validateUnorderedGroup(n, tickActivity)
    }
}

func (v *InterceptorValidator) validateOrderedGroup(
    n *marble.GroupNode,
    activity []activityEntry,
) {
    if len(n.Children) != len(activity) {
        if v.config.Strict {
            v.errors = append(v.errors,
                fmt.Errorf("ordered group at tick %d: expected %d events, got %d",
                    v.currentTick, len(n.Children), len(activity)))
        }
        return
    }
    
    for i, child := range n.Children {
        if i >= len(activity) {
            break
        }
        
        switch c := child.(type) {
        case *marble.EventNode:
            if m := v.matchers[c.Name]; !m.Match(activity[i].event) {
                v.errors = append(v.errors,
                    fmt.Errorf("tick %d, event %d: expected %s matching %v, got %v",
                        v.currentTick, i, c.Name, m, activity[i].event))
            }
        case *marble.FollowupNode:
            if m := v.matchers[c.NewEvent]; !m.Match(activity[i].event) {
                v.errors = append(v.errors,
                    fmt.Errorf("tick %d, event %d: expected %s matching %v, got %v",
                        v.currentTick, i, c.NewEvent, m, activity[i].event))
            }
        }
    }
}

func (v *InterceptorValidator) validateUnorderedGroup(
    n *marble.GroupNode,
    activity []activityEntry,
) {
    if len(n.Children) != len(activity) {
        if v.config.Strict {
            v.errors = append(v.errors,
                fmt.Errorf("unordered group at tick %d: expected %d events, got %d",
                    v.currentTick, len(n.Children), len(activity)))
        }
        return
    }
    
    // Create map of expected matchers
    expected := make(map[string]event.Matcher)
    expectedCount := make(map[string]int)
    
    for _, child := range n.Children {
        switch c := child.(type) {
        case *marble.EventNode:
            expected[c.Name] = v.matchers[c.Name]
            expectedCount[c.Name]++
        case *marble.FollowupNode:
            expected[c.NewEvent] = v.matchers[c.NewEvent]
            expectedCount[c.NewEvent]++
        }
    }
    
    // Match actual events to expected
    matched := make(map[string]int)
    
    for _, act := range activity {
        found := false
        for name, matcher := range expected {
            if matched[name] < expectedCount[name] && matcher.Match(act.event) {
                matched[name]++
                found = true
                break
            }
        }
        if !found && v.config.Strict {
            v.errors = append(v.errors,
                fmt.Errorf("tick %d: unexpected event in unordered group: %v",
                    v.currentTick, act.event))
        }
    }
}

func (v *InterceptorValidator) VisitEvent(n *marble.EventNode) {
    // Top-level events handled by sequence
}

func (v *InterceptorValidator) VisitWait(n *marble.WaitNode) {
    // Wait nodes handled by sequence
}

func (v *InterceptorValidator) VisitStart(n *marble.StartNode) {
    // Start node handled by sequence
}

func (v *InterceptorValidator) VisitFollowup(n *marble.FollowupNode) {
    // Followup nodes handled by group validation
}
```

#### 6.2 Update InterceptorRecorder

**File:** `eventest/internal/runtime/interceptor.go` (UPDATE)

**Changes:**
```go
// Update FromMarble to use new AST
func (r *InterceptorRecorder) FromMarble(seq string) *InterceptorRecorder {
    if r.expectedSeq != "" {
        panic("already expecting a marble sequence")
    }
    r.expectedSeq = seq
    
    // Parse as node
    node, err := marble.ParseAsNode(seq)
    if err != nil {
        panic(err)
    }
    
    // Validate
    if err := marble.ValidateNode(node, marble.WaitlessGroupsRule{}); err != nil {
        panic(err)
    }
    
    // Build timeline from node
    r.it.expectedOps = marble.ToOpList(node)
    r.timeline = NewTimelineFromNode(node)
    
    // Set up matchers
    for _, tick := range r.timeline.Ticks() {
        for _, op := range tick.Ops {
            switch o := op.(type) {
            case marble.EventOp:
                if _, ok := r.matchers[o.Name]; !ok {
                    r.matchers[o.Name] = event.HasPayload(DefaultPayload(o.Name))
                }
            case marble.EventWithFollowupOp:
                if _, ok := r.matchers[o.EventName]; !ok {
                    r.matchers[o.EventName] = event.HasPayload(DefaultPayload(o.EventName))
                }
            }
        }
    }
    
    return r
}

// Update Failures to use new validator
func (r *InterceptorRecorder) Failures() []error {
    if r.it.clock.Started() {
        return []error{fmt.Errorf("clock has not been stopped")}
    }
    
    // Parse expected sequence as node
    node, err := marble.ParseAsNode(r.expectedSeq)
    if err != nil {
        return []error{err}
    }
    
    // Use new validator
    validator := NewInterceptorValidator(
        r.timeline,
        r.it.actualActivityEntries,
        r.matchers,
        InterceptorConfig{Strict: true, FailFast: false},
    )
    
    return validator.Validate(node)
}
```

#### 6.3 Remove Dead Code

- [ ] Delete commented out code (lines 298-325, 375-391)
- [ ] Remove unused utility functions
- [ ] Remove incomplete implementations

**Verifications:**
- [ ] Interceptor works with new validation
- [ ] All interceptor tests pass
- [ ] Backward compatibility maintained

---

## 7. Phase 6: Runtime Integration

### Objective
Fully integrate the new AST into the runtime, ensuring all components work together.

### Duration
- **Estimated**: 0.5 day

### Location
`eventest/internal/runtime/`

### Tasks

#### 7.1 Update Runtime.Run to Use New Parser

**File:** `eventest/internal/runtime/runtime.go` (UPDATE)

```go
// Update to use ParseAsNode
func (r *Runtime) Run(marbleSeq string) (*RunningSession, error) {
    // Parse as node
    node, err := marble.ParseAndValidateAsNode(marbleSeq,
        marble.StartEventAnywhereRule{},
        marble.WaitlessGroupsRule{},
    )
    if err != nil {
        return nil, err
    }
    
    // Build timeline from node
    tl := NewTimelineFromNode(node, TimelineWithTickDuration(r.baseTickDuration))
    ticks := tl.Ticks()
    
    return &RunningSession{
        rt:         r,
        ticks:      ticks,
        clock:      r.clock,
        bus:        r.bus,
        payloadMap: r.payloadMap,
        eventMap:   r.eventMap,
    }, nil
}
```

#### 7.2 Update Harness

**File:** `eventest/harness.go` (UPDATE)

```go
func NewHarness(bus event.Bus, expected string, opts ...Option) *Harness {
    h := &Harness{
        bus:      bus,
        expected: expected,
    }
    
    for _, opt := range opts {
        opt(h)
    }
    
    return h
}

func (h *Harness) Run(t *testing.T, f func()) {
    // Create runtime
    rt := runtime.NewRuntime(h.bus)
    
    // Parse expected sequence
    node, err := marble.ParseAsNode(h.expected)
    if err != nil {
        t.Fatalf("Failed to parse expected marble: %v", err)
    }
    
    // Validate
    if err := marble.ValidateNode(node,
        marble.StartEventAnywhereRule{},
        marble.WaitlessGroupsRule{},
    ); err != nil {
        t.Fatalf("Invalid expected marble: %v", err)
    }
    
    // Create interceptor
    intercept := runtime.NewInterceptor(t, h.bus, rt.Clock())
    recorder := intercept.EXPECT().FromMarble(h.expected)
    
    // Set up matchers if provided
    if h.matchers != nil {
        recorder.ShouldMatch(h.matchers)
    }
    if h.payloads != nil {
        // Set up payloads
    }
    
    // Execute test function
    f()
    
    // Finalize
    intercept.Finish()
}
```

#### 7.3 Complete Option Implementations

**File:** `eventest/harness.go` (UPDATE)

```go
func WithPayloads(payloads map[string]event.Payload) Option {
    return func(h *Harness) {
        if h.payloads == nil {
            h.payloads = make(map[string]event.Payload)
        }
        for k, v := range payloads {
            h.payloads[k] = v
        }
    }
}

func WithMatchers(matchers map[string]event.Matcher) Option {
    return func(h *Harness) {
        if h.matchers == nil {
            h.matchers = make(map[string]event.Matcher)
        }
        for k, v := range matchers {
            h.matchers[k] = v
        }
    }
}

func WithSideEffect(marble string) Option {
    return func(h *Harness) {
        if h.sideEffects == nil {
            h.sideEffects = make([]string, 0)
        }
        h.sideEffects = append(h.sideEffects, marble)
    }
}
```

**Verifications:**
- [ ] Runtime fully integrated with new AST
- [ ] Harness works correctly
- [ ] All options are functional

---

## 8. Phase 7: Validation & Testing

### Objective
Ensure all changes work correctly and maintain backward compatibility.

### Duration
- **Estimated**: 1 day

### Tasks

#### 8.1 Run All Existing Tests

```bash
# Run all tests in the eventest package
cd /home/thomas/Documents/projects/opensource/it-happened
go test ./eventest/... -v

# Run marble tests
go test ./eventest/internal/marble/... -v

# Run runtime tests
go test ./eventest/internal/runtime/... -v
```

**Expected:** All tests pass

#### 8.2 Add Integration Tests

**File:** `eventest/internal/marble/integration_test.go` (NEW)

```go
package marble_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestFullRoundTrip(t *testing.T) {
    testCases := []string{
        "abc",
        "a-b-c",
        "[ab]",
        "(ab)",
        "a-[bc]-(de)",
        "^a-(bc)[de]f",
        "a<-b",
        "/eventName",
        "a /event b",
    }
    
    for _, tc := range testCases {
        t.Run(tc, func(t *testing.T) {
            // Parse with old method
            ops, err := marble.Parse(tc)
            assert.NoError(t, err)
            
            // Parse with new method
            node, err := marble.ParseAsNode(tc)
            assert.NoError(t, err)
            
            // Convert node to ops
            convertedOps := marble.ToOpList(node)
            
            // Should be equal
            assert.Equal(t, ops, convertedOps)
            
            // Convert back to node
            node2 := marble.SequenceNodeFromOps(convertedOps)
            
            // Parse again
            ops2 := marble.ToOpList(&node2)
            
            // Should still be equal
            assert.Equal(t, ops, ops2)
        })
    }
}

func TestASTStructure(t *testing.T) {
    t.Run("should create correct hierarchy for nested groups", func(t *testing.T) {
        node, _ := marble.ParseAsNode("[a(b[c]d)e]")
        
        seq := node.(*marble.SequenceNode)
        assert.Len(t, seq.Children, 1)
        
        group1 := seq.Children[0].(*marble.GroupNode)
        assert.True(t, group1.Ordered)
        assert.Len(t, group1.Children, 3)
        
        // a
        assert.IsType(t, &marble.EventNode{}, group1.Children[0])
        
        // (b[c]d)
        group2 := group1.Children[1].(*marble.GroupNode)
        assert.False(t, group2.Ordered)
        assert.Len(t, group2.Children, 2)
        
        // b
        assert.IsType(t, &marble.EventNode{}, group2.Children[0])
        
        // [c]
        group3 := group2.Children[1].(*marble.GroupNode)
        assert.True(t, group3.Ordered)
        
        // e
        assert.IsType(t, &marble.EventNode{}, group1.Children[2])
    })
}
```

#### 8.3 Test Timeline Builder

**File:** `eventest/internal/runtime/timeline_builder_test.go` (NEW)

```go
package runtime_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
    "github.com/thomas-marquis/it-happened/eventest/internal/runtime"
)

func TestTimelineBuilder_BuildsCorrectTicks(t *testing.T) {
    t.Run("simple sequence", func(t *testing.T) {
        node, _ := marble.ParseAsNode("abc")
        builder := runtime.NewTimelineBuilder(runtime.TimelineBuilderConfig{})
        ticks, err := builder.Build(node)
        
        assert.NoError(t, err)
        assert.Len(t, ticks, 3)
        assert.Len(t, ticks[0].Ops, 1)
        assert.Len(t, ticks[1].Ops, 1)
        assert.Len(t, ticks[2].Ops, 1)
    })
    
    t.Run("ordered group", func(t *testing.T) {
        node, _ := marble.ParseAsNode("[abc]")
        builder := runtime.NewTimelineBuilder(runtime.TimelineBuilderConfig{})
        ticks, err := builder.Build(node)
        
        assert.NoError(t, err)
        assert.Len(t, ticks, 1)
        assert.Len(t, ticks[0].Ops, 5) // start, a, b, c, end
    })
    
    t.Run("mixed sequence and groups", func(t *testing.T) {
        node, _ := marble.ParseAsNode("a-[bc]-(de)")
        builder := runtime.NewTimelineBuilder(runtime.TimelineBuilderConfig{})
        ticks, err := builder.Build(node)
        
        assert.NoError(t, err)
        assert.Len(t, ticks, 4)
    })
}
```

#### 8.4 Test Interceptor Validator

**File:** `eventest/internal/runtime/interceptor_validator_test.go` (NEW)

```go
package runtime_test

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/thomas-marquis/it-happened/event"
    "github.com/thomas-marquis/it-happened/eventest/internal/marble"
    "github.com/thomas-marquis/it-happened/eventest/internal/runtime"
)

func TestInterceptorValidator_ValidatesCorrectly(t *testing.T) {
    t.Run("simple sequence", func(t *testing.T) {
        node, _ := marble.ParseAsNode("abc")
        tl := runtime.NewTimelineFromNode(node)
        
        // Create mock activity
        activity := []runtime.ActivityEntry{
            {event: event.New(marble.DefaultPayload("a")), elapsedFromStart: 0},
            {event: event.New(marble.DefaultPayload("b")), elapsedFromStart: 10 * time.Millisecond},
            {event: event.New(marble.DefaultPayload("c")), elapsedFromStart: 20 * time.Millisecond},
        }
        
        matchers := map[string]event.Matcher{
            "a": event.HasPayload(marble.DefaultPayload("a")),
            "b": event.HasPayload(marble.DefaultPayload("b")),
            "c": event.HasPayload(marble.DefaultPayload("c")),
        }
        
        validator := runtime.NewInterceptorValidator(
            tl,
            activity,
            matchers,
            runtime.InterceptorConfig{Strict: true},
        )
        
        errors := validator.Validate(node)
        assert.Empty(t, errors)
    })
    
    t.Run("detects missing event", func(t *testing.T) {
        node, _ := marble.ParseAsNode("abc")
        tl := runtime.NewTimelineFromNode(node)
        
        // Missing event "b"
        activity := []runtime.ActivityEntry{
            {event: event.New(marble.DefaultPayload("a")), elapsedFromStart: 0},
            {event: event.New(marble.DefaultPayload("c")), elapsedFromStart: 20 * time.Millisecond},
        }
        
        matchers := map[string]event.Matcher{
            "a": event.HasPayload(marble.DefaultPayload("a")),
            "b": event.HasPayload(marble.DefaultPayload("b")),
            "c": event.HasPayload(marble.DefaultPayload("c")),
        }
        
        validator := runtime.NewInterceptorValidator(
            tl,
            activity,
            matchers,
            runtime.InterceptorConfig{Strict: true},
        )
        
        errors := validator.Validate(node)
        assert.NotEmpty(t, errors)
    })
}
```

**Verifications:**
- [ ] All new tests pass
- [ ] All existing tests still pass
- [ ] Integration between components works correctly

---

## 9. Phase 8: Cleanup & Optimization

### Objective
Remove old code, optimize new code, and finalize the implementation.

### Duration
- **Estimated**: 0.5 day

### Tasks

#### 9.1 Remove Old Code

- [ ] Remove old `buildTicks()` and `handleGroup()` from `timeline.go`
- [ ] Remove old position-based code from `interceptor.go`
- [ ] Clean up deprecated functions
- [ ] Remove commented code

#### 9.2 Optimize New Code

- [ ] Add caching for parsed nodes
- [ ] Optimize visitor implementations
- [ ] Add pool for frequently used nodes

#### 9.3 Add String Representation

**File:** `eventest/internal/marble/node.go` (UPDATE)

```go
// String returns a string representation of any Node
func String(node Node) string {
    var b strings.Builder
    stringVisitor := &stringVisitor{Builder: &b}
    node.Accept(stringVisitor)
    return b.String()
}

type stringVisitor struct {
    *strings.Builder
}

func (v *stringVisitor) VisitEvent(n *EventNode) {
    v.WriteString(n.Name)
}

func (v *stringVisitor) VisitWait(n *WaitNode) {
    v.WriteRune('-')
}

func (v *stringVisitor) VisitStart(n *StartNode) {
    v.WriteRune('^')
}

func (v *stringVisitor) VisitFollowup(n *FollowupNode) {
    v.WriteString(fmt.Sprintf("%s<-%s", n.NewEvent, n.OfEvent))
}

func (v *stringVisitor) VisitSequence(n *SequenceNode) {
    for i, child := range n.Children {
        if i > 0 {
            // Add separator if previous wasn't a group end
            prev := n.Children[i-1]
            if _, isGroup := prev.(*GroupNode); !isGroup {
                // Don't add separator
            }
        }
        child.Accept(v)
    }
}

func (v *stringVisitor) VisitGroup(n *GroupNode) {
    open := '['
    close := ']'
    if !n.Ordered {
        open = '('
        close = ')'
    }
    v.WriteRune(open)
    for _, child := range n.Children {
        child.Accept(v)
    }
    v.WriteRune(close)
}
```

#### 9.4 Add Debug Utilities

**File:** `eventest/internal/marble/debug.go` (NEW)

```go
package marble

import "fmt"

// DebugPrint prints a Node tree with indentation
func DebugPrint(node Node) {
    debugPrint(node, 0)
}

func debugPrint(node Node, indent int) {
    prefix := ""
    for i := 0; i < indent; i++ {
        prefix += "  "
    }
    
    switch n := node.(type) {
    case *SequenceNode:
        fmt.Printf("%sSequenceNode (children: %d)\n", prefix, len(n.Children))
        for _, child := range n.Children {
            debugPrint(child, indent+1)
        }
    case *GroupNode:
        groupType := "Unordered"
        if n.Ordered {
            groupType = "Ordered"
        }
        fmt.Printf("%sGroupNode (%s, children: %d)\n", prefix, groupType, len(n.Children))
        for _, child := range n.Children {
            debugPrint(child, indent+1)
        }
    case *EventNode:
        fmt.Printf("%sEventNode (name: %q)\n", prefix, n.Name)
    case *WaitNode:
        fmt.Printf("%sWaitNode\n", prefix)
    case *StartNode:
        fmt.Printf("%sStartNode\n", prefix)
    case *FollowupNode:
        fmt.Printf("%sFollowupNode (new: %q, of: %q)\n", prefix, n.NewEvent, n.OfEvent)
    }
}

// ValidateAndDebug validates and prints a Node tree
func ValidateAndDebug(marbleStr string, rules ...Rule) (Node, error) {
    node, err := ParseAsNode(marbleStr)
    if err != nil {
        return nil, err
    }
    
    if err := ValidateNode(node, rules...); err != nil {
        return nil, err
    }
    
    fmt.Println("AST Structure:")
    DebugPrint(node)
    
    return node, nil
}
```

#### 9.5 Update Documentation

- [ ] Update `docs/marble.md` with new AST information
- [ ] Update `docs/marble-technical.md` with migration notes
- [ ] Add examples using new API
- [ ] Document Visitor pattern usage

**Verifications:**
- [ ] All code compiles without warnings
- [ ] All tests pass
- [ ] Documentation is up to date

---

## 10. Risk Management

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing API | Low | High | Maintain dual AST during migration |
| Performance regression | Medium | Medium | Benchmark before/after, optimize if needed |
| Bug in new AST | Medium | High | Comprehensive testing, dual implementation |
| Migration takes too long | Medium | Medium | Phase approach, prioritize critical paths |
| Team unfamiliarity with Visitor pattern | Low | Medium | Document pattern, provide examples |

### Risk Mitigation Strategies

1. **Dual AST**: Maintain both representations during migration
2. **Phase Approach**: Small, incremental changes with verification at each step
3. **Comprehensive Testing**: Extensive tests for both old and new implementations
4. **Backward Compatibility**: Ensure all existing code continues to work
5. **Documentation**: Clear documentation of changes and new patterns

---

## 11. Rollback Plan

### Partial Rollback

If issues are discovered during migration, we can roll back individual phases:

1. **Phase 8 (Cleanup)**: Revert cleanup commits, keep dual AST
2. **Phase 7 (Validation)**: Keep old tests, add new tests separately
3. **Phase 6 (Runtime)**: Revert runtime changes, keep dual AST
4. **Phase 5 (Interceptor)**: Revert interceptor changes, keep old implementation
5. **Phase 4 (Timeline)**: Revert timeline changes, keep old implementation

### Full Rollback

If major issues require full rollback:

```bash
# Rollback all changes
git checkout main -- eventest/

# Or revert specific commits
git revert <commit-hash>
```

### Rollback Verification

After any rollback:
- [ ] All tests pass
- [ ] No breaking changes to existing code
- [ ] Documentation reflects current state

---

## 12. Success Criteria

### Phase Completion Criteria

| Phase | Completion Criteria |
|-------|---------------------|
| Preparation | `plan.md` created, code inventory complete |
| Phase 1 | New AST types created, tests pass |
| Phase 2 | Parser builds hierarchical AST, backward compatible |
| Phase 3 | Dual AST works, both representations usable |
| Phase 4 | Timeline builder uses Visitor, tests pass |
| Phase 5 | Interceptor uses Visitor, tests pass |
| Phase 6 | Runtime fully integrated, Harness works |
| Phase 7 | All tests pass, integration verified |
| Phase 8 | Old code removed, cleanup complete |

### Overall Success Criteria

- [ ] **Functionality**: All existing functionality works correctly
- [ ] **Backward Compatibility**: No breaking changes to public API
- [ ] **Code Quality**: Code is clean, well-documented, and maintainable
- [ ] **Testing**: Comprehensive test coverage for new and old code
- [ ] **Performance**: No significant performance regression
- [ ] **Documentation**: All changes are documented

### Metrics

| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Test Coverage | >90% | TBD | ⏳ |
| Code Lines | <800 | ~830 | ⏳ |
| Build Time | <5s | TBD | ⏳ |
| Test Time | <30s | TBD | ⏳ |

---

## Appendix A: File Changes Summary

### New Files
- `eventest/internal/marble/node.go` - New AST types
- `eventest/internal/marble/visitor.go` - Visitor interface
- `eventest/internal/marble/node_helpers.go` - Helper functions
- `eventest/internal/marble/node_test.go` - Node tests
- `eventest/internal/marble/conversions.go` - Conversion utilities
- `eventest/internal/marble/integration_test.go` - Integration tests
- `eventest/internal/marble/debug.go` - Debug utilities
- `eventest/internal/runtime/timeline_builder.go` - New timeline builder
- `eventest/internal/runtime/interceptor_validator.go` - New validator
- `eventest/internal/runtime/timeline_builder_test.go` - Timeline builder tests
- `eventest/internal/runtime/interceptor_validator_test.go` - Validator tests

### Modified Files
- `eventest/internal/marble/parser.go` - Updated to build hierarchical AST
- `eventest/internal/marble/op.go` - Added ToNode() methods
- `eventest/internal/marble/parser_test.go` - Added ParseAsNode tests
- `eventest/internal/runtime/timeline.go` - Added NewTimelineFromNode
- `eventest/internal/runtime/runtime.go` - Updated to use new parser
- `eventest/internal/runtime/interceptor.go` - Updated validation, removed dead code
- `eventest/harness.go` - Completed implementation

### Deprecated/Removed Files
- None (initially) - Old code kept for backward compatibility

---

## Appendix B: Git Workflow

### Branch Strategy
```bash
# Create feature branch
git checkout -b feat/marble-hierarchical-ast

# Commit changes at each phase
git commit -m "feat: add hierarchical AST types"
git commit -m "feat: migrate parser to hierarchical AST"
git commit -m "feat: add timeline builder with Visitor pattern"
# ... etc
```

### Commit Messages
Follow conventional commits:
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `test:` - Test additions/updates
- `docs:` - Documentation changes
- `chore:` - Miscellaneous changes

### Pull Request
Create PR when all phases complete:
- Title: `feat: migrate marble to hierarchical AST with Visitor pattern`
- Description: Link to this plan document
- Reviewers: Team members

---

## Appendix C: Testing Strategy

### Unit Tests
- Each new type has comprehensive unit tests
- Each visitor implementation has tests
- Conversion functions have round-trip tests

### Integration Tests
- Test interaction between parser, timeline, runtime
- Test end-to-end with Harness
- Test interceptor with real event buses

### Regression Tests
- All existing tests must pass
- Add tests for edge cases discovered during migration

### Performance Tests
- Benchmark parser performance
- Benchmark timeline building
- Benchmark interceptor validation

---

## Appendix D: Timeline

### Day-by-Day Plan

| Day | Phase | Tasks |
|-----|-------|-------|
| 1 | Preparation + Phase 1 | Create plan, implement new AST types |
| 2 | Phase 2 + Phase 3 | Parser migration, dual AST |
| 3 | Phase 4 | Timeline builder migration |
| 4 | Phase 5 | Interceptor migration |
| 5 | Phase 6 + Phase 7 | Runtime integration, validation |
| 6 | Phase 8 | Cleanup, optimization, documentation |
| 7 | Buffer | Catch up, final testing |

---

## Conclusion

This implementation plan provides a **clear, incremental path** to migrate the Marble language from a flat operation list to a hierarchical AST with the Visitor pattern. The phased approach minimizes risk and ensures backward compatibility throughout the migration.

**Key Benefits:**
- ~30% code reduction
- Massive simplification of timeline and interceptor logic
- Elimination of error-prone position tracking
- Better maintainability and extensibility

**Next Steps:**
1. Review and approve this plan
2. Begin Phase 1: Foundation
3. Proceed through phases sequentially
4. Verify at each step

**Owner:** [Your Name]

**Approvers:** [Team Members]
