package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

type fakePayload string

func (fakePayload) Type() event.Type {
	return "fake.payload"
}

type fakePayload2 struct{}

func (fakePayload2) Type() event.Type {
	return "fake.payload.2"
}

func TestIsFollowupOf(t *testing.T) {
	t.Run("should match when event is a followup of at least one of the given events", func(t *testing.T) {
		// Given
		events := []event.Event{
			event.New(fakePayload("event1")),
			event.New(fakePayload("event2")),
		}

		m := event.IsFollowupOf(events...)

		incoming := event.NewFollowup(events[0], fakePayload2{})

		// When
		res := m.Match(incoming)

		// Then
		assert.True(t, res)
	})

	t.Run("should not match when event is not a followup of at least one of the given events", func(t *testing.T) {
		// Given
		events := []event.Event{
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
		events := []event.Event{
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

func TestIsExactly_Match(t *testing.T) {
	t.Run("should return true when the events are equals", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("an event"))
		m := event.IsExactly(evt)

		// When
		res := m.Match(evt)

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when the events are not equals even with the same payload", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("an event"))
		m := event.IsExactly(evt)

		// When
		res := m.Match(event.New(fakePayload("an event")))

		// Then
		assert.False(t, res)
	})
}

func TestHasPayload_Match(t *testing.T) {
	t.Run("should return true when the events have the same payload", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("an event"))
		m := event.HasPayload(fakePayload("an event"))

		// When
		res := m.Match(evt)

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when payloads are different", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("an event"))
		m := event.HasPayload(fakePayload2{})

		// When
		res := m.Match(evt)

		// Then
		assert.False(t, res)
	})
}
