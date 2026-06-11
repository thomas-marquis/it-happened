package timeline

import "time"

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
