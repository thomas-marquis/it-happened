package marble

// Visitor defines the interface for AST traversal
type Visitor interface {
	// Leaf nodes
	VisitEvent(*EventNode)
	VisitWait(*WaitNode)
	VisitPlaceholder(*PlaceholderNode)
	VisitFollowup(*FollowupNode)

	// Composite nodes
	VisitSequence(*SequenceNode)
	VisitGroup(*GroupNode)
}

// BaseVisitor provides no-op implementations for all Visit methods
type BaseVisitor struct{}

func (v *BaseVisitor) VisitEvent(*EventNode)             {}
func (v *BaseVisitor) VisitWait(*WaitNode)               {}
func (v *BaseVisitor) VisitPlaceholder(*PlaceholderNode) {}
func (v *BaseVisitor) VisitFollowup(*FollowupNode)       {}
func (v *BaseVisitor) VisitSequence(*SequenceNode)       {}
func (v *BaseVisitor) VisitGroup(*GroupNode)             {}
