package eventest

import (
	"encoding/json"
	"fmt"

	"github.com/thomas-marquis/it-happened/event"
	"go.uber.org/mock/gomock"
)

type payloadEqMatcher struct {
	pl event.Payload
}

// PayloadEq returns a gomock.Matcher that matches events with the given payload.
//
// Example usage:
//
//	PayloadEq(myPayload("test").Matches(event.New(myPayload("test"))) // returns true
//	PayloadEq(myPayload("test").Matches(event.New(myPayload("other"))) // returns false
func PayloadEq(pl event.Payload) gomock.Matcher {
	return payloadEqMatcher{pl: pl}
}

func (m payloadEqMatcher) Matches(x any) bool {
	if evt, ok := x.(event.Event); ok {
		return gomock.Eq(m.pl).Matches(evt.Payload())
	}
	return false
}

func (m payloadEqMatcher) String() string {
	repr, err := json.Marshal(m.pl)
	if err != nil {
		return gomock.Eq(m.pl).String()
	}
	return fmt.Sprintf("is equal to %s (%T)", string(repr), m.pl)
}

type isFollowupOfMatcher struct {
	from event.Event
}

// IsFollowupOf returns a gomock.Matcher that matches events that are followups of the given event.
//
// Example usage:
//
//	a := event.New(myPayload("test"))
//	b := a.NewFollowup(myPayload("followup"))
//	c := event.New(myPayload("other"))
//
//	IsFollowupOf(a).Matches(b) // returns true
//	IsFollowupOf(a).Matches(c) // returns false
func IsFollowupOf(from event.Event) gomock.Matcher {
	return isFollowupOfMatcher{from: from}
}

func (m isFollowupOfMatcher) Matches(x any) bool {
	evt, ok := x.(event.Event)
	if !ok {
		return false
	}
	return event.IsFollowupOf(m.from).Match(evt)
}

func (m isFollowupOfMatcher) String() string {
	repr, err := json.Marshal(m.from.Payload())
	if err != nil {
		return fmt.Sprintf("is a followup of %s", m.from.ID())
	}
	return fmt.Sprintf("is a followup of %s (%T)", string(repr), m.from.Payload())
}
