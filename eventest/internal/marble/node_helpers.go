package marble

import (
	"fmt"
	"strings"
)

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

func (b *opListBuilder) VisitPlaceholder(n *PlaceholderNode) {
	b.ops = append(b.ops, PlaceholderEventOp{})
}

func (b *opListBuilder) VisitFollowup(n *FollowupNode) {
	b.ops = append(b.ops, EventWithFollowupOp{
		NewEvent: n.NewEvent,
		OfEvent:  n.OfEvent,
	})
}

func (b *opListBuilder) VisitSequence(n *SequenceNode) {
	for _, child := range n.Children {
		child.Accept(b)
	}
}

func (b *opListBuilder) VisitGroup(n *GroupNode) {
	startIdx := len(b.ops)
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

	endIdx := len(b.ops)
	// Add end marker
	if n.Ordered {
		b.ops = append(b.ops, OrderedGroupEndOp{StartPos: startIdx})
	} else {
		b.ops = append(b.ops, UnorderedGroupEndOp{StartPos: startIdx})
	}

	// Update start marker with end index
	if n.Ordered {
		b.ops[startIdx] = OrderedGroupStartOp{EndPos: endIdx}
	} else {
		b.ops[startIdx] = UnorderedGroupStartOp{EndPos: endIdx}
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

func (b *stringBuilder) VisitPlaceholder(n *PlaceholderNode) {
	b.WriteRune('^')
}

func (b *stringBuilder) VisitFollowup(n *FollowupNode) {
	b.WriteString(fmt.Sprintf("%s<-%s", n.NewEvent, n.OfEvent))
}

func (b *stringBuilder) VisitSequence(n *SequenceNode) {
	for _, child := range n.Children {
		child.Accept(b)
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
