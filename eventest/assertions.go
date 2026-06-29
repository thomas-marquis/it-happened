package eventest

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

// AssertContainsExactlyAllPayloads checks if the events contain exactly all the expected payloads.
func AssertContainsExactlyAllPayloads(t *testing.T, events []event.Event, expectedPayloads ...event.Payload) bool {
	t.Helper()

	if len(events) != len(expectedPayloads) {
		return assert.Fail(t, fmt.Sprintf("expected %d events, got %d", len(expectedPayloads), len(events)))
	}

	payloads := make([]event.Payload, len(events))
	for i, e := range events {
		payloads[i] = e.Payload()
	}

	res := true
	for _, e := range expectedPayloads {
		res = assert.Contains(t, payloads, e, "expected payload %v to be contained in %v", e, payloads)
		if !res {
			res = false
			break
		}
	}

	return res
}

// AssertIsFollowup checks if the target event is a followup of the source event.
func AssertIsFollowup(t *testing.T, source, target event.Event) bool {
	t.Helper()

	res := event.IsFollowupOf(source).Match(target)
	return assert.True(t, res, "expected %s to be a followup of %s", target, source)
}

// AssertIsType checks if the event is of the expected type.
func AssertIsType(t *testing.T, evt event.Event, expectedType event.Type) bool {
	t.Helper()

	return assert.Equal(t, expectedType, evt.Type(), "expected %s to be of type %s", evt, expectedType)
}
