package event

import "context"

// Option is a function that configures an event implementation.
// Options use the functional options pattern to allow flexible event configuration.
type Option func(e *impl)

// WithContext sets the context for the event.
//
// This option allows associating a context with the event, which can be used
// for cancellation, deadlines, or passing request-scoped values.
//
// Parameters:
//
//	ctx - The context to associate with the event
//
// Returns:
//
//	An Option that configures the event's context
func WithContext(ctx context.Context) Option {
	return func(e *impl) {
		e.ctx = ctx
	}
}

// WithRef sets the chain reference for the event.
//
// This option allows specifying a custom ChainRef, which is useful when
// creating an event that should be part of an existing chain.
//
// Parameters:
//
//	ref - The chain reference to use for the event
//
// Returns:
//
//	An Option that configures the event's chain reference
func WithRef(ref string) Option {
	return func(e *impl) {
		e.ref = ref
	}
}
