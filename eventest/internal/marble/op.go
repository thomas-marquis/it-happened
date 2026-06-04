package marble

import "time"

type OpType int

type Op interface {
	Type() OpType
}

const (
	WaitOpType OpType = iota
	EventOpType
	StartEventOpType
	OrderedGroupOpType
	UnorderedGroupOpType
	EventWithFollowupOpType
)

type WaitOp struct {
	Duration time.Duration
}

func (o WaitOp) Type() OpType {
	return WaitOpType
}

type EventOp struct {
	Name string
}

func (o EventOp) Type() OpType {
	return EventOpType
}

type OrderedGroupOp struct {
	Ops []Op
}

func (o OrderedGroupOp) Type() OpType {
	return OrderedGroupOpType
}

type UnorderedGroupOp struct {
	Ops []Op
}

func (o UnorderedGroupOp) Type() OpType {
	return UnorderedGroupOpType
}

type StartEventOp struct{}

func (o StartEventOp) Type() OpType {
	return StartEventOpType
}

type EventWithFollowupOp struct {
	EventName string
	From      string
}

func (o EventWithFollowupOp) Type() OpType {
	return EventWithFollowupOpType
}
