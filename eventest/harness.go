package eventest

import (
	"testing"

	"github.com/thomas-marquis/it-happened/event"
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
