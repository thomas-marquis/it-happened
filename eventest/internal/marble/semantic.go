package marble

import (
	"errors"
	"fmt"
)

var (
	ErrSemantic = errors.New("semantic error")
)

type Rule interface {
	Validate(node Node) error
}

type NotEmptyRule struct{}

func (r NotEmptyRule) Validate(node Node) error {
	if node == nil {
		return errors.Join(ErrSemantic, errors.New("a timeline cannot be empty"))
	}
	if seq, ok := node.(*SequenceNode); ok && len(seq.Children) == 0 {
		return errors.Join(ErrSemantic, errors.New("a timeline cannot be empty"))
	}
	return nil
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

// MandatoryInitEventRule ensures expectation chain starts with exactly one initEvent at position 0
type MandatoryInitEventRule struct{}

func (r MandatoryInitEventRule) Validate(node Node) error {
	v := &initEventVisitor{}
	node.Accept(v)

	// Must have exactly one initEvent
	if v.count != 1 {
		if v.count == 0 {
			return errors.Join(ErrSemantic, errors.New("expectation must contain exactly one initEvent (^) at the beginning"))
		}
		return errors.Join(ErrSemantic, errors.New("expectation must contain exactly one initEvent (^)"))
	}

	// First node must be initEvent (directly or nested in group)
	seq, ok := node.(*SequenceNode)
	if !ok || len(seq.Children) == 0 {
		return nil // Should be caught by NotEmptyRule
	}
	if !isFirstNodeInitEvent(seq.Children[0]) {
		return errors.Join(ErrSemantic, errors.New("initEvent (^) must be the first element in the expectation"))
	}

	return nil
}

// isFirstNodeInitEvent checks if a node is or contains an InitEventNode as its first element
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

// NoInitEventInSideEffectRule ensures side effect chain never contains initEvent
type NoInitEventInSideEffectRule struct{}

func (r NoInitEventInSideEffectRule) Validate(node Node) error {
	v := &initEventVisitor{}
	node.Accept(v)

	if v.count > 0 {
		return errors.Join(ErrSemantic, errors.New("side effect must not contain initEvent (^)"))
	}

	return nil
}

// SideEffectDurationRule ensures side effect duration does not exceed expectation duration
type SideEffectDurationRule struct {
	ExpectedDuration int
}

func (r SideEffectDurationRule) Validate(node Node) error {
	// Calculate the number of ticks in the side effect
	sideEffectTicks := countTicks(node)

	if sideEffectTicks > r.ExpectedDuration {
		return errors.Join(ErrSemantic,
			fmt.Errorf("side effect duration (%d ticks) exceeds expectation duration (%d ticks)", sideEffectTicks, r.ExpectedDuration))
	}

	return nil
}

// countTicks counts the number of top-level ticks in a node
func countTicks(node Node) int {
	seq, ok := node.(*SequenceNode)
	if !ok {
		return 1 // Single node = one tick
	}
	return len(seq.Children)
}

type WaitlessGroupsRule struct{}

func (r WaitlessGroupsRule) Validate(node Node) error {
	v := &waitlessGroupsVisitor{}
	node.Accept(v)
	if len(v.errors) > 0 {
		return errors.Join(v.errors...)
	}
	return nil
}

type waitlessGroupsVisitor struct {
	BaseVisitor
	inGroup bool
	errors  []error
}

func (v *waitlessGroupsVisitor) VisitWait(*WaitNode) {
	if v.inGroup {
		v.errors = append(v.errors, errors.Join(
			ErrSemantic,
			errors.New("a group is a single tick operation so a wait operator can't be used here"),
		))
	}
}

func (v *waitlessGroupsVisitor) VisitSequence(n *SequenceNode) {
	for _, child := range n.Children {
		child.Accept(v)
	}
}

func (v *waitlessGroupsVisitor) VisitGroup(n *GroupNode) {
	oldInGroup := v.inGroup
	v.inGroup = true
	for _, child := range n.Children {
		child.Accept(v)
	}
	v.inGroup = oldInGroup
}

func Validate(node Node, rules ...Rule) error {
	for _, rule := range rules {
		if err := rule.Validate(node); err != nil {
			return err
		}
	}
	return nil
}
