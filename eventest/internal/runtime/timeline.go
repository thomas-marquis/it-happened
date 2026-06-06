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

func NewTimeline(rowOps []marble.Op, opts ...TimelineOption) *Timeline {
	if err := marble.Validate(rowOps, marble.WaitlessGroupsRule{}); err != nil {
		panic(err)
	}

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

	t.ticks = t.buildTicks(rowOps)

	return t
}

func (t *Timeline) Ticks() []Tick {
	return t.ticks
}

func (t *Timeline) buildTicks(rowOps []marble.Op) []Tick {
	var (
		current    int
		ticks      []Tick
		grpExitPos int // 0 if not in a group
		pos        int
	)

	for pos < len(rowOps) {
		op := rowOps[pos]
		switch o := op.(type) {
		case marble.EventOp, marble.EventWithFollowupOp:
			if grpExitPos == 0 {
				t := Tick{
					Duration: t.tickDuration,
					Ops:      []marble.Op{o},
				}
				ticks = append(ticks, t)
				current++
				pos++
				continue
			}

			// check whether we need to create the new tick for the entier group
			if len(ticks) == current {
				ticks = append(ticks, Tick{
					Duration: t.tickDuration,
					Ops:      make([]marble.Op, 0),
				})
			}

			ticks[current].Ops = append(ticks[current].Ops, o)

		case marble.OrderedGroupStartOp, marble.UnorderedGroupStartOp:
			grpOps := t.handleGroup(rowOps, &pos)
			ticks = append(ticks, Tick{
				Duration: t.tickDuration,
				Ops:      grpOps,
			})
			current++
		case marble.WaitOp:
			if grpExitPos == 0 { // not supposed to happen, the validator should have caught it... or not...
				continue
			}
			ticks = append(ticks, Tick{
				Duration: t.tickDuration,
				Ops:      []marble.Op{o},
			})
			pos++
			current++
		}
	}

	return ticks
}

func (t *Timeline) handleGroup(ops []marble.Op, pos *int) (result []marble.Op) {
	var (
		exitPos       int
		parts         [][]marble.Op
		startPos      = *pos
		shuffleNeeded bool
	)
	defer func() {
		if shuffleNeeded {
			t.shuffleParts(parts)
		}
		result = flattenParts(parts)
	}()

	for *pos < len(ops) {
		switch op := ops[*pos].(type) {
		case marble.OrderedGroupStartOp:
			if *pos == startPos {
				exitPos = op.EndPos
				*pos++
			} else {
				parts = append(parts, t.handleGroup(ops, pos))
			}

		case marble.UnorderedGroupStartOp:
			if *pos == startPos {
				shuffleNeeded = true
				exitPos = op.EndPos
				*pos++
			} else {
				parts = append(parts, t.handleGroup(ops, pos))
			}

		case marble.EventOp, marble.EventWithFollowupOp:
			parts = append(parts, []marble.Op{op})
			*pos++

		case marble.UnorderedGroupEndOp:
			if exitPos == *pos {
				*pos++
				return
			}

		case marble.OrderedGroupEndOp:
			if exitPos == *pos {
				*pos++
				return
			}

		default:
			panic("unknown group op")
		}
	}

	return
}

func (t *Timeline) shuffleParts(ops [][]marble.Op) {
	t.randGen.Shuffle(len(ops), func(i, j int) {
		ops[i], ops[j] = ops[j], ops[i]
	})
}

func (t *Timeline) seed(seed uint64) {
	t.randGen = rand.New(rand.NewPCG(seed, seed*2))
}

func flattenParts(parts [][]marble.Op) []marble.Op {
	result := make([]marble.Op, 0, len(parts))
	for _, part := range parts {
		result = append(result, part...)
	}
	return result
}
