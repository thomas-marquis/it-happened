package runtime

import (
	"math/rand/v2"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

const (
	DefaultTickDuration = 10 * time.Millisecond
)

type Tick struct {
	Duration time.Duration
	Ops      []marble.Op
}

type Timeline struct {
	events       map[string]event.Event
	ticks        []Tick
	randGen      *rand.Rand
	tickDuration time.Duration
}

func NewTimeline(node marble.Node, opts ...TimelineOption) *Timeline {
	t := &Timeline{
		events: make(map[string]event.Event),
		randGen: rand.New(
			rand.NewPCG(
				uint64(time.Now().UnixNano()), uint64(time.Now().UnixMilli()))),
		tickDuration: DefaultTickDuration,
	}

	for _, opt := range opts {
		opt(t)
	}

	builder := NewTimelineBuilder(t.tickDuration, t.randGen)
	ticks, err := builder.Build(node)
	if err != nil {
		panic(err)
	}
	t.ticks = ticks

	return t
}

func NewTimelineFromOps(ops []marble.Op, opts ...TimelineOption) *Timeline {
	t := &Timeline{
		events: make(map[string]event.Event),
		randGen: rand.New(
			rand.NewPCG(
				uint64(time.Now().UnixNano()), uint64(time.Now().UnixMilli()))),
		tickDuration: DefaultTickDuration,
	}

	for _, opt := range opts {
		opt(t)
	}

	t.ticks = buildTicksFromOps(ops, t.tickDuration)

	return t
}

func buildTicksFromOps(ops []marble.Op, duration time.Duration) []Tick {
	var (
		ticks []Tick
		pos   int
	)

	for pos < len(ops) {
		op := ops[pos]
		switch op.Type() {
		case marble.OrderedGroupStartType, marble.UnorderedGroupStartType:
			var endPos int
			if o, ok := op.(marble.OrderedGroupStartOp); ok {
				endPos = o.EndPos
			} else {
				endPos = op.(marble.UnorderedGroupStartOp).EndPos
			}
			ticks = append(ticks, Tick{
				Duration: duration,
				Ops:      ops[pos : endPos+1],
			})
			pos = endPos + 1
		case marble.WaitOpType, marble.EventOpType, marble.StartEventOpType, marble.EventWithFollowupOpType:
			ticks = append(ticks, Tick{
				Duration: duration,
				Ops:      []marble.Op{op},
			})
			pos++
		default:
			pos++
		}
	}
	return ticks
}

func (t *Timeline) Ticks() []Tick {
	return t.ticks
}

func (t *Timeline) seed(seed uint64) {
	t.randGen = rand.New(rand.NewPCG(seed, seed*2))
}
