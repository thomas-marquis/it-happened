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

func (m *isAny) Match(Event) bool {
	return true
}

type isFollowup struct {
	events []ChainableEvent
}

func IsFollowupOf(event ...ChainableEvent) Matcher {
	return &isFollowup{events: event}
}

func (m *isFollowup) Match(event Event) bool {
	evt, ok := event.(ChainableEvent)
	if !ok {
		return false
	} else if evt.ChainPosition() == 0 {
		return false
	}

	for _, e := range m.events {
		if e.ID() != evt.ID() && e.ChainRef() == evt.ChainRef() {
			return true
		}
	}
	return false
}
