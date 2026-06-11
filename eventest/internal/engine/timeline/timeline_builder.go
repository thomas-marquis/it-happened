package timeline

import (
	"math/rand/v2"
	"time"

	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

type TimelineBuilder struct {
	ticks        []Tick
	currentOps   []marble.Op
	tickDuration time.Duration
	randGen      *rand.Rand
	err          error
}

func NewTimelineBuilder(tickDuration time.Duration, randGen *rand.Rand) *TimelineBuilder {
	return &TimelineBuilder{
		tickDuration: tickDuration,
		randGen:      randGen,
	}
}

func (b *TimelineBuilder) Build(root marble.Node) ([]Tick, error) {
	root.Accept(b)
	return b.ticks, b.err
}

func (b *TimelineBuilder) VisitEvent(n *marble.EventNode) {
	op := marble.EventOp{Name: n.Name}
	if b.currentOps == nil {
		// Top-level event = new tick
		b.ticks = append(b.ticks, Tick{
			Duration: b.tickDuration,
			Ops:      []marble.Op{op},
		})
	} else {
		// Inside group = add to current tick
		b.currentOps = append(b.currentOps, op)
	}
}

func (b *TimelineBuilder) VisitWait(n *marble.WaitNode) {
	op := marble.WaitOp{}
	if b.currentOps == nil {
		b.ticks = append(b.ticks, Tick{
			Duration: b.tickDuration,
			Ops:      []marble.Op{op},
		})
	} else {
		b.currentOps = append(b.currentOps, op)
	}
}

func (b *TimelineBuilder) VisitStart(n *marble.StartNode) {
	op := marble.StartEventOp{}
	if b.currentOps == nil {
		b.ticks = append(b.ticks, Tick{
			Duration: b.tickDuration,
			Ops:      []marble.Op{op},
		})
	} else {
		b.currentOps = append(b.currentOps, op)
	}
}

func (b *TimelineBuilder) VisitFollowup(n *marble.FollowupNode) {
	op := marble.EventWithFollowupOp{
		NewEvent: n.NewEvent,
		OfEvent:  n.OfEvent,
	}
	if b.currentOps == nil {
		b.ticks = append(b.ticks, Tick{
			Duration: b.tickDuration,
			Ops:      []marble.Op{op},
		})
	} else {
		b.currentOps = append(b.currentOps, op)
	}
}

func (b *TimelineBuilder) VisitSequence(n *marble.SequenceNode) {
	for _, child := range n.Children {
		child.Accept(b)
	}
}

func (b *TimelineBuilder) VisitGroup(n *marble.GroupNode) {
	isTopLevel := b.currentOps == nil

	if isTopLevel {
		b.currentOps = []marble.Op{}
	}

	startIdx := len(b.currentOps)

	// Start marker
	var startOp marble.Op
	if n.Ordered {
		startOp = marble.OrderedGroupStartOp{}
	} else {
		startOp = marble.UnorderedGroupStartOp{}
	}
	b.currentOps = append(b.currentOps, startOp)

	if n.Ordered {
		// Ordered: visit children directly to maintain absolute positions
		for _, child := range n.Children {
			child.Accept(b)
		}
	} else {
		// Unordered: collect, shuffle, and then adjust positions
		var childrenOps [][]marble.Op
		oldOps := b.currentOps
		for _, child := range n.Children {
			b.currentOps = []marble.Op{}
			child.Accept(b)
			childrenOps = append(childrenOps, b.currentOps)
		}
		b.currentOps = oldOps

		// Shuffle
		if b.randGen != nil {
			b.randGen.Shuffle(len(childrenOps), func(i, j int) {
				childrenOps[i], childrenOps[j] = childrenOps[j], childrenOps[i]
			})
		}

		// Append and adjust
		for _, ops := range childrenOps {
			adjustment := len(b.currentOps)
			for _, op := range ops {
				b.currentOps = append(b.currentOps, adjustOpPositions(op, adjustment))
			}
		}
	}

	// End marker
	endIdx := len(b.currentOps)
	var endOp marble.Op
	if n.Ordered {
		endOp = marble.OrderedGroupEndOp{StartPos: startIdx}
	} else {
		endOp = marble.UnorderedGroupEndOp{StartPos: startIdx}
	}
	b.currentOps = append(b.currentOps, endOp)

	// Update start marker with absolute end position
	if n.Ordered {
		b.currentOps[startIdx] = marble.OrderedGroupStartOp{EndPos: endIdx}
	} else {
		b.currentOps[startIdx] = marble.UnorderedGroupStartOp{EndPos: endIdx}
	}

	if isTopLevel {
		b.ticks = append(b.ticks, Tick{
			Duration: b.tickDuration,
			Ops:      b.currentOps,
		})
		b.currentOps = nil
	}
}

func adjustOpPositions(op marble.Op, adjustment int) marble.Op {
	switch o := op.(type) {
	case marble.OrderedGroupStartOp:
		return marble.OrderedGroupStartOp{EndPos: o.EndPos + adjustment}
	case marble.OrderedGroupEndOp:
		return marble.OrderedGroupEndOp{StartPos: o.StartPos + adjustment}
	case marble.UnorderedGroupStartOp:
		return marble.UnorderedGroupStartOp{EndPos: o.EndPos + adjustment}
	case marble.UnorderedGroupEndOp:
		return marble.UnorderedGroupEndOp{StartPos: o.StartPos + adjustment}
	default:
		return op
	}
}
