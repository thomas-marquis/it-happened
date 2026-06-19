package inmemory

// BusOption is a function that configures an in-memory bus.
// Options use the functional options pattern for flexible bus configuration.
type BusOption func(b *inMemoryBus)

// WithBufferSize sets the size of the buffer used to publish events.
//
// Published events are stacked in the channel. The bus starts to be blocking at the moment
// the buffer is full.
//
// Parameters:
//
//	size - The buffer size for event publishing
//
// Returns:
//
//	A BusOption that configures the buffer size
//
// Default: 100
func WithBufferSize(size int) BusOption {
	return func(b *inMemoryBus) {
		b.bufferSize = size
	}
}

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
