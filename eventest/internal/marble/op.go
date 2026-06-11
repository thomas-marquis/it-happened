package marble

import "fmt"

type OpType int

type Op interface {
	Type() OpType
}

const (
	WaitOpType OpType = iota
	EventOpType
	StartEventOpType
	EventWithFollowupOpType

	OrderedGroupStartType
	OrderedGroupEndType
	UnorderedGroupStartType
	UnorderedGroupEndType
)

type WaitOp struct{}

func (o WaitOp) Type() OpType {
	return WaitOpType
}

func (o WaitOp) ToNode() Node {
	return &WaitNode{}
}

func (o WaitOp) String() string {
	return "-"
}

type EventOp struct {
	Name string
}

func (o EventOp) Type() OpType {
	return EventOpType
}

func (o EventOp) ToNode() Node {
	return &EventNode{Name: o.Name}
}

func (o EventOp) String() string {
	return o.Name
}

type StartEventOp struct{}

func (o StartEventOp) Type() OpType {
	return StartEventOpType
}

func (o StartEventOp) ToNode() Node {
	return &StartNode{}
}

func (o StartEventOp) String() string {
	return "^"
}

type EventWithFollowupOp struct {
	NewEvent string
	OfEvent  string
}

func (o EventWithFollowupOp) String() string {
	return fmt.Sprintf("%s<-%s", o.NewEvent, o.OfEvent)
}

func (o EventWithFollowupOp) Type() OpType {
	return EventWithFollowupOpType
}

func (o EventWithFollowupOp) ToNode() Node {
	return &FollowupNode{
		NewEvent: o.NewEvent,
		OfEvent:  o.OfEvent,
	}
}

type OrderedGroupStartOp struct {
	EndPos int
}

func (o OrderedGroupStartOp) Type() OpType {
	return OrderedGroupStartType
}

func (o OrderedGroupStartOp) String() string {
	return "["
}

func (o OrderedGroupStartOp) ToNode() Node {
	return &GroupNode{Ordered: true}
}

type OrderedGroupEndOp struct {
	StartPos int
}

func (o OrderedGroupEndOp) Type() OpType {
	return OrderedGroupEndType
}

func (o OrderedGroupEndOp) String() string {
	return "]"
}

func (o OrderedGroupEndOp) ToNode() Node {
	return &GroupNode{Ordered: true}
}

type UnorderedGroupStartOp struct {
	EndPos int
}

func (o UnorderedGroupStartOp) Type() OpType {
	return UnorderedGroupStartType
}

func (o UnorderedGroupStartOp) String() string {
	return "("
}

func (o UnorderedGroupStartOp) ToNode() Node {
	return &GroupNode{Ordered: false}
}

type UnorderedGroupEndOp struct {
	StartPos int
}

func (o UnorderedGroupEndOp) Type() OpType {
	return UnorderedGroupEndType
}

func (o UnorderedGroupEndOp) String() string {
	return ")"
}

func (o UnorderedGroupEndOp) ToNode() Node {
	return &GroupNode{Ordered: false}
}
