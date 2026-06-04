package eventest

import "errors"

var (
	ErrSemantic = errors.New("semantic error")
)

type semanticRule interface {
	Validate(seq []op) error
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
