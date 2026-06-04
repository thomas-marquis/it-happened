package eventest

import "time"

type opType int

type op interface {
	Type() opType
}

const (
	waitOpType opType = iota
	eventOpType
	startEventOpType
	orderedGroupOpType
	unorderedGroupOpType
	eventWithFollowupOpType
)

type waitOp struct {
	Duration time.Duration
}

func (o waitOp) Type() opType {
	return waitOpType
}

type eventOp struct {
	Name string
}

func (o eventOp) Type() opType {
	return eventOpType
}

type orderedGroupOp struct {
	Ops []op
}

func (o orderedGroupOp) Type() opType {
	return orderedGroupOpType
}

type unorderedGroupOp struct {
	Ops []op
}

func (o unorderedGroupOp) Type() opType {
	return unorderedGroupOpType
}

type startEventOp struct{}

func (o startEventOp) Type() opType {
	return startEventOpType
}

type eventWithFollowupOp struct {
	EventName string
	From      string
}

func (o eventWithFollowupOp) Type() opType {
	return eventWithFollowupOpType
}
