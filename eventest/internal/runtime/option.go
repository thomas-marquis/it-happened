package runtime

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

type Option func(*Runtime)

func WithBaseTickDuration(d time.Duration) Option {
	return func(r *Runtime) {
		r.baseTickDuration = d
	}
}

func WithPayloadsMapping(pl map[string]event.Payload) Option {
	return func(r *Runtime) {
		r.payloadMap = pl
	}
}

type TimelineOption func(*Timeline)

func TimelineWithSeed(seed uint64) TimelineOption {
	return func(t *Timeline) {
		t.seed(seed)
	}
}

func TimelineWithTickDuration(d time.Duration) TimelineOption {
	return func(t *Timeline) {
		t.tickDuration = d
	}
}
