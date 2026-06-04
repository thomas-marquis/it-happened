package marble

import "errors"

var (
	ErrSemantic = errors.New("semantic error")
)

type Rule interface {
	Validate(seq []Op) error
}

type NotEmptyRule struct{}

func (r NotEmptyRule) Validate(seq []Op) error {
	if len(seq) == 0 {
		return errors.Join(ErrSemantic, errors.New("a timeline cannot be empty"))
	}

	return nil
}

type StartEventAtBeginningRule struct{}

func (r StartEventAtBeginningRule) Validate(seq []Op) error {
	count := countStartEvents(seq)
	if count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	if count == 1 && len(seq) > 0 && !isStartEvent(seq[0]) {
		return errors.Join(ErrSemantic, errors.New("the start event must be at the beginning of the timeline"))
	}

	return nil
}

type StartEventAnywhereRule struct{}

func (r StartEventAnywhereRule) Validate(seq []Op) error {
	count := countStartEvents(seq)
	if count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	return nil
}

type UniqueStartEventRule struct{}

func (r UniqueStartEventRule) Validate(seq []Op) error {
	count := countStartEvents(seq)
	if count != 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline must have exactly one start event"))
	}

	return nil
}

func countStartEvents(seq []Op) int {
	count := 0
	for _, o := range seq {
		count += countInOp(o)
	}
	return count
}

func countInOp(o Op) int {
	if o.Type() == StartEventOpType {
		return 1
	}

	var subOps []Op
	if o.Type() == UnorderedGroupOpType {
		x := o.(UnorderedGroupOp)
		subOps = x.Ops
	} else if o.Type() == OrderedGroupOpType {
		x := o.(OrderedGroupOp)
		subOps = x.Ops
	}

	count := 0
	for _, subOp := range subOps {
		count += countInOp(subOp)
	}

	return count
}

func isStartEvent(o Op) bool {
	if o.Type() == StartEventOpType {
		return true
	}

	var subOps []Op
	if o.Type() == UnorderedGroupOpType {
		x := o.(UnorderedGroupOp)
		subOps = x.Ops
	} else if o.Type() == OrderedGroupOpType {
		x := o.(OrderedGroupOp)
		subOps = x.Ops
	}

	for _, subOp := range subOps {
		if isStartEvent(subOp) {
			return true
		}
	}

	return false
}

type SingleTickGroupRule struct{}

func (r SingleTickGroupRule) Validate(seq []Op) error {
	for _, o := range seq {
		var subOps []Op
		if o.Type() == UnorderedGroupOpType {
			x := o.(UnorderedGroupOp)
			subOps = x.Ops
		} else if o.Type() == OrderedGroupOpType {
			x := o.(OrderedGroupOp)
			subOps = x.Ops
		}

		for _, subOp := range subOps {
			if subOp.Type() == WaitOpType {
				return errors.Join(
					ErrSemantic,
					errors.New("a group is a single tick operation so a wait operator can't be used here"),
				)
			}
			if subOp.Type() == UnorderedGroupOpType || subOp.Type() == OrderedGroupOpType {
				if err := r.Validate([]Op{subOp}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func Validate(ops []Op, rules ...Rule) error {
	for _, rule := range rules {
		if err := rule.Validate(ops); err != nil {
			return err
		}
	}
	return nil
}
