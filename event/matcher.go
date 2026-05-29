package event

type Matcher interface {
	Match(event Event) bool
}
