package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

// testPayload is a test payload for notifier tests.
type testPayload string

// EventType implements the Payload interface for testPayload.
func (testPayload) EventType() event.Type {
	return "notifier.test.payload"
}

func TestNotifier_Notify(t *testing.T) {
	t.Run("Given notifier with registered callbacks, When it notifies subscribers, Then all registered callbacks are invoked", func(t *testing.T) {
		// Given
		var callback1Called, callback2Called bool

		notifier := &TestNotifier{
			Callbacks: []func(event.Event){
				func(evt event.Event) { callback1Called = true },
				func(evt event.Event) { callback2Called = true },
			},
		}

		testEvent := event.New(testPayload("test"))

		// When
		notifier.Notify(testEvent)

		// Then
		assert.True(t, callback1Called, "callback 1 should be called")
		assert.True(t, callback2Called, "callback 2 should be called")
	})
}

func TestNotifier_Empty(t *testing.T) {
	t.Run("Given notifier with no registered callbacks, When it attempts to notify, Then no panic occurs", func(t *testing.T) {
		// Given
		notifier := &TestNotifier{}
		testEvent := event.New(testPayload("test"))

		// When/Then - should not panic
		assert.NotPanics(t, func() {
			notifier.Notify(testEvent)
		})
	})
}

func TestNopNotifier_Notify(t *testing.T) {
	t.Run("Given NopNotifier, When Notify is called, Then it does nothing without panicking", func(t *testing.T) {
		// Given
		notifier := &event.NopNotifier{}
		testEvent := event.New(testPayload("test"))

		// When/Then - should not panic
		assert.NotPanics(t, func() {
			notifier.Notify(testEvent)
		})
	})
}

// TestNotifier is a test implementation of the Notifier interface.
type TestNotifier struct {
	Callbacks []func(event.Event)
}

// Notify implements the Notifier interface.
func (n *TestNotifier) Notify(evt event.Event) {
	for _, callback := range n.Callbacks {
		callback(evt)
	}
}
