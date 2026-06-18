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

type publishedLoad struct {
	evt              event.Event
	subscriberChanel chan event.Event
}

type inMemoryBus struct {
	sync.Mutex

	subscribers    map[chan event.Event]*event.Subscriber
	publishingChan chan publishedLoad
	done           <-chan struct{}
	notifier       event.Notifier
	wg             sync.WaitGroup

	bufferSize, nbPubWorkers int
}

// NewBus creates a new in-memory event bus.
// This implementation allows blocking carrier Dispatch method.
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

func (b *inMemoryBus) Subscribe() *event.Subscriber {
	b.Lock()
	defer b.Unlock()

	events := make(chan event.Event)
	subscriber := event.NewSubscriber(events)
	b.subscribers[events] = subscriber
	return subscriber
}

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
