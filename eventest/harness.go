package eventest

import (
	"testing"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

type Option func(*Harness)

func WithSideEffect(marble string) Option {
	return func(h *Harness) {

	}
}

func WithPayloads(payloads map[string]event.Payload) Option {
	return func(h *Harness) {}
}

func WithMatchers(matchers map[string]event.Matcher) Option {
	return func(h *Harness) {}
}

type Harness struct {
}

func NewHarness(bus event.Bus, expected string, opts ...Option) *Harness {
	h := &Harness{}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *Harness) Run(t *testing.T, f func()) {}

type fakePayload string

func (fakePayload) Type() event.Type { return "fake.payload" }

func test() {
	done := make(chan struct{})
	defer close(done)
	bus := inmemory.NewBus(done, nil) // TODO: pass notifier as an option

	exp := ""
	se := ""
	th := NewHarness(bus,
		exp,
		WithSideEffect(se),
		WithPayloads(map[string]event.Payload{}),
		WithMatchers(map[string]event.Matcher{}),
	)
	in := event.New(fakePayload("my value"))
	var t *testing.T
	th.Run(t, func() {
		bus.Publish(in)
	})
}
