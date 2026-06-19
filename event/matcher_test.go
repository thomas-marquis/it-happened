package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

type fakePayload string

func (fakePayload) EventType() event.Type {
	return "fake.payload"
}

type fakePayload2 struct{}

func (fakePayload2) EventType() event.Type {
	return "fake.payload.2"
}

func TestIsFollowupOf(t *testing.T) {
	t.Run("should match when event is a followup of at least one of the given events", func(t *testing.T) {
		// Given
		events := []event.ChainableEvent{
			event.New(fakePayload("event1")),
			event.New(fakePayload("event2")),
		}

		m := event.IsFollowupOf(events...)

		incoming := events[0].NewFollowup(fakePayload2{})

		// When
		res := m.Match(incoming)

		// Then
		assert.True(t, res)
	})

	t.Run("should not match when event is not a followup of at least one of the given events", func(t *testing.T) {
		// Given
		events := []event.ChainableEvent{
			event.New(fakePayload("event1")),
			event.New(fakePayload("event2")),
		}

		m := event.IsFollowupOf(events...)

		incoming := event.New(fakePayload2{})

		// When
		res := m.Match(incoming)

		// Then
		assert.False(t, res)
	})

	t.Run("should not match the same event as the given ones", func(t *testing.T) {
		e := event.New(fakePayload("event2"))
		events := []event.ChainableEvent{
			event.New(fakePayload("event1")),
			e,
		}

		m := event.IsFollowupOf(events...)

		// When
		res := m.Match(e)

		// Then
		assert.False(t, res)
	})
}

func TestIsOneOf(t *testing.T) {
	t.Run("should match when event type matches one of the given types", func(t *testing.T) {
		// Given
		m := event.IsOneOf("fake.payload", "other.payload", "fake.payload.2")
		evt := event.New(fakePayload("test"))

		// When
		res := m.Match(evt)

		// Then
		assert.True(t, res)
	})

	t.Run("should not match when event type does not match any of the given types", func(t *testing.T) {
		// Given
		m := event.IsOneOf("other.payload", "different.payload")
		evt := event.New(fakePayload("test"))

		// When
		res := m.Match(evt)

		// Then
		assert.False(t, res)
	})
}

func TestIsAny(t *testing.T) {
	t.Run("should match any event", func(t *testing.T) {
		// Given
		m := event.IsAny()

		// When/Then - should match all event types
		assert.True(t, m.Match(event.New(fakePayload("test"))))
		assert.True(t, m.Match(event.New(fakePayload2{})))
	})
}
