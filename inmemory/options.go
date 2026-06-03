package inmemory

type BusOption func(b *inMemoryBus)

// WithBufferSize sets the size of the buffer used to publish events.
// Published events are stacked nin the chanel. The bus starts to be blocking as the moment the buffer is full.
// Default: 100
func WithBufferSize(size int) BusOption {
	return func(b *inMemoryBus) {
		b.bufferSize = size
	}
}

// WithWorkers sets the number of workers used to publish events
// Keep in mind that, internally, one goroutine is started and run permanently for each worker.
// A higher number of workers allows unstacking more events from the buffer, improving throughput but increasing memory usage.
// Default: 16
func WithWorkers(nbr int) BusOption {
	return func(b *inMemoryBus) {
		b.nbPubWorkers = nbr
	}
}
