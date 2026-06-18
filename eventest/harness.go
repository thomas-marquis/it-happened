package eventest

import (
	"testing"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/clock"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/interceptor"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/timeline"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
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

func (h *Harness) RunAndWait(t *testing.T, initEvent event.Event) {
	clk := clock.NewClock()
	intercept := interceptor.New(t, h.bus, clk)

	expectedNode, err := marble.ParseAsNode(h.expected)
	if err != nil {
		t.Fatalf("failed to parse expectation marble: %v", err)
	}

	if err := marble.Validate(expectedNode,
		marble.MandatoryInitEventRule{},
		marble.WaitlessGroupsRule{}); err != nil {
		t.Fatalf("expectation validation failed: %v", err)
	}

	recorder := intercept.EXPECT().FromMarble(h.expected)

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

	expectedTimeline := timeline.New(expectedNode, timeline.WithTickDuration(h.tickDuration)) // ???
	expectedTicks := len(expectedTimeline.Ticks())

	if h.sideEffect != "" {
		sideEffectNode, err := marble.ParseAsNode(h.sideEffect)
		if err != nil {
			t.Fatalf("failed to parse side effect marble: %v", err)
		}

		if err := marble.Validate(sideEffectNode,
			marble.NoInitEventInSideEffectRule{},
			marble.WaitlessGroupsRule{}); err != nil {
			t.Fatalf("side effect validation failed: %v", err)
		}

		sideEffectTimeline := timeline.New(sideEffectNode, timeline.WithTickDuration(h.tickDuration))
		sideEffectTicks := len(sideEffectTimeline.Ticks())

		if sideEffectTicks > expectedTicks {
			t.Fatalf("side effect duration (%d ticks) exceeds expectation duration (%d ticks)", sideEffectTicks, expectedTicks)
		}

		rt := runtime.New(intercept,
			runtime.WithClock(clk),
			runtime.WithPayloadsMapping(h.payloadMap),
			runtime.WithEventsMapping(h.eventMap),
			runtime.WithBaseTickDuration(h.tickDuration))

		if err := rt.RunAll(h.sideEffect); err != nil {
			t.Fatalf("failed to run side effect: %v", err)
		}
		clk.Stop()

	} else {
		// No side effect: start clock and wait for full expectation duration
		clk.Start()
		defer clk.Stop()

		// Calculate total duration from timeline
		totalDuration := time.Duration(0)
		for _, tick := range expectedTimeline.Ticks() {
			totalDuration += tick.Duration
		}

		time.Sleep(totalDuration)
	}

	intercept.Finish()
}
