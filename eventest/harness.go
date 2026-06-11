package eventest

import (
	"testing"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/clock"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/interceptor"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/timeline"
)

type Option func(*Harness)

func WithSideEffect(marble string) Option {
	return func(h *Harness) {
		h.sideEffect = marble
	}
}

func WithPayloads(payloads map[string]event.Payload) Option {
	return func(h *Harness) {
		h.payloadMap = payloads
	}
}

func WithMatchers(matchers map[string]event.Matcher) Option {
	return func(h *Harness) {
		h.matchers = matchers
	}
}

func WithEvents(events map[string]event.Event) Option {
	return func(h *Harness) {
		h.eventMap = events
	}
}

func WithTickDuration(d time.Duration) Option {
	return func(h *Harness) {
		h.tickDuration = d
	}
}

type Harness struct {
	bus          event.Bus
	expected     string
	sideEffect   string
	payloadMap   map[string]event.Payload
	eventMap     map[string]event.Event
	matchers     map[string]event.Matcher
	tickDuration time.Duration
}

func NewHarness(bus event.Bus, expected string, opts ...Option) *Harness {
	h := &Harness{
		bus:          bus,
		expected:     expected,
		tickDuration: timeline.DefaultTickDuration,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *Harness) Run(t *testing.T, f func(bus event.Bus, clock clock.Clock)) {
	clock := clock.NewClock()
	intercept := interceptor.NewInterceptor(t, h.bus, clock)

	recorder := intercept.EXPECT().FromMarble(h.expected)

	// Automatically add matchers from payload and event maps
	if h.payloadMap != nil {
		for label, pl := range h.payloadMap {
			recorder.ShouldMatch(map[string]event.Matcher{
				label: event.HasPayload(pl),
			})
		}
	}
	if h.eventMap != nil {
		for label, evt := range h.eventMap {
			recorder.ShouldMatch(map[string]event.Matcher{
				label: event.IsExactly(evt),
			})
		}
	}

	if h.matchers != nil {
		recorder.ShouldMatch(h.matchers)
	}

	if h.sideEffect != "" {
		rt := runtime.NewRuntime(intercept,
			runtime.WithClock(clock),
			runtime.WithPayloadsMapping(h.payloadMap),
			runtime.WithEventsMapping(h.eventMap),
			runtime.WithBaseTickDuration(h.tickDuration))

		sess, err := rt.Run(h.sideEffect)
		if err != nil {
			t.Fatalf("failed to parse side effect marble: %v", err)
		}

		// Run all side effects synchronously for now to establish state,
		// or we could run them tick by tick.
		// If we run them all here, the clock advances.
		for sess.HasNext() {
			if err := sess.Next(); err != nil {
				t.Fatalf("side effect failed: %v", err)
			}
		}
	}

	clock.Start()
	f(intercept, clock)
	clock.Stop()

	intercept.Finish()
}
