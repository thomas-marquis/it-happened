package inmemory

import (
	"sync"

	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

const (
	// publicationWorkers defines the number of concurrent worker goroutines responsible for managing app events.
	publicationWorkers = 16

	// pubChanBufferSize defines the size of the channel used to publish events.
	// Increase this value to manage more subscribers without blocking event publishing.
	pubChanBufferSize = 100
)

// publishedLoad represents a publication task for a worker.
type publishedLoad struct {
	evt              event.Event
	subscriberChanel chan event.Event
}

// inMemoryBus is an in-memory implementation of the event.Bus interface.
// It manages event publishing and subscription with concurrent worker goroutines.
type inMemoryBus struct {
	sync.Mutex

	// subscribers maps event channels to their corresponding subscriber instances.
	subscribers map[chan event.Event]*event.Subscriber
	// publishingChan is the channel through which publication tasks are sent to workers.
	publishingChan chan publishedLoad
	// done is the channel that signals the bus to shut down.
	done <-chan struct{}
	// notifier is used to notify about published events.
	notifier event.Notifier
	wg       sync.WaitGroup

	bufferSize, nbPubWorkers int
}

// NewBus creates a new in-memory event bus.
//
// This implementation allows blocking carrier Dispatch method.
// The bus uses worker goroutines to handle event publishing concurrently.
//
// Parameters:
//
//	done - Channel that signals when the bus should shut down
//	notifier - Optional notifier for published events (defaults to NopNotifier)
//	opts - Optional configuration options for the bus
//
// Returns:
//
//	A new in-memory event Bus instance
func NewBus(done <-chan struct{}, notifier event.Notifier, opts ...BusOption) event.Bus {
	b := &inMemoryBus{
		subscribers:  make(map[chan event.Event]*event.Subscriber),
		done:         done,
		bufferSize:   pubChanBufferSize,
		nbPubWorkers: publicationWorkers,
	}

	for _, opt := range opts {
		opt(b)
	}

	b.publishingChan = make(chan publishedLoad, b.bufferSize)

	if notifier != nil {
		b.notifier = notifier
	} else {
		b.notifier = &event.NopNotifier{}
	}

	for i := 0; i < b.nbPubWorkers; i++ {
		b.wg.Add(1)
		go b.pubWorker()
	}

	go b.terminate()

	return b
}

// Subscribe creates a new subscriber and returns it.
//
// The subscriber will receive events through its own channel.
//
// Returns:
//
//	A new Subscriber instance
func (b *inMemoryBus) Subscribe() *event.Subscriber {
	b.Lock()
	defer b.Unlock()

	events := make(chan event.Event)
	subscriber := event.NewSubscriber(events)
	b.subscribers[events] = subscriber
	return subscriber
}

// Publish publishes an event to all subscribers.
//
// If the event payload implements the Carrier interface, it dispatches the carrier's events
// asynchronously. Otherwise, it sends the event to all matching subscribers.
//
// Parameters:
//
//	evt - The event to publish
func (b *inMemoryBus) Publish(evt event.Event) {
	b.notifier.Notify(evt)
	if c, ok := evt.Payload().(carrier.Carrier); ok {
		go c.Dispatch(b)
		return
	}

	b.Lock()
	defer b.Unlock()

	for channel, subscriber := range b.subscribers {
		if !subscriber.Accept(evt) {
			continue
		}
		select {
		case b.publishingChan <- publishedLoad{evt, channel}:
		case <-b.done:
		}
	}
}

// pubWorker is a worker goroutine that processes publication tasks.
// It continuously receives publication tasks from the publishing channel and
// forwards events to subscriber channels.
func (b *inMemoryBus) pubWorker() {
	defer b.wg.Done()
	for {
		select {
		case <-b.done:
			return
		case i := <-b.publishingChan:
			select {
			case i.subscriberChanel <- i.evt:
			case <-b.done:
			}
		}
	}
}

// terminate handles the shutdown of the bus.
// It waits for all workers to finish and closes all subscriber channels.
func (b *inMemoryBus) terminate() {
	<-b.done
	b.wg.Wait()
	b.Lock()
	defer b.Unlock()
	for subChanel := range b.subscribers {
		close(subChanel)
	}
	clear(b.subscribers)
}
