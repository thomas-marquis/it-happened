package marble

type OpType int

type Op interface {
	Type() OpType
}

const (
	WaitOpType OpType = iota
	EventOpType
	StartEventOpType
	EventWithFollowupOpType

	// Deprecated
	OrderedGroupOpType
	// Deprecated
	UnorderedGroupOpType

	OrderedGroupStartType
	OrderedGroupEndType
	UnorderedGroupStartType
	UnorderedGroupEndType
)

type WaitOp struct{}

func (o WaitOp) Type() OpType {
	return WaitOpType
}

type EventOp struct {
	Name string
}

func (o EventOp) Type() OpType {
	return EventOpType
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

// Deprecated
type OrderedGroupOp struct {
	Ops []Op
}

func (o OrderedGroupOp) Type() OpType {
	return OrderedGroupOpType
}

// Deprecated
type UnorderedGroupOp struct {
	Ops []Op
}

type OrderedGroupStartOp struct {
	EndPos int
}

func (o OrderedGroupStartOp) Type() OpType {
	return OrderedGroupStartType
}

type OrderedGroupEndOp struct {
	StartPos int
}

func (o OrderedGroupEndOp) Type() OpType {
	return OrderedGroupEndType
}

type UnorderedGroupStartOp struct {
	EndPos int
}

func (o UnorderedGroupStartOp) Type() OpType {
	return UnorderedGroupStartType
}

type UnorderedGroupEndOp struct {
	StartPos int
}

func (o UnorderedGroupEndOp) Type() OpType {
	return UnorderedGroupEndType
}
