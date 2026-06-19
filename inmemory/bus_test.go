package inmemory_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

// testPayload is a test payload for creating test events.
type testPayload string

// EventType implements the Payload interface for testPayload.
func (testPayload) EventType() event.Type {
	return "test.payload"
}

// testPayload2 is another test payload for creating test events.
type testPayload2 struct {
	Value string
}

// EventType implements the Payload interface for testPayload2.
func (testPayload2) EventType() event.Type {
	return "test.payload.2"
}

// setupBus creates a new bus with a done channel that will be closed when the test completes.
// t.Helper() is called to mark this as a helper function.
func setupBus(t *testing.T) (func(), event.Bus) {
	t.Helper()
	done := make(chan struct{})
	bus := inmemory.NewBus(done, &event.NopNotifier{})
	return func() { close(done) }, bus
}

// setupSubscriber creates a subscriber on the given bus that collects received events.
// t.Helper() is called to mark this as a helper function.
func setupSubscriber(t *testing.T, bus event.Bus, matcher event.Matcher, workers int) (*event.Subscriber, *[]event.Event, *sync.Mutex, *sync.WaitGroup) {
	t.Helper()
	var received []event.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	sub := bus.Subscribe().On(matcher, func(evt event.Event) {
		mu.Lock()
		received = append(received, evt)
		mu.Unlock()
		wg.Done()
	})
	sub.ListenWithWorkers(workers)

	return sub, &received, &mu, &wg
}

// waitForEvents waits for the waitgroup and returns the received events.
// t.Helper() is called to mark this as a helper function.
func waitForEvents(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) chan struct{} {
	t.Helper()
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	return doneCh
}

func TestInmemoryBus_Publish(t *testing.T) {
	t.Run("should deliver published event to subscriber", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		wg := sync.WaitGroup{}
		wg.Add(1)

		var received event.Event
		var mu sync.Mutex

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			mu.Lock()
			received = evt
			mu.Unlock()
		})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		testEvent := event.New(testPayload("test"))

		// When
		bus.Publish(testEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, time.Second):
			mu.Lock()
			defer mu.Unlock()
			require.NotNil(t, received)
			assert.Equal(t, testEvent.ID(), received.ID())
			assert.Equal(t, testEvent.Type(), received.Type())
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for event")
		}
	})
}

func TestInmemoryBus_MultipleSubscribers(t *testing.T) {
	t.Run("should deliver published event to all subscribers", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		testEvent := event.New(testPayload("test"))

		var wg sync.WaitGroup
		wg.Add(3)

		var received1, received2, received3 event.Event
		var mutex1, mutex2, mutex3 sync.Mutex

		sub1 := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			mutex1.Lock()
			received1 = evt
			mutex1.Unlock()
		})
		sub1.ListenWithWorkers(1)
		defer sub1.Detach()

		sub2 := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			mutex2.Lock()
			received2 = evt
			mutex2.Unlock()
		})
		sub2.ListenWithWorkers(1)
		defer sub2.Detach()

		sub3 := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			mutex3.Lock()
			received3 = evt
			mutex3.Unlock()
		})
		sub3.ListenWithWorkers(1)
		defer sub3.Detach()

		// When
		bus.Publish(testEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, time.Second):
			mutex1.Lock()
			mutex2.Lock()
			mutex3.Lock()
			defer mutex1.Unlock()
			defer mutex2.Unlock()
			defer mutex3.Unlock()

			assert.Equal(t, testEvent.ID(), received1.ID())
			assert.Equal(t, testEvent.ID(), received2.ID())
			assert.Equal(t, testEvent.ID(), received3.ID())
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for all subscribers")
		}
	})
}

func TestInmemoryBus_ConcurrentPublish(t *testing.T) {
	t.Run("should handle concurrent publish without data races", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		var received []event.Event
		var mu sync.Mutex

		numEvents := 100
		var wg sync.WaitGroup
		wg.Add(numEvents)

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			mu.Lock()
			received = append(received, evt)
			mu.Unlock()
			wg.Done()
		})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		// When
		for i := 0; i < numEvents; i++ {
			go func(idx int) {
				evt := event.New(testPayload2{Value: "event"})
				bus.Publish(evt)
			}(i)
		}

		// Then
		select {
		case <-waitForEvents(t, &wg, 2*time.Second):
			mu.Lock()
			defer mu.Unlock()

			require.Len(t, received, numEvents)

			idSet := make(map[string]struct{})
			for _, evt := range received {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, numEvents, "all events should have unique IDs")
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestInmemoryBus_EventMatching(t *testing.T) {
	t.Run("should deliver events only to subscribers with matching criteria", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		var received1, received2, received3 event.Event
		var mutex1, mutex2, mutex3 sync.Mutex

		// Subscriber for type "test.payload"
		var wg1 sync.WaitGroup
		wg1.Add(1)
		sub1 := bus.Subscribe().On(event.Is("test.payload"), func(evt event.Event) {
			defer wg1.Done()
			mutex1.Lock()
			received1 = evt
			mutex1.Unlock()
		})
		sub1.ListenWithWorkers(1)
		defer sub1.Detach()

		// Subscriber for type "test.payload.2"
		var wg2 sync.WaitGroup
		wg2.Add(1)
		sub2 := bus.Subscribe().On(event.Is("test.payload.2"), func(evt event.Event) {
			defer wg2.Done()
			mutex2.Lock()
			received2 = evt
			mutex2.Unlock()
		})
		sub2.ListenWithWorkers(1)
		defer sub2.Detach()

		// Subscriber for all events
		var wg3 sync.WaitGroup
		wg3.Add(1)
		sub3 := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg3.Done()
			mutex3.Lock()
			received3 = evt
			mutex3.Unlock()
		})
		sub3.ListenWithWorkers(1)
		defer sub3.Detach()

		event1 := event.New(testPayload("test1"))
		event2 := event.New(testPayload2{Value: "test2"})

		// When - publish event1 (type: test.payload)
		bus.Publish(event1)

		// Then - verify event1 delivery
		select {
		case <-waitForEvents(t, &wg1, time.Second):
			mutex1.Lock()
			assert.Equal(t, event1.ID(), received1.ID())
			mutex1.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 1")
		}

		select {
		case <-waitForEvents(t, &wg3, time.Second):
			mutex3.Lock()
			assert.Equal(t, event1.ID(), received3.ID())
			mutex3.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 3")
		}

		// Verify subscriber2 did NOT receive event1
		time.Sleep(100 * time.Millisecond)
		mutex2.Lock()
		assert.Nil(t, received2, "subscriber2 should not receive test.payload events")
		mutex2.Unlock()

		// When - publish event2 (type: test.payload.2)
		wg3 = sync.WaitGroup{}
		wg3.Add(1)
		bus.Publish(event2)

		// Then - verify event2 delivery
		select {
		case <-waitForEvents(t, &wg2, time.Second):
			mutex2.Lock()
			assert.Equal(t, event2.ID(), received2.ID())
			mutex2.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 2")
		}

		select {
		case <-waitForEvents(t, &wg3, time.Second):
			mutex3.Lock()
			assert.Equal(t, event2.ID(), received3.ID())
			mutex3.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 3")
		}

		// Verify subscriber1 did NOT receive event2
		time.Sleep(100 * time.Millisecond)
		mutex1.Lock()
		assert.Equal(t, event1.ID(), received1.ID(), "subscriber1 should still have event1")
		mutex1.Unlock()
	})
}

func TestInmemoryBus_Subscribe(t *testing.T) {
	t.Run("should return a valid subscriber", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		// When
		sub := bus.Subscribe()

		// Then
		require.NotNil(t, sub)
		assert.NotNil(t, sub)
	})
}

func TestInmemoryBus_ThreadSafety(t *testing.T) {
	t.Run("should handle concurrent publish and subscribe without race conditions", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		numPublishers := 10
		numEventsPerPublisher := 10
		totalEvents := numPublishers * numEventsPerPublisher

		var received []event.Event
		var mu sync.Mutex
		var wg sync.WaitGroup
		wg.Add(totalEvents)

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			mu.Lock()
			received = append(received, evt)
			mu.Unlock()
			wg.Done()
		})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		var opWg sync.WaitGroup

		// Publishers
		for i := 0; i < numPublishers; i++ {
			opWg.Add(1)
			go func(publisherID int) {
				defer opWg.Done()
				for j := 0; j < numEventsPerPublisher; j++ {
					payload := testPayload2{Value: "event"}
					evt := event.New(payload)
					bus.Publish(evt)
				}
			}(i)
		}

		// Concurrent subscribers
		for i := 0; i < 5; i++ {
			opWg.Add(1)
			go func() {
				defer opWg.Done()
				_ = bus.Subscribe()
			}()
		}

		opWg.Wait()

		// Then
		select {
		case <-waitForEvents(t, &wg, 2*time.Second):
			mu.Lock()
			defer mu.Unlock()

			require.Len(t, received, totalEvents)

			idSet := make(map[string]struct{})
			for _, evt := range received {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, totalEvents, "all events should have unique IDs")
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}
