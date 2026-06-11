package marble

import "errors"

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

type StartEventAtBeginningRule struct{}

func (r StartEventAtBeginningRule) Validate(node Node) error {
	v := &startEventVisitor{}
	node.Accept(v)

	if v.count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	if v.count == 1 {
		seq, ok := node.(*SequenceNode)
		if !ok || len(seq.Children) == 0 {
			return nil // Should be caught by NotEmptyRule
		}
		if !isFirstNodeStart(seq.Children[0]) {
			return errors.Join(ErrSemantic, errors.New("the start event must be at the beginning of the timeline"))
		}
	}

	return nil
}

func isFirstNodeStart(n Node) bool {
	switch node := n.(type) {
	case *StartNode:
		return true
	case *GroupNode:
		if len(node.Children) > 0 {
			return isFirstNodeStart(node.Children[0])
		}
	case *SequenceNode:
		if len(node.Children) > 0 {
			return isFirstNodeStart(node.Children[0])
		}
	}
	return false
}

type StartEventAnywhereRule struct{}

func (r StartEventAnywhereRule) Validate(node Node) error {
	v := &startEventVisitor{}
	node.Accept(v)

	if v.count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	return nil
}

type UniqueStartEventRule struct{}

func (r UniqueStartEventRule) Validate(node Node) error {
	v := &startEventVisitor{}
	node.Accept(v)

	if v.count != 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline must have exactly one start event"))
	}

	return nil
}

type startEventVisitor struct {
	BaseVisitor
	count int
}

func (v *startEventVisitor) VisitStart(*StartNode) {
	v.count++
}

func (v *startEventVisitor) VisitSequence(n *SequenceNode) {
	for _, child := range n.Children {
		child.Accept(v)
	}
}

func (v *startEventVisitor) VisitGroup(n *GroupNode) {
	for _, child := range n.Children {
		child.Accept(v)
	}
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

func (v *waitlessGroupsVisitor) VisitWait(n *WaitNode) {
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
