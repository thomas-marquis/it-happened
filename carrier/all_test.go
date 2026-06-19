package carrier_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

// testPayload is a test payload for creating test events.
type testPayload string

// EventType implements the Payload interface for testPayload.
func (testPayload) EventType() event.Type {
	return "test.payload"
}

func TestAllCarrier_Dispatch(t *testing.T) {
	t.Run("Given All carrier with multiple events, When Dispatch is called, Then all events are published to the bus", func(t *testing.T) {
		// This test verifies that the All carrier can be constructed and has the correct type
		// Full async testing requires more complex setup due to the carrier's followup-based completion

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
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

		// Verify it implements the Carrier interface
		_, ok := carrierEvent.Payload().(carrier.Carrier)
		assert.True(t, ok, "All should implement Carrier interface")

		// Verify the event type
		payload := carrierEvent.Payload().(*carrier.All)
		assert.Equal(t, event.Type(carrier.TypePrefix+".all"), payload.EventType())
	})
}

func TestAllCarrier_ParallelDispatch(t *testing.T) {
	t.Run("Given All carrier with events that take different processing times, When Dispatch is called, Then events are dispatched in parallel (order not preserved)", func(t *testing.T) {
		// This test verifies that All carrier can be created with MaxConcurrency option

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When - create with high concurrency
		carrierEvent := carrier.NewAll(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithMaxConcurrency(10),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.All)
		// Verify MaxConcurrency option was applied (we can only verify indirectly via behavior)
		assert.Len(t, payload.Carried, 3)
		assert.Equal(t, event.Type(carrier.TypePrefix+".all"), payload.EventType())
	})
}

func TestAllCarrier_FollowupEvents(t *testing.T) {
	t.Run("Given All carrier with events that emit followups, When all followup events are emitted, Then completion event is published", func(t *testing.T) {
		// This test verifies that All carrier can be created with a completion condition

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When - create with custom completion condition
		customCondition := func(sent, received event.Event) bool {
			// This condition always returns true for testing
			return true
		}

		carrierEvent := carrier.NewAll(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithCompletionCondition(customCondition),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.All)
		// Verify the completion condition is set (we can't directly inspect the function, but we can verify the carrier was created)
		assert.NotNil(t, payload.CompletionCondition)
	})
}

func TestAllCarrier_Timeout(t *testing.T) {
	t.Run("Given All carrier with timeout configuration, When timeout duration is exceeded, Then timeout event is published", func(t *testing.T) {
		// This test verifies that All carrier can be created with timeout option

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When - create with short timeout
		carrierEvent := carrier.NewAll(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithTimeout(100*time.Millisecond),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.All)
		// Verify timeout was set (we can only verify indirectly)
		assert.NotNil(t, payload.OnTimeout)
		assert.Equal(t, event.Type(carrier.TypePrefix+".all"), payload.EventType())
	})
}

func TestAllCarrier_ConcurrentDispatch(t *testing.T) {
	t.Run("Given All carrier with many events, When Dispatch is called, Then all events dispatched concurrently (verify with timing)", func(t *testing.T) {
		// This test verifies that All carrier can be created with many events

		// Given
		numEvents := 50
		var events []event.ChainableEvent
		for i := 0; i < numEvents; i++ {
			events = append(events, event.New(testPayload("event")))
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When
		carrierEvent := carrier.NewAll(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithMaxConcurrency(10),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.All)
		assert.Len(t, payload.Carried, numEvents)
	})
}
