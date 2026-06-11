package gomockevent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/gomockevent"
)

type fakePayload string

func (fakePayload) Type() event.Type {
	return "fake"
}

func TestPayloadEqMatcher_Matches(t *testing.T) {
	t.Run("should return true when payload are equal", func(t *testing.T) {
		// Given
		m := gomockevent.PayloadEq(fakePayload("my value"))

		// When
		res := m.Matches(event.New(fakePayload("my value")))

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when payload are not equal", func(t *testing.T) {
		// Given
		m := gomockevent.PayloadEq(fakePayload("my value"))

		// When
		res := m.Matches(event.New(fakePayload("other value")))

		// Then
		assert.False(t, res)
	})

	t.Run("should return false when the provided argument is not an event", func(t *testing.T) {
		// Given
		m := gomockevent.PayloadEq(fakePayload("my value"))

		// When
		res := m.Matches("not an event")

		// Then
		assert.False(t, res)
	})
}

func TestPayloadEqMatcher_String(t *testing.T) {
	t.Run("should return a string representation of the matcher", func(t *testing.T) {
		// Given
		m := gomockevent.PayloadEq(fakePayload("my value"))

		// When
		res := m.String()

		// Then
		assert.Equal(t, "is equal to \"my value\" (eventest_test.fakePayload)", res)
	})
}

func TestIsFollowupOf_Matches(t *testing.T) {
	t.Run("should return true when the event is a followup", func(t *testing.T) {
		// Given
		fromEvt := event.New(fakePayload("my value"))
		evt := event.NewFollowup(fromEvt, fakePayload("my new value"))

		m := gomockevent.IsFollowupOf(fromEvt)

		// When
		res := m.Matches(evt)

		// Then
		assert.True(t, res)
	})

	t.Run("should return false when the event is not a followup", func(t *testing.T) {
		// Given
		otherEvt := event.New(fakePayload("a value"))
		evt := event.New(fakePayload("another value"))

		m := gomockevent.IsFollowupOf(otherEvt)

		// When
		res := m.Matches(evt)

		// Then
		assert.False(t, res)
	})
}

func TestIsFollowupOf_String(t *testing.T) {
	t.Run("should return a string representation of the matcher", func(t *testing.T) {
		// Given
		fromEvt := event.New(fakePayload("my value"))
		m := gomockevent.IsFollowupOf(fromEvt)

		// When
		res := m.String()

		// Then
		assert.Equal(t, "is a followup of \"my value\" (eventest_test.fakePayload)", res)

	})
}
