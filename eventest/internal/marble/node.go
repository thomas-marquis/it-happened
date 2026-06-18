package marble

// Position represents a source location for error reporting
type Position struct {
	Line   int
	Column int
	Offset int
}

// Node represents any element in the marble AST
type Node interface {
	// Accept allows the Visitor pattern to traverse the AST
	Accept(Visitor)

	// Position provides source location (optional, for error reporting)
	Position() Position
}

// Leaf Nodes (Terminal)

type EventNode struct {
	Name string
	pos  Position
}

func (n *EventNode) Accept(v Visitor)   { v.VisitEvent(n) }
func (n *EventNode) Position() Position { return n.pos }

type WaitNode struct {
	pos Position
}

func (n *WaitNode) Accept(v Visitor)   { v.VisitWait(n) }
func (n *WaitNode) Position() Position { return n.pos }

type InitEventNode struct {
	pos Position
}

func (n *InitEventNode) Accept(v Visitor)   { v.VisitInitEvent(n) }
func (n *InitEventNode) Position() Position { return n.pos }

type FollowupNode struct {
	NewEvent string
	OfEvent  string
	pos      Position
}

func (n *FollowupNode) Accept(v Visitor)   { v.VisitFollowup(n) }
func (n *FollowupNode) Position() Position { return n.pos }

// Composite Nodes (Non-Terminal)

type SequenceNode struct {
	Children []Node
	pos      Position
}

func (n *SequenceNode) Accept(v Visitor)   { v.VisitSequence(n) }
func (n *SequenceNode) Position() Position { return n.pos }

type GroupNode struct {
	Children []Node
	Ordered  bool
	pos      Position
}

func (n *GroupNode) Accept(v Visitor)   { v.VisitGroup(n) }
func (n *GroupNode) Position() Position { return n.pos }
