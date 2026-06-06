package runtime

import (
	"errors"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

const (
	DefaultPayloadType event.Type = "__eventest__.default"
)

var (
	ErrRuntime   = errors.New("runtime error")
	SessionEnded = errors.New("session ended")
)

type DefaultPayload string

func (DefaultPayload) Type() event.Type {
	return DefaultPayloadType
}

type Runtime struct {
	clock            VirtualClock
	payloadMap       map[string]event.Payload
	matchers         map[string]event.Matcher
	baseTickDuration time.Duration
	bus              event.Bus
}

func NewRuntime(bus event.Bus, opts ...Option) *Runtime {
	clock := NewVirtualClock()

	r := &Runtime{
		clock:            clock,
		bus:              bus,
		baseTickDuration: DefaultTickDuration,
		matchers:         make(map[string]event.Matcher),
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.payloadMap == nil {
		r.payloadMap = make(map[string]event.Payload)
	}

	return r
}

func (r *Runtime) RunAll(marbleSeq string) error {
	sess, err := r.Run(marbleSeq)
	if err != nil {
		return err
	}

	for sess.HasNext() {
		if err := sess.Next(); err != nil {
			if errors.Is(err, SessionEnded) {
				err = nil
			}
			return err
		}
	}

	return nil
}

func (r *Runtime) Run(marbleSeq string) (*RunningSession, error) {
	ops, err := marble.Parse(marbleSeq)
	if err != nil {
		return nil, err
	}

	if err := marble.Validate(ops,
		marble.StartEventAnywhereRule{},
		marble.WaitlessGroupsRule{},
	); err != nil {
		return nil, err
	}

	tl := NewTimeline(ops, TimelineWithTickDuration(r.baseTickDuration))
	ticks := tl.Ticks()

	return &RunningSession{
		rt:         r,
		ticks:      ticks,
		clock:      r.clock,
		bus:        r.bus,
		payloadMap: r.payloadMap,
	}, nil
}

type RunningSession struct {
	rt         *Runtime
	ticks      []Tick
	clock      VirtualClock
	bus        event.Bus
	payloadMap map[string]event.Payload

	current int
}

func (s *RunningSession) Next() error {
	if s.current >= len(s.ticks) {
		if s.clock.Started() {
			s.clock.Stop()
		}
		return SessionEnded
	}
	if s.current == 0 && !s.clock.Started() {
		s.clock.Start()
	}

	tick := s.ticks[s.current]

	for _, op := range tick.Ops {
		switch o := op.(type) {
		case marble.EventOp:
			p, found := s.payloadMap[o.Name]
			if !found {
				p = DefaultPayload(o.Name)
			}
			s.bus.Publish(event.New(p))
		}
	}
	s.clock.Forward(tick.Duration)
	s.current++
	return nil
}

func (s *RunningSession) HasNext() bool {
	return s.current < len(s.ticks)
}

func (s *RunningSession) CurrentTick() Tick {
	if s.current >= len(s.ticks) {
		return Tick{}
	}
	return s.ticks[s.current]
}

func (s *RunningSession) Clock() VirtualClock {
	return s.clock
}
