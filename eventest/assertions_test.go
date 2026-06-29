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

func TestContainsExactlyAllPayloads(t *testing.T) {
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
