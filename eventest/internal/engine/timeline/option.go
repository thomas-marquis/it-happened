package timeline

import (
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

type Option func(*Timeline)

func WithSeed(seed uint64) Option {
	return func(t *Timeline) {
		t.seed(seed)
	}
}

func WithTickDuration(d time.Duration) Option {
	return func(t *Timeline) {
		t.tickDuration = d
	}
}

func WithPlaceholderEvents(placeholders []event.Event) Option {
	return func(t *Timeline) {
		t.placeholderEvents = placeholders
	}
}
