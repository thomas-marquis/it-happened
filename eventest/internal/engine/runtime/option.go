package runtime

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/clock"
)

type Option func(*Runtime)

func WithBaseTickDuration(d time.Duration) Option {
	return func(r *Runtime) {
		r.baseTickDuration = d
	}
}

func WithClock(clock clock.Clock) Option {
	return func(r *Runtime) {
		r.clock = clock
	}
}

func WithPayloadsMapping(pl map[string]event.Payload) Option {
	return func(r *Runtime) {
		if pl == nil {
			return
		}
		if r.eventMap != nil {
			for label := range r.eventMap {
				if _, ok := pl[label]; ok {
					panic("the payload corresponding to the event '" + label + "' has already been defined as an event")
				}
			}
		}
		r.payloadMap = pl
	}
}

func WithEventsMapping(ev map[string]event.Event) Option {
	return func(r *Runtime) {
		if ev == nil {
			return
		}
		if r.payloadMap != nil {
			for label, _ := range r.payloadMap {
				if _, ok := ev[label]; ok {
					panic("the event '" + label + "' has already been defined as a payload")
				}
			}
		}
		r.eventMap = ev
	}
}
