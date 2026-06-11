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
	clock            Clock
	payloadMap       map[string]event.Payload
	eventMap         map[string]event.Event
	matchers         map[string]event.Matcher
	baseTickDuration time.Duration
	bus              event.Bus
	publishedEvents  map[string]event.Event
}

func NewRuntime(bus event.Bus, opts ...Option) *Runtime {
	clock := NewClock()

	r := &Runtime{
		clock:            clock,
		bus:              bus,
		baseTickDuration: DefaultTickDuration,
		matchers:         make(map[string]event.Matcher),
		publishedEvents:  make(map[string]event.Event),
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.payloadMap == nil {
		r.payloadMap = make(map[string]event.Payload)
	}
	if r.eventMap == nil {
		r.eventMap = make(map[string]event.Event)
	}

	return r
}

func (r *Runtime) PublishedEvents() map[string]event.Event {
	return r.publishedEvents
}

func (r *Runtime) PublishedEvent(label string) (event.Event, bool) {
	evt, ok := r.publishedEvents[label]
	return evt, ok
}

func (r *Runtime) RunAll(marbleSeq string) error {
	node, err := marble.ParseAsNode(marbleSeq)
	if err != nil {
		return err
	}
	return r.RunAllFromNode(node)
}

func (r *Runtime) RunAllFromNode(node marble.Node) error {
	sess, err := r.RunFromNode(node)
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
	node, err := marble.ParseAsNode(marbleSeq)
	if err != nil {
		return nil, err
	}
	return r.RunFromNode(node)
}

func (r *Runtime) RunFromNode(node marble.Node) (*RunningSession, error) {
	if err := marble.Validate(node,
		marble.StartEventAnywhereRule{},
		marble.WaitlessGroupsRule{},
	); err != nil {
		return nil, err
	}

	tl := NewTimeline(node, TimelineWithTickDuration(r.baseTickDuration))
	ticks := tl.Ticks()

	return &RunningSession{
		rt:         r,
		ticks:      ticks,
		clock:      r.clock,
		bus:        r.bus,
		payloadMap: r.payloadMap,
		eventMap:   r.eventMap,
	}, nil
}

func (r *Runtime) RunFromOps(ops []marble.Op) (*RunningSession, error) {
	tl := NewTimelineFromOps(ops, TimelineWithTickDuration(r.baseTickDuration))
	ticks := tl.Ticks()

	return &RunningSession{
		rt:         r,
		ticks:      ticks,
		clock:      r.clock,
		bus:        r.bus,
		payloadMap: r.payloadMap,
		eventMap:   r.eventMap,
	}, nil
}

type RunningSession struct {
	rt         *Runtime
	ticks      []Tick
	clock      Clock
	bus        event.Bus
	payloadMap map[string]event.Payload
	eventMap   map[string]event.Event

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
			s.bus.Publish(s.getEvent(o.Name))
			s.rt.publishedEvents[o.Name] = s.getEvent(o.Name)
		case marble.EventWithFollowupOp:
			from := s.getEvent(o.OfEvent)
			to, ok := s.payloadMap[o.NewEvent]
			if !ok {
				to = DefaultPayload(o.NewEvent)
			}

			toEvt := event.NewFollowup(from, to)
			s.bus.Publish(toEvt)
			s.rt.publishedEvents[o.NewEvent] = toEvt
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

func (s *RunningSession) Clock() Clock {
	return s.clock
}

func (s *RunningSession) getEvent(label string) event.Event {
	evt, ok := s.eventMap[label]
	if ok {
		return evt
	}

	pl, ok := s.payloadMap[label]
	if !ok {
		return event.New(DefaultPayload(label))
	}
	return event.New(pl)
}
