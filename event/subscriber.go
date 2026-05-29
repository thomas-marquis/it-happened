package event

import "sync"

type Subscriber struct {
	sync.RWMutex

	registered map[Matcher]func(Event)
	events     chan Event
	started    bool
	done       chan struct{}
}
