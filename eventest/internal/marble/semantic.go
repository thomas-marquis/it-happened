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

	if count == 1 && len(seq) > 0 && !isStartEvent(seq, 0) {
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
		if o.Type() == StartEventOpType {
			count++
		}
	}
	return count
}

func isStartEvent(seq []Op, index int) bool {
	o := seq[index]
	if o.Type() == StartEventOpType {
		return true
	}

	var endPos int
	if o.Type() == UnorderedGroupStartType {
		endPos = o.(UnorderedGroupStartOp).EndPos
	} else if o.Type() == OrderedGroupStartType {
		endPos = o.(OrderedGroupStartOp).EndPos
	} else {
		return false
	}

	for i := index + 1; i < endPos; i++ {
		if seq[i].Type() == StartEventOpType {
			return true
		}
	}

	return false
}

type WaitlessGroupsRule struct{}

func (r WaitlessGroupsRule) Validate(seq []Op) error {
	for i, o := range seq {
		var endPos int
		if o.Type() == UnorderedGroupStartType {
			endPos = o.(UnorderedGroupStartOp).EndPos
		} else if o.Type() == OrderedGroupStartType {
			endPos = o.(OrderedGroupStartOp).EndPos
		} else {
			continue
		}

		for j := i + 1; j < endPos; j++ {
			if seq[j].Type() == WaitOpType {
				return errors.Join(
					ErrSemantic,
					errors.New("a group is a single tick operation so a wait operator can't be used here"),
				)
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
