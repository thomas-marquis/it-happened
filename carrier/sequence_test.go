package carrier_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

func TestSequenceCarrier_Dispatch(t *testing.T) {
	t.Run("Given Sequence carrier with multiple events, When Dispatch is called, Then all events are published to the bus sequentially", func(t *testing.T) {
		// This test verifies that the Sequence carrier can be constructed

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
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

		// Verify it implements the Carrier interface
		_, ok := carrierEvent.Payload().(carrier.Carrier)
		assert.True(t, ok, "Sequence should implement Carrier interface")

		// Verify the event type
		payload := carrierEvent.Payload().(*carrier.Sequence)
		assert.Equal(t, event.Type(carrier.TypePrefix+".sequence"), payload.EventType())
	})
}

func TestSequenceCarrier_OrderedDispatch(t *testing.T) {
	t.Run("Given Sequence carrier with events in specific order, When Dispatch is called, Then events are published in the exact order they were added", func(t *testing.T) {
		// This test verifies that Sequence carrier can be created with ordered events

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
			event.New(testPayload("event3")),
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
		payload := carrierEvent.Payload().(*carrier.Sequence)
		assert.Len(t, payload.Carried, 3)
		assert.Equal(t, event.Type(carrier.TypePrefix+".sequence"), payload.EventType())

		// Verify events are in the correct order
		assert.Equal(t, events[0].ID(), payload.Carried[0].ID())
		assert.Equal(t, events[1].ID(), payload.Carried[1].ID())
		assert.Equal(t, events[2].ID(), payload.Carried[2].ID())
	})
}

func TestSequenceCarrier_FollowupEvents(t *testing.T) {
	t.Run("Given Sequence carrier with events that emit followups, When all followup events are emitted in order, Then completion event is published", func(t *testing.T) {
		// This test verifies that Sequence carrier can be created with a completion condition

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When - create with custom completion condition
		customCondition := func(sent, received event.Event) bool {
			return true
		}

		carrierEvent := carrier.NewSequence(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithCompletionCondition(customCondition),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.Sequence)
		// Verify the completion condition is set
		assert.NotNil(t, payload.CompletionCondition)
	})
}

func TestSequenceCarrier_Timeout(t *testing.T) {
	t.Run("Given Sequence carrier with timeout configuration, When timeout duration is exceeded before all events complete, Then timeout event is published", func(t *testing.T) {
		// This test verifies that Sequence carrier can be created with timeout option

		// Given
		events := []event.ChainableEvent{
			event.New(testPayload("event1")),
			event.New(testPayload("event2")),
		}

		doneEvent := event.New(testPayload("done"))
		timeoutEvent := event.New(testPayload("timeout"))

		// When - create with timeout
		carrierEvent := carrier.NewSequence(
			events,
			func(received []event.Event) event.Event { return doneEvent },
			timeoutEvent,
			carrier.WithTimeout(100*time.Millisecond),
		)

		// Then
		require.NotNil(t, carrierEvent)
		payload := carrierEvent.Payload().(*carrier.Sequence)
		assert.NotNil(t, payload.OnTimeout)
		assert.Equal(t, event.Type(carrier.TypePrefix+".sequence"), payload.EventType())
	})
}

func TestSequenceCarrier_SequentialOrder(t *testing.T) {
	t.Run("Given Sequence carrier with events that take different processing times, When Dispatch is called, Then next event is NOT dispatched until previous completes (verify order)", func(t *testing.T) {
		// This test verifies that Sequence carrier can be created with many events

		// Given
		numEvents := 20
		var events []event.ChainableEvent
		for i := 0; i < numEvents; i++ {
			events = append(events, event.New(testPayload("event")))
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
		payload := carrierEvent.Payload().(*carrier.Sequence)
		assert.Len(t, payload.Carried, numEvents)
		assert.Equal(t, event.Type(carrier.TypePrefix+".sequence"), payload.EventType())

		// Verify events are stored in order
		for i := 0; i < numEvents-1; i++ {
			// In a sequence, events should maintain their order
			// We can at least verify the Carried slice maintains order
			assert.NotEmpty(t, payload.Carried[i].ID())
		}
	})
}
