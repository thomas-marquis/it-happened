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

func PayloadEq(pl event.Payload) gomock.Matcher {
	return payloadEqMatcher{pl: pl}
}

func (m payloadEqMatcher) Matches(x any) bool {
	if evt, ok := x.(event.Event); ok {
		return gomock.Eq(m.pl).Matches(evt.Payload)
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
