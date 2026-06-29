package eventest_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
)

type fakePayload string

func (fakePayload) EventType() event.Type {
	return "fake.payload"
}

func TestAssertContainsExactlyAllPayloads(t *testing.T) {
	for _, tc := range []struct {
		In       []event.Event
		Payloads []event.Payload
		Expected bool
		CaseDesc string
	}{
		{
			CaseDesc: "contains exactly all payloads",
			In: []event.Event{
				event.New(fakePayload("payload1")),
				event.New(fakePayload("payload2")),
			},
			Payloads: []event.Payload{
				fakePayload("payload1"), fakePayload("payload2"),
			},
			Expected: true,
		},
		{
			CaseDesc: "empty payloads and events",
			In:       []event.Event{},
			Payloads: []event.Payload{},
			Expected: true,
		},
		{
			CaseDesc: "more payloads than events",
			In: []event.Event{
				event.New(fakePayload("payload1")),
			},
			Payloads: []event.Payload{
				fakePayload("payload1"), fakePayload("payload2"),
			},
			Expected: false,
		},
		{
			CaseDesc: "less payloads than events",
			In: []event.Event{
				event.New(fakePayload("payload1")),
				event.New(fakePayload("payload2")),
			},
			Payloads: []event.Payload{
				fakePayload("payload1"),
			},
			Expected: false,
		},
		{
			CaseDesc: "one payload is not the same",
			In: []event.Event{
				event.New(fakePayload("payload1")),
				event.New(fakePayload("payload2")),
			},
			Payloads: []event.Payload{
				fakePayload("payload1"),
				fakePayload("payload3"),
			},
			Expected: false,
		},
	} {
		t.Run(fmt.Sprintf("should return %v when %s", tc.Expected, tc.CaseDesc), func(t *testing.T) {
			// Given
			var tt testing.T

			// When
			res := eventest.AssertContainsExactlyAllPayloads(&tt, tc.In, tc.Payloads...)

			// Then
			assert.Equal(t, tc.Expected, res)
			assert.Equal(t, !tc.Expected, tt.Failed())
		})
	}
}

func TestAssertIsFollowup(t *testing.T) {
	t.Run("should assert that the event is a followup of the given event when it is the case", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))
		followup := evt.NewFollowup(fakePayload("test"))
		var tt testing.T

		// Then
		res := eventest.AssertIsFollowup(&tt, evt, followup)

		// Then
		assert.False(t, tt.Failed())
		assert.True(t, res)
	})

	t.Run("should not assert that the event is a followup of the given event when it is not the case", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))
		other := event.New(fakePayload("test"))
		var tt testing.T

		// Then
		res := eventest.AssertIsFollowup(&tt, evt, other)

		// Then
		assert.True(t, tt.Failed())
		assert.False(t, res)
	})
}

func TestAssertIsType(t *testing.T) {
	t.Run("should assert that the event is of the given type when it is the case", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))
		var tt testing.T

		// Then
		res := eventest.AssertIsType(&tt, evt, "fake.payload")

		// Then
		assert.False(t, tt.Failed())
		assert.True(t, res)
	})

	t.Run("should not assert the event is of the given type when it is not the case", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))
		var tt testing.T

		// Then
		res := eventest.AssertIsType(&tt, evt, "other.payload")

		// Then
		assert.True(t, tt.Failed())
		assert.False(t, res)
	})
}
