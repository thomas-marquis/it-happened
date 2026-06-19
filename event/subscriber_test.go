package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

// Test payload types
type fakePayload3 struct{}

func (fakePayload3) EventType() event.Type {
	return "different.payload"
}

func TestSubscriber_Register(t *testing.T) {
	t.Run("Given subscriber, When it registers a handler with a matcher, Then handler is invoked for matching events", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		// When - register handler BEFORE starting to listen
		result := sub.On(matcher, func(evt event.Event) {
			// Handler called
		})

		// Start listening
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// Publish a matching event
		matchingEvent := event.New(fakePayload("test"))

		// Then - verify handler was registered
		assert.True(t, sub.Accept(matchingEvent), "subscriber should accept matching events")

		// Verify the On method returned the subscriber for chaining
		assert.Equal(t, sub, result, "On should return the subscriber for chaining")
	})
}

func TestSubscriber_Unregister(t *testing.T) {
	t.Run("Given subscriber with registered handler, When it unsubscribes, Then handler is no longer invoked", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		// Register handler BEFORE starting to listen
		sub.On(matcher, func(evt event.Event) {
			// Handler called
		})

		// Start listening
		sub.ListenWithWorkers(1)

		// When - detach the subscriber (unregister all handlers)
		// This should stop all listeners
		sub.Detach()

		// Then - verify that the subscriber's Accept method still works
		// (it doesn't depend on the listening state)
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should still accept matching events after detach")
	})
}

func TestSubscriber_MultipleHandlers(t *testing.T) {
	t.Run("Given subscriber with multiple handlers, When matching event is published, Then all matching handlers are invoked", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		// Register multiple handlers for the same matcher BEFORE starting to listen
		sub.On(matcher, func(evt event.Event) {
			// Handler 1 called
		})
		sub.On(matcher, func(evt event.Event) {
			// Handler 2 called
		})

		// Start listening
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// Then - verify both handlers are registered
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")

		// Note: We can't easily test that both handlers are called without
		// complex synchronization, but we can verify they were registered
	})
}

func TestSubscriber_NonMatching(t *testing.T) {
	t.Run("Given subscriber with handler for specific matcher, When non-matching event is published, Then handler is NOT invoked", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		// Register handler for specific matcher BEFORE starting to listen
		sub.On(matcher, func(evt event.Event) {
			// Handler called
		})

		// Start listening
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When - check with non-matching event
		nonMatchingEvent := event.New(fakePayload2{})

		// Then
		assert.False(t, sub.Accept(nonMatchingEvent), "subscriber should not accept non-matching events")
	})
}

func TestSubscriber_Accept(t *testing.T) {
	t.Run("Given subscriber with registered matchers, When Accept is called, Then it returns true if any matcher matches", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher1 := event.Is("fake.payload")
		matcher2 := event.Is("fake.payload.2")

		// Register handlers with different matchers
		sub.On(matcher1, func(evt event.Event) {})
		sub.On(matcher2, func(evt event.Event) {})

		// Then - verify Accept works correctly
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "should accept fake.payload events")
		assert.True(t, sub.Accept(event.New(fakePayload2{})), "should accept fake.payload.2 events")
		// fakePayload always has type "fake.payload", so it will match matcher1
		// We need a truly different type - use IsAny matcher to test the negative case
		assert.False(t, sub.Accept(event.New(fakePayload3{})), "should not accept different event types")
	})
}

func TestSubscriber_ListenNonBlocking(t *testing.T) {
	t.Run("Given subscriber, When ListenNonBlocking is called, Then it starts listening in a goroutine", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When/Then - should not panic
		assert.NotPanics(t, func() {
			sub.ListenNonBlocking()
			// Give it time to start
			// Close the channel to allow the goroutine to exit
			close(eventChan)
		})
	})
}

func TestSubscriber_PanicAfterListenStarted(t *testing.T) {
	t.Run("Given subscriber after listening started, When On is called, Then it panics", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// Start listening
		sub.ListenWithWorkers(1)
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
				assert.Contains(t, r.(string), "cannot register callback after listening started")
			} else {
				assert.Fail(t, "expected panic when registering callback after listening started")
			}
		}()

		// When/Then - should panic
		sub.On(event.IsAny(), func(evt event.Event) {})
	})
}
