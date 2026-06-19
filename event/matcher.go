package event

// Matcher is the interface for matching events against specific criteria.
// Matchers are used by subscribers to filter which events they want to receive.
type Matcher interface {
	// Match checks if the given event matches this matcher's criteria.
	//
	// Parameters:
	//   event - The event to check
	//
	// Returns:
	//   true if the event matches, false otherwise
	Match(event Event) bool
}

// isMatcher matches events of a specific type.
type isMatcher struct {
	baseType Type
}

// Is creates a matcher that matches events of the given type.
//
// Parameters:
//
//	event - The event type to match
//
// Returns:
//
//	A Matcher that matches events with the specified type
func Is(event Type) Matcher {
	return &isMatcher{event}
}

// Match implements the Matcher interface for isMatcher.
// It returns true if the event's type matches the matcher's base type.
func (m *isMatcher) Match(event Event) bool {
	return m.baseType == event.Type()
}

// isOneOfMatcher matches events of any of the specified types.
type isOneOfMatcher struct {
	types []Type
}

// IsOneOf creates a matcher that matches events of any of the given types.
//
// Parameters:
//
//	eventTypes - The event types to match (one or more)
//
// Returns:
//
//	A Matcher that matches events with any of the specified types
func IsOneOf(eventTypes ...Type) Matcher {
	return &isOneOfMatcher{types: eventTypes}
}

// Match implements the Matcher interface for isOneOfMatcher.
// It returns true if the event's type matches any of the matcher's types.
func (m *isOneOfMatcher) Match(event Event) bool {
	for _, t := range m.types {
		if event.Type() == t {
			return true
		}
	}

	return false
}

// isAny matches all events.
type isAny struct{}

// IsAny creates a matcher that matches all events.
//
// Returns:
//
//	A Matcher that always returns true
func IsAny() Matcher {
	return &isAny{}
}

// Match implements the Matcher interface for isAny.
// It always returns true, matching any event.
func (m *isAny) Match(Event) bool {
	return true
}

// isFollowup matches followup events from specific parent events.
type isFollowup struct {
	events []Event
}

// IsFollowupOf creates a matcher that matches followup events of the given parent events.
//
// A followup event is one that was created as a followup of a parent event,
// sharing the same ChainRef but having a higher ChainPosition.
//
// Parameters:
//
//	event - One or more parent events whose followups should be matched
//
// Returns:
//
//	A Matcher that matches followup events of the specified parents
func IsFollowupOf(event ...Event) Matcher {
	return &isFollowup{events: event}
}

// Match implements the Matcher interface for isFollowup.
// It returns true if the event is a followup of any of the matcher's parent events.
func (m *isFollowup) Match(evt Event) bool {
	if evt.ChainPosition() == 0 {
		return false
	}

	for _, e := range m.events {
		if e.ID() != evt.ID() && e.ChainRef() == evt.ChainRef() && e.ChainPosition() < evt.ChainPosition() {
			return true
		}
	}
	return false
}
