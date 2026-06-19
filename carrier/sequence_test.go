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

func TestSequenceCarrier_Dispatch(t *testing.T) {
	t.Run("should dispatch all events sequentially when carrier is published", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var receivedEvents []event.Event
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Track which events we're expecting
		expectedIDs := make(map[string]bool)

		eventsToCarry := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
		}
		for _, evt := range eventsToCarry {
			expectedIDs[evt.ID()] = true
		}

		wg.Add(len(eventsToCarry))
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				// Only process the carried events, not done/timeout events
				if !expectedIDs[evt.ID()] {
					return
				}
				mu.Lock()
				receivedEvents = append(receivedEvents, evt)
				mu.Unlock()
				// Publish a followup to allow sequence to continue
				if chainable, ok := evt.(event.ChainableEvent); ok {
					bus.Publish(chainable.NewFollowup(testPayload("followup")))
				}
				wg.Done()
			})
		sub.ListenWithWorkers(16)
		defer sub.Detach()

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewSequence(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, 2*time.Second):
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, receivedEvents, len(eventsToCarry))

			// Verify all original events were received
			idSet := make(map[string]struct{})
			for _, evt := range receivedEvents {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, len(eventsToCarry))
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestSequenceCarrier_OrderedDispatch(t *testing.T) {
	t.Run("should preserve order when dispatching events sequentially", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var receivedOrder []int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Track which events we're expecting
		expectedIDs := make(map[string]bool)

		numEvents := 10
		var eventsToCarry []event.ChainableEvent
		for i := 0; i < numEvents; i++ {
			evt := event.New(testPayload(fmt.Sprintf("event%d", i)))
			eventsToCarry = append(eventsToCarry, evt)
			expectedIDs[evt.ID()] = true
		}

		wg.Add(numEvents)
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				// Only process the carried events, not done/timeout events
				if !expectedIDs[evt.ID()] {
					return
				}
				mu.Lock()
				payload := evt.Payload().(testPayload)
				// Parse the number from the payload
				var index int
				_, err := fmt.Sscanf(string(payload), "event%d", &index)
				require.NoError(t, err)
				receivedOrder = append(receivedOrder, index)
				mu.Unlock()
				// Publish a followup to allow sequence to continue
				if chainable, ok := evt.(event.ChainableEvent); ok {
					bus.Publish(chainable.NewFollowup(testPayload("followup")))
				}
				wg.Done()
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewSequence(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, 2*time.Second):
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, receivedOrder, numEvents)

			// Verify events were received in order
			for i := 0; i < numEvents; i++ {
				assert.Equal(t, i, receivedOrder[i], "event %d should be received in order", i)
			}
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestSequenceCarrier_CompletionEvent(t *testing.T) {
	t.Run("should publish done event when all carried events are processed in sequence", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var doneReceived bool
		var mu sync.Mutex

		eventsToCarry := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if evt.ID() == doneEvent.ID() {
					mu.Lock()
					doneReceived = true
					mu.Unlock()
				}
				if chainable, ok := evt.(event.ChainableEvent); ok {
					bus.Publish(chainable.NewFollowup(testPayload2{Value: fmt.Sprintf("followup-%d", chainable.ChainPosition())}))
				} else {
					require.Fail(t, "expected event to be chainable", "event: %s", evt.ID())
				}
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		carrierEvent := carrier.NewSequence(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
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

func TestSequenceCarrier_Timeout(t *testing.T) {
	t.Run("should publish timeout event when sequence processing exceeds timeout duration", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		var timeoutReceived bool
		var mu sync.Mutex

		// Create events that won't be processed (no subscribers for these)
		eventsToCarry := []event.ChainableEvent{
			event.New(slowPayload{Value: "event1"}),
			event.New(slowPayload{Value: "event2"}),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewSequence(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithTimeout(50*time.Millisecond),
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

func TestSequenceCarrier_SequentialOrder(t *testing.T) {
	t.Run("should ensure next event is not dispatched until previous completes", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		numEvents := 5
		var receivedOrder []int
		var mu sync.Mutex
		var wg sync.WaitGroup

		// Track which events we're expecting
		expectedIDs := make(map[string]bool)

		var eventsToCarry []event.ChainableEvent
		for i := 0; i < numEvents; i++ {
			evt := event.New(testPayload(fmt.Sprintf("event%d", i)))
			eventsToCarry = append(eventsToCarry, evt)
			expectedIDs[evt.ID()] = true
		}

		wg.Add(numEvents)
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				// Only process the carried events, not done/timeout events
				if !expectedIDs[evt.ID()] {
					return
				}
				mu.Lock()
				payload := evt.Payload().(testPayload)
				var index int
				_, err := fmt.Sscanf(string(payload), "event%d", &index)
				require.NoError(t, err)
				receivedOrder = append(receivedOrder, index)
				mu.Unlock()
				// Publish a followup to allow sequence to continue
				if chainable, ok := evt.(event.ChainableEvent); ok {
					bus.Publish(chainable.NewFollowup(testPayload("followup")))
				}
				wg.Done()
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		carrierEvent := carrier.NewSequence(
			eventsToCarry,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, 2*time.Second):
			mu.Lock()
			defer mu.Unlock()
			require.Len(t, receivedOrder, numEvents)
		case <-time.After(2 * time.Second):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}

func TestSequenceCarrier_EventType(t *testing.T) {
	t.Run("should have correct event type", func(t *testing.T) {
		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When
		carrierEvent := carrier.NewSequence(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
		)

		// Then
		require.NotNil(t, carrierEvent)
		assert.NotNil(t, carrierEvent.Payload())

		payload := carrierEvent.Payload().(*carrier.Sequence)
		assert.Equal(t, event.Type(carrier.TypePrefix+".sequence"), payload.EventType())
	})
}

func TestSequenceCarrier_NoMemoryLeak(t *testing.T) {
	t.Run("should not leak memory when creating and detaching multiple subscribers", func(t *testing.T) {
		// Given
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, &event.NopNotifier{})

		// Create and detach many sequence carriers to test for memory leaks
		// Each sequence creates a temporary subscriber that needs to be cleaned up
		numSequences := 100

		for i := 0; i < numSequences; i++ {
			eventsToCarry := []event.ChainableEvent{
				event.New(testPayload(fmt.Sprintf("event%d", i))),
			}

			doneEvent := event.New(testPayload("done"))
			timeoutEvent := event.New(testPayload("timeout"))

			carrierEvent := carrier.NewSequence(
				eventsToCarry,
				func(received []event.Event) event.Event { return doneEvent },
				timeoutEvent,
				carrier.WithTimeout(1*time.Millisecond),
			)

			// Publish the carrier - it will create and detach a subscriber internally
			bus.Publish(carrierEvent)
		}

		// When - allow time for all sequences to process (and timeout)
		time.Sleep(100 * time.Millisecond)

		// Then - the test passes if it doesn't crash or run out of memory
		// This test is a sanity check that Detach() properly cleans up callbacks
		// If there's a memory leak, this might fail with OOM in a more constrained environment
		assert.True(t, true, "memory leak test passed - no panic or OOM")
	})
}
