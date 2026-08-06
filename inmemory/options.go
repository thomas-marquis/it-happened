package inmemory

import "github.com/thomas-marquis/it-happened/event"

// BusOption is a function that configures an in-memory bus.
// Options use the functional options pattern for flexible bus configuration.
type BusOption func(b *inMemoryBus)

// WithWorkers sets the number of workers used to publish events.
//
// Keep in mind that, internally, one goroutine is started and run permanently for each worker.
// A higher number of workers allows unstacking more events from the buffer, improving
// throughput but increasing memory usage.
//
// Parameters:
//
//	nbr - The number of worker goroutines
//
// Returns:
//
//	A BusOption that configures the number of workers
//
// Default: 16
func WithWorkers(nbr int) BusOption {
	return func(b *inMemoryBus) {
		b.nbPubWorkers = nbr
	}
}

// WithNotifier sets a notifier to the event bus
//
// The notifier's Notify method will be called each time an event is published on the bus.
func WithNotifier(notifier event.Notifier) BusOption {
	return func(b *inMemoryBus) {
		b.notifier = notifier
	}
}

// WithReadiness sets an optional user-defined chanel that is closed by the bus when it is ready to publish/receive events.
func WithReadiness(ready chan struct{}) BusOption {
	return func(b *inMemoryBus) {
		b.ready = ready
	}
}
