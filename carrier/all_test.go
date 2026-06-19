package carrier_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

func TestAllCarrier_Dispatch(t *testing.T) {
	t.Run("should dispatch all events in parallel when carrier is published", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var receivedEvents []event.Event
		var mu sync.Mutex
		var wg sync.WaitGroup

		eventsToCarry := []event.Event{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
		}

		wg.Add(len(eventsToCarry))
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				mu.Lock()
				receivedEvents = append(receivedEvents, evt)
				mu.Unlock()
				wg.Done()
			})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		doneEvent := event.New(testPayload2{Value: "done"})
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewAll(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithMaxConcurrency(10),
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		doneCh := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCh)
		}()

		select {
		case <-doneCh:
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, receivedEvents, len(eventsToCarry))

			idSet := make(map[string]struct{})
			for _, evt := range receivedEvents {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, len(eventsToCarry))
			assert.Contains(t, idSet, eventsToCarry[0].ID())
			assert.Contains(t, idSet, eventsToCarry[1].ID())
			assert.Contains(t, idSet, eventsToCarry[2].ID())
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestAllCarrier_CompletionEvent(t *testing.T) {
	t.Run("should publish done event when all carried events are processed", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var doneReceived bool
		var mu sync.Mutex

		event1 := event.New(testPayload("event1"))
		event2 := event.New(testPayload("event2"))
		eventsToCarry := []event.Event{event1, event2}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if evt.ID() == event1.ID() || evt.ID() == event2.ID() {
					// Publish a followup event which will trigger completion
					bus.Publish(evt.NewFollowup(testPayload("followup")))
				}
				if evt.ID() == doneEvent.ID() {
					mu.Lock()
					doneReceived = true
					mu.Unlock()
				}
			})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		carrierEvent := carrier.NewAll(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithMaxConcurrency(10),
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		time.Sleep(200 * time.Millisecond)
		mu.Lock()
		assert.True(t, doneReceived, "done event should be published")
		mu.Unlock()
	})
}

func TestAllCarrier_Timeout(t *testing.T) {
	t.Run("should publish timeout event when processing exceeds timeout duration", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var timeoutReceived bool
		var mu sync.Mutex

		eventsToCarry := []event.Event{
			event.New(testPayload2{Value: "event1"}),
			event.New(testPayload2{Value: "event2"}),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewAll(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithTimeout(50*time.Millisecond),
			carrier.WithMaxConcurrency(10),
		)

		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if evt.ID() == timeoutEvent.ID() {
					mu.Lock()
					timeoutReceived = true
					mu.Unlock()
				}
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		bus.Publish(carrierEvent)

		// Then
		time.Sleep(200 * time.Millisecond)
		mu.Lock()
		assert.True(t, timeoutReceived, "timeout event should be published")
		mu.Unlock()
	})
}

func TestAllCarrier_ConcurrentProcessing(t *testing.T) {
	t.Run("should process events concurrently up to max concurrency", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		numEvents := 20
		maxConcurrency := 5

		var receivedOrder []int
		var mu sync.Mutex
		var wg sync.WaitGroup

		wg.Add(numEvents)
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				mu.Lock()
				payload := evt.Payload().(testPayload)
				index := int(payload[len(payload)-1] - '0')
				receivedOrder = append(receivedOrder, index)
				mu.Unlock()
				wg.Done()
			})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		var eventsToCarry []event.Event
		for i := 0; i < numEvents; i++ {
			eventsToCarry = append(eventsToCarry, event.New(testPayload(fmt.Sprintf("event%d", i))))
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewAll(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithMaxConcurrency(maxConcurrency),
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		doneCh := make(chan struct{})
		go func() {
			wg.Wait()
			close(doneCh)
		}()

		select {
		case <-doneCh:
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, receivedOrder, numEvents)
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestAllCarrier_EmptyEvents(t *testing.T) {
	t.Run("should handle empty events list gracefully", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var receivedCount int
		var mu sync.Mutex

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewAll(
			[]event.Event{},
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				mu.Lock()
				receivedCount++
				mu.Unlock()
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		bus.Publish(carrierEvent)

		// Then
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		assert.Equal(t, 1, receivedCount)
		mu.Unlock()
	})
}

func TestAllCarrier_EventType(t *testing.T) {
	t.Run("should have correct event type", func(t *testing.T) {
		// Given
		events := []event.Event{
			event.New(testPayload("event1")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When
		carrierEvent := carrier.NewAll(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		// Then
		require.NotNil(t, carrierEvent)
		assert.NotNil(t, carrierEvent.Payload())

		payload := carrierEvent.Payload().(*carrier.All)
		assert.Equal(t, event.Type(carrier.TypePrefix+".all"), payload.EventType())
	})
}
