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

func (h *Harness) PublishAndWait(t *testing.T, placeholders ...event.Event) {
	clk := clock.NewClock()
	intercept := interceptor.NewInterceptor(t, h.bus, clk)

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
			runtime.WithClock(clk),
			runtime.WithPlaceholderEvents(placeholders),
			runtime.WithPayloadsMapping(h.payloadMap),
			runtime.WithEventsMapping(h.eventMap),
			runtime.WithBaseTickDuration(h.tickDuration))

		if err := rt.RunAll(h.sideEffect); err != nil {
			t.Fatalf("failed to parse side effect marble: %v", err)
		}
		clk.Stop()

	} else {
		clk.Start()
		defer clk.Stop()
		if len(placeholders) > 1 {
			t.Fatalf("specify a side effect to publish multiple placeholder events")
		}
		if len(placeholders) == 0 {
			t.Fatalf("specify a side effect to publish a placeholder event")
		}
		intercept.Publish(placeholders[0])
	}

	intercept.Finish()
}
