package inmemory

import (
	"container/heap"
	"context"
	"sync"

	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

const (
	// defaultWorkers defines the default number of concurrent worker goroutines responsible for managing app events.
	defaultWorkers = 16
)

// inMemoryBus is an in-memory implementation of the event.Bus interface.
// It manages event publishing and subscription with concurrent worker goroutines.
type inMemoryBus struct {
	subMu sync.RWMutex
	pubMu sync.Mutex

	// subscribers maps event channels to their corresponding subscriber instances.
	subscribers map[chan event.Event]*event.Subscriber

	ctx context.Context
	// notifier is used to notify about published events.
	notifier event.Notifier
	wg       sync.WaitGroup

	queue           eventQueue
	publishedEvents chan event.Event
	pubSignal       chan struct{}
	nbPubWorkers    int
}

// NewBus creates a new in-memory event bus.
//
// This implementation allows blocking carrier Dispatch method.
// The bus uses worker goroutines to handle event publishing concurrently.
//
// Parameters:
//
//	ctx - A context the bus will use for cancellation
//	notifier - Optional notifier for published events (defaults to NopNotifier)
//	opts - Optional configuration options for the bus
//
// Returns:
//
//	A new in-memory event Bus instance
func NewBus(ctx context.Context, opts ...BusOption) event.Bus {
	queue := eventQueue{}
	heap.Init(&queue)

	b := &inMemoryBus{
		subscribers:     make(map[chan event.Event]*event.Subscriber),
		ctx:             ctx,
		notifier:        &event.NopNotifier{},
		queue:           queue,
		publishedEvents: make(chan event.Event),
		pubSignal:       make(chan struct{}),
		nbPubWorkers:    defaultWorkers,
	}

	for _, opt := range opts {
		opt(b)
	}

	for i := 0; i < b.nbPubWorkers; i++ {
		b.wg.Add(1)
		go b.worker()
	}

	go b.publisher()
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
func (b *inMemoryBus) Subscribe(defaultMatchers ...event.Matcher) *event.Subscriber {
	b.subMu.Lock()
	defer b.subMu.Unlock()

	events := make(chan event.Event)
	subscriber := event.NewSubscriber(events, defaultMatchers...)
	b.subscribers[events] = subscriber
	b.notifier.NotifySubscribed(subscriber)
	return subscriber
}

func (b *inMemoryBus) Unsubscribe(sub *event.Subscriber) {
	b.subMu.Lock()
	defer b.subMu.Unlock()

	if !sub.Detached() {
		sub.Detach()
	}

	for channel, subscriber := range b.subscribers {
		if subscriber == sub {
			delete(b.subscribers, channel)
			break
		}
	}
	b.notifier.NotifyUnsubscribed(sub)
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
	b.notifier.NotifyPublished(evt)
	if c, ok := evt.Payload().(carrier.Carrier); ok {
		go c.Dispatch(b)
		return
	}

	b.pubMu.Lock()
	heap.Push(&b.queue, evt)
	b.pubMu.Unlock()

	select {
	case b.pubSignal <- struct{}{}:
	default:
	}
}

func (b *inMemoryBus) publisher() {
	defer close(b.publishedEvents)
	for {
		select {
		case <-b.ctx.Done():
			return
		default:
		}

		b.pubMu.Lock()
		if len(b.queue) > 0 {
			next := heap.Pop(&b.queue).(event.Event)
			b.pubMu.Unlock()
			select {
			case b.publishedEvents <- next:
			case <-b.ctx.Done():
				return
			}
			continue
		}
		b.pubMu.Unlock()

		select {
		case <-b.pubSignal:
		case <-b.ctx.Done():
			return
		}
	}
}

func (b *inMemoryBus) worker() {
	defer b.wg.Done()
	for {
		select {
		case <-b.ctx.Done():
			return
		case evt, ok := <-b.publishedEvents:
			if !ok {
				return
			}

			b.subMu.RLock()
			//localSubs := make(map[chan event.Event]*event.Subscriber)
			//for ch, sub := range b.subscribers {
			//	localSubs[ch] = sub
			//}
			//b.subMu.Unlock()

			for channel, subscriber := range b.subscribers {
				if !subscriber.Accept(evt) {
					continue
				}
				select {
				case channel <- evt:
				case <-b.ctx.Done():
				}
			}
			b.subMu.RUnlock()
		}
	}
}

// terminate handles the shutdown of the bus.
// It waits for all workers to finish and closes all subscriber channels.
func (b *inMemoryBus) terminate() {
	<-b.ctx.Done()
	b.wg.Wait()
	b.subMu.Lock()
	defer b.subMu.Unlock()
	for subChanel := range b.subscribers {
		close(subChanel)
	}
	clear(b.subscribers)
}

type eventQueue []event.Event

func (q eventQueue) Len() int {
	return len(q)
}

func (q eventQueue) Less(i, j int) bool {
	return q[i].Priority() > q[j].Priority()
}

func (q eventQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
}

func (q *eventQueue) Push(e any) {
	*q = append(*q, e.(event.Event))
}

func (q *eventQueue) Pop() any {
	old := *q
	n := len(old)
	lastEvt := old[n-1]
	old[n-1] = nil
	*q = old[:n-1]
	return lastEvt
}
