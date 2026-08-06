package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

type testPayload string

func (testPayload) EventType() event.Type {
	return "notifier.test.payload"
}

func TestNopNotifier_Notify(t *testing.T) {
	t.Run("Given NopNotifier, When Notify is called, Then it does nothing without panicking", func(t *testing.T) {
		// Given
		notifier := &event.NopNotifier{}
		testEvent := event.New(testPayload("test"))

		// When & Then
		assert.NotPanics(t, func() {
			notifier.NotifyPublished(testEvent)
		})
	})
}

func TestCombinedNotifier(t *testing.T) {
	// Given
	n1 := &event.HarvesterNotifier{}
	n2 := &event.HarvesterNotifier{}

	cn := event.NewCombinedNotifier(n1, n2)

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	bus := inmemory.NewBus(ctx, inmemory.WithNotifier(cn))

	e1 := event.New(testPayload("test1"))
	e2 := event.New(testPayload("test2"))

	// When
	bus.Publish(e1)
	bus.Publish(e2)

	// Then
	assert.Len(t, n1.UnsafeEvents, 2)
	assert.Len(t, n2.UnsafeEvents, 2)
}
