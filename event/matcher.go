package event

type Matcher interface {
	Match(event Event) bool
}

type isMatcher struct {
	baseType Type
}

func Is(event Type) Matcher {
	return &isMatcher{event}
}

func (m *isMatcher) Match(event Event) bool {
	return m.baseType == event.Type()
}

type isOneOfMatcher struct {
	types []Type
}

func IsOneOf(eventTypes ...Type) Matcher {
	return &isOneOfMatcher{types: eventTypes}
}

func (m *isOneOfMatcher) Match(event Event) bool {
	for _, t := range m.types {
		if event.Type() == t {
			return true
		}
	}

	return false
}

type isAny struct{}

func IsAny() Matcher {
	return &isAny{}
}

func (m *isAny) Match(event Event) bool {
	return true
}

type isFollowup struct {
	events []Event
}

func IsFollowupOf(event ...Event) Matcher {
	return &isFollowup{events: event}
}

func (m *isFollowup) Match(event Event) bool {
	for _, e := range m.events {
		if e.ID != event.ID && e.Ref == event.Ref {
			return true
		}
	}
	return false
}

type isExactly struct {
	event Event
}

func IsExactly(event Event) Matcher {
	return &isExactly{event}
}

func (m *isExactly) Match(event Event) bool {
	return m.event == event
}

type hasPayload struct {
	pl Payload
}

func HasPayload(pl Payload) Matcher {
	return &hasPayload{pl: pl}
}

func (m *hasPayload) Match(event Event) bool {
	return m.pl == event.Payload
}
