package eventest

import "errors"

var (
	ErrSemantic = errors.New("semantic error")
)

type semanticRule interface {
	Validate(seq []op) error
}

type notEmptyRule struct{}

func (r notEmptyRule) Validate(seq []op) error {
	if len(seq) == 0 {
		return errors.Join(ErrSemantic, errors.New("a timeline cannot be empty"))
	}

	return nil
}

type startEventAtBeginningRule struct{}

func (r startEventAtBeginningRule) Validate(seq []op) error {
	count := countStartEvents(seq)
	if count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	if count == 1 && len(seq) > 0 && !isStartEvent(seq[0]) {
		return errors.Join(ErrSemantic, errors.New("the start event must be at the beginning of the timeline"))
	}

	return nil
}

type startEventAnywhereRule struct{}

func (r startEventAnywhereRule) Validate(seq []op) error {
	count := countStartEvents(seq)
	if count > 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline can have at most one start event"))
	}

	return nil
}

type uniqueStartEventRule struct{}

func (r uniqueStartEventRule) Validate(seq []op) error {
	count := countStartEvents(seq)
	if count != 1 {
		return errors.Join(ErrSemantic, errors.New("a timeline must have exactly one start event"))
	}

	return nil
}

func countStartEvents(seq []op) int {
	count := 0
	for _, o := range seq {
		count += countInOp(o)
	}
	return count
}

func countInOp(o op) int {
	if o.Type() == startEventOpType {
		return 1
	}

	var subOps []op
	if o.Type() == unorderedGroupOpType {
		x := o.(unorderedGroupOp)
		subOps = x.Ops
	} else if o.Type() == orderedGroupOpType {
		x := o.(orderedGroupOp)
		subOps = x.Ops
	}

	count := 0
	for _, subOp := range subOps {
		count += countInOp(subOp)
	}

	return count
}

func isStartEvent(o op) bool {
	if o.Type() == startEventOpType {
		return true
	}

	var subOps []op
	if o.Type() == unorderedGroupOpType {
		x := o.(unorderedGroupOp)
		subOps = x.Ops
	} else if o.Type() == orderedGroupOpType {
		x := o.(orderedGroupOp)
		subOps = x.Ops
	}

	for _, subOp := range subOps {
		if isStartEvent(subOp) {
			return true
		}
	}

	return false
}

type singleTickGroupRule struct{}

func (r singleTickGroupRule) Validate(seq []op) error {
	for _, o := range seq {
		var subOps []op
		if o.Type() == unorderedGroupOpType {
			x := o.(unorderedGroupOp)
			subOps = x.Ops
		} else if o.Type() == orderedGroupOpType {
			x := o.(orderedGroupOp)
			subOps = x.Ops
		}

		for _, subOp := range subOps {
			if subOp.Type() == waitOpType {
				return errors.Join(
					ErrSemantic,
					errors.New("a group is a single tick operation so a wait operator can't be used here"),
				)
			}
			if subOp.Type() == unorderedGroupOpType || subOp.Type() == orderedGroupOpType {
				if err := r.Validate([]op{subOp}); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateMarble(ops []op, rules ...semanticRule) error {
	for _, rule := range rules {
		if err := rule.Validate(ops); err != nil {
			return err
		}
	}
	return nil
}
