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
	ErrRuntime = errors.New("runtime error")
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

func NewRuntime(bus event.Bus, payloadMap map[string]event.Payload) *Runtime {
	clock := NewVirtualClock()
	if payloadMap == nil {
		payloadMap = make(map[string]event.Payload)
	}

	return &Runtime{
		clock:            clock,
		payloadMap:       payloadMap,
		bus:              bus,
		baseTickDuration: 100 * time.Millisecond,
		matchers:         make(map[string]event.Matcher),
	}
}

func (r *Runtime) Run(marbleSeq string) error {
	ops, err := marble.Parse(marbleSeq)
	if err != nil {
		return err
	}

	if err := marble.Validate(ops,
		marble.StartEventAnywhereRule{},
		marble.WaitlessGroupsRule{},
	); err != nil {
		return err
	}

	tl := NewTimeline(ops)
	ticks := tl.Ticks()

	r.clock.Start()
	defer r.clock.Stop()
	for _, tick := range ticks {
		for _, op := range tick.Ops {
			switch o := op.(type) {
			case marble.EventOp:
				p, found := r.payloadMap[o.Name]
				if !found {
					p = DefaultPayload(o.Name)
				}
				r.bus.Publish(event.New(p))
			}
		}
		r.clock.Forward(tick.Duration)
	}

	return nil
}
