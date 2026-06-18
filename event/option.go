package event

import "context"

type Option func(e *impl)

func WithContext(ctx context.Context) Option {
	return func(e *impl) {
		e.ctx = ctx
	}
}

func WithRef(ref string) Option {
	return func(e *impl) {
		e.ref = ref
	}
}
