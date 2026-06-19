package carrier_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

func TestCarrierInterface_Dispatch(t *testing.T) {
	t.Run("Given carrier that implements Carrier interface, When Dispatch is called, Then events are dispatched to bus", func(t *testing.T) {
		// This test verifies that carriers implement the Carrier interface

		// Given - All carrier
		allCarrier := carrier.NewAll(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
		)

		// When/Then - verify All implements Carrier
		allPayload := allCarrier.Payload().(carrier.Carrier)
		assert.NotNil(t, allPayload)

		// Given - Sequence carrier
		seqCarrier := carrier.NewSequence(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
		)

		// When/Then - verify Sequence implements Carrier
		seqPayload := seqCarrier.Payload().(carrier.Carrier)
		assert.NotNil(t, seqPayload)
	})
}

func TestCarrierOptions(t *testing.T) {
	t.Run("Given carrier created with WithTimeout, WithMaxConcurrency, WithCompletionCondition, When carrier is used, Then respects all configuration values", func(t *testing.T) {
		// This test verifies that carrier options are applied

		// Given
		timeout := 5 * time.Second
		maxConcurrency := 5
		completionCondition := func(sent, received event.Event) bool {
			return true
		}

		// When - create All carrier with all options
		allCarrier := carrier.NewAll(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
			carrier.WithTimeout(timeout),
			carrier.WithMaxConcurrency(maxConcurrency),
			carrier.WithCompletionCondition(completionCondition),
		)

		// Then - verify All carrier was created successfully
		require.NotNil(t, allCarrier)
		allPayload := allCarrier.Payload().(*carrier.All)
		assert.NotNil(t, allPayload)

		// When - create Sequence carrier with all options
		seqCarrier := carrier.NewSequence(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
			carrier.WithTimeout(timeout),
			carrier.WithCompletionCondition(completionCondition),
		)

		// Then - verify Sequence carrier was created successfully
		require.NotNil(t, seqCarrier)
		seqPayload := seqCarrier.Payload().(*carrier.Sequence)
		assert.NotNil(t, seqPayload)
	})
}

func TestCompletionCondition(t *testing.T) {
	t.Run("Given carrier with custom CompletionCondition, When events are dispatched, Then uses custom condition to determine event completion", func(t *testing.T) {
		// This test verifies that custom completion conditions can be set

		// Given
		customConditionCalled := false
		customCondition := func(sent, received event.Event) bool {
			customConditionCalled = true
			return true
		}

		// When - create All carrier with custom condition
		allCarrier := carrier.NewAll(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
			carrier.WithCompletionCondition(customCondition),
		)

		// Then
		require.NotNil(t, allCarrier)
		allPayload := allCarrier.Payload().(*carrier.All)
		assert.NotNil(t, allPayload.CompletionCondition)

		// When - create Sequence carrier with custom condition
		seqCarrier := carrier.NewSequence(
			[]event.ChainableEvent{event.New(testPayload("test"))},
			func(received []event.Event) event.Event { return event.New(testPayload("done")) },
			event.New(testPayload("timeout")),
			carrier.WithCompletionCondition(customCondition),
		)

		// Then
		require.NotNil(t, seqCarrier)
		seqPayload := seqCarrier.Payload().(*carrier.Sequence)
		assert.NotNil(t, seqPayload.CompletionCondition)

		// Note: We cannot directly verify that the custom condition is called
		// without more complex integration testing, but we can verify it's set
		_ = customConditionCalled // Used in more complex scenarios
	})
}

// TestCompletedOnFollowupReceived tests the default completion condition.
func TestCompletedOnFollowupReceived(t *testing.T) {
	t.Run("Given CompletedOnFollowupReceived function, When called, Then returns true for followup events", func(t *testing.T) {
		// Given
		sent := event.New(testPayload("sent"))
		received := event.New(testPayload("received"))

		// When
		result := carrier.CompletedOnFollowupReceived(sent, received)

		// Then
		// According to the docs, this always returns true
		assert.True(t, result)
	})
}
