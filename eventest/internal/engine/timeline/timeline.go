package timeline

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
	events            map[string]event.Event
	ticks             []Tick
	randGen           *rand.Rand
	tickDuration      time.Duration
	placeholderEvents []event.Event
}

func NewTimeline(node marble.Node, opts ...Option) *Timeline {
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

func (t *Timeline) Ticks() []Tick {
	return t.ticks
}

func (t *Timeline) seed(seed uint64) {
	t.randGen = rand.New(rand.NewPCG(seed, seed*2))
}
