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

func TestInmemoryBus_Publish(t *testing.T) {
	t.Run("Given inmemory bus with registered subscriber, When event is published, Then subscriber receives the event", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var received event.Event
		var receivedMutex sync.Mutex
		var wg sync.WaitGroup
		wg.Add(1)

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			receivedMutex.Lock()
			received = evt
			receivedMutex.Unlock()
		})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		testEvent := event.New(testPayload("test"))

		// When
		bus.Publish(testEvent)

		// Then
		waitDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			receivedMutex.Lock()
			defer receivedMutex.Unlock()
			require.NotNil(t, received)
			assert.Equal(t, testEvent.ID(), received.ID())
			assert.Equal(t, testEvent.Type(), received.Type())
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for event")
		}

		close(done)
	})
}

func TestInmemoryBus_MultipleSubscribers(t *testing.T) {
	t.Run("Given inmemory bus with multiple subscribers, When event is published, Then all subscribers receive the event", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		// Create three subscribers
		var received1, received2, received3 event.Event
		var mutex1, mutex2, mutex3 sync.Mutex
		var wg sync.WaitGroup
		wg.Add(3)

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

		testEvent := event.New(testPayload("test"))

		// When
		bus.Publish(testEvent)

		// Then
		waitDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			// Verify all subscribers received the same event
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

		close(done)
	})
}

func TestInmemoryBus_ConcurrentPublish(t *testing.T) {
	t.Run("Given inmemory bus with concurrent publish calls, When multiple events are published simultaneously, Then all events are delivered correctly without data races", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var received []event.Event
		var mutex sync.Mutex

		numEvents := 100
		var wg sync.WaitGroup
		wg.Add(numEvents)

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			mutex.Lock()
			received = append(received, evt)
			mutex.Unlock()
			wg.Done()
		})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		// When
		for i := 0; i < numEvents; i++ {
			go func(idx int) {
				eventPayload := testPayload2{Value: "event"}
				evt := event.New(eventPayload)
				bus.Publish(evt)
			}(i)
		}

		// Then
		waitDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			mutex.Lock()
			defer mutex.Unlock()

			require.Len(t, received, numEvents)

			// Verify all events have unique IDs
			idSet := make(map[string]struct{})
			for _, evt := range received {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, numEvents, "all events should have unique IDs")
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}

		close(done)
	})
}

func TestInmemoryBus_EventMatching(t *testing.T) {
	t.Run("Given inmemory bus with subscribers using different matchers, When event is published, Then only subscribers with matching criteria receive the event", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		// Subscriber for type "test.payload"
		var received1 event.Event
		var mutex1 sync.Mutex
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
		var received2 event.Event
		var mutex2 sync.Mutex
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
		var received3 event.Event
		var mutex3 sync.Mutex
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

		// Create events of different types
		event1 := event.New(testPayload("test1"))
		event2 := event.New(testPayload2{Value: "test2"})

		// When - publish event1 (type: test.payload)
		bus.Publish(event1)

		// Then - verify event1 delivery
		waitDone1 := make(chan struct{})
		go func() {
			wg1.Wait()
			close(waitDone1)
		}()

		waitDone3a := make(chan struct{})
		go func() {
			wg3.Wait()
			close(waitDone3a)
		}()

		select {
		case <-waitDone1:
			mutex1.Lock()
			assert.Equal(t, event1.ID(), received1.ID())
			mutex1.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 1 to receive event1")
		}

		select {
		case <-waitDone3a:
			mutex3.Lock()
			assert.Equal(t, event1.ID(), received3.ID())
			mutex3.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 3 to receive event1")
		}

		// Verify subscriber2 did NOT receive event1
		time.Sleep(100 * time.Millisecond) // Give it time to potentially receive
		mutex2.Lock()
		assert.Equal(t, event.Event(nil), received2, "subscriber2 should not receive test.payload events")
		mutex2.Unlock()

		// Reset for event2 - create new wg3 since the old one is done
		wg3 = sync.WaitGroup{}
		wg3.Add(1)

		// When - publish event2 (type: test.payload.2)
		bus.Publish(event2)

		// Then - verify event2 delivery
		waitDone2 := make(chan struct{})
		go func() {
			wg2.Wait()
			close(waitDone2)
		}()

		waitDone3b := make(chan struct{})
		go func() {
			wg3.Wait()
			close(waitDone3b)
		}()

		select {
		case <-waitDone2:
			mutex2.Lock()
			assert.Equal(t, event2.ID(), received2.ID())
			mutex2.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 2 to receive event2")
		}

		select {
		case <-waitDone3b:
			mutex3.Lock()
			assert.Equal(t, event2.ID(), received3.ID())
			mutex3.Unlock()
		case <-time.After(time.Second):
			assert.Fail(t, "timeout waiting for subscriber 3 to receive event2")
		}

		// Verify subscriber1 did NOT receive event2
		time.Sleep(100 * time.Millisecond) // Give it time to potentially receive
		mutex1.Lock()
		assert.Equal(t, event1.ID(), received1.ID(), "subscriber1 should not receive test.payload.2 events, still has event1")
		mutex1.Unlock()

		close(done)
	})
}

func TestInmemoryBus_Subscribe(t *testing.T) {
	t.Run("Given inmemory bus, When Subscribe() is called, Then returns a valid Subscriber", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		// When
		sub := bus.Subscribe()

		// Then
		require.NotNil(t, sub)
		assert.NotNil(t, sub)

		close(done)
	})
}

func TestInmemoryBus_ThreadSafety(t *testing.T) {
	t.Run("Given inmemory bus with concurrent publish and subscribe operations, When operations execute simultaneously, Then no race conditions detected, all events delivered correctly", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		bus := inmemory.NewBus(done, &event.NopNotifier{})

		numPublishers := 10
		numEventsPerPublisher := 10
		totalEvents := numPublishers * numEventsPerPublisher

		var received []event.Event
		var mutex sync.Mutex
		var wg sync.WaitGroup
		wg.Add(totalEvents)

		sub := bus.Subscribe().On(event.IsAny(), func(evt event.Event) {
			mutex.Lock()
			received = append(received, evt)
			mutex.Unlock()
			wg.Done()
		})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		// When - concurrent publish and subscribe operations
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
		waitDone := make(chan struct{})
		go func() {
			wg.Wait()
			close(waitDone)
		}()

		select {
		case <-waitDone:
			mutex.Lock()
			defer mutex.Unlock()

			require.Len(t, received, totalEvents)

			// Verify all events have unique IDs
			idSet := make(map[string]struct{})
			for _, evt := range received {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, totalEvents, "all events should have unique IDs")
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}

		close(done)
	})
}
