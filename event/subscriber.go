package event

import "sync"

// Subscriber manages event subscriptions and callback execution.
// It matches incoming events against registered matchers and invokes the corresponding callbacks.
type Subscriber struct {
	sync.RWMutex

	registered map[Matcher][]func(Event)
	events     chan Event
	started    bool
	done       chan struct{}
}

// NewSubscriber creates a new Subscriber that listens on the given event channel.
//
// Parameters:
//
//	event - The channel through which events will be received
//
// Returns:
//
//	A new Subscriber instance ready to register callbacks
func NewSubscriber(event chan Event) *Subscriber {
	return &Subscriber{
		registered: make(map[Matcher][]func(Event)),
		events:     event,
		done:       make(chan struct{}),
	}
}

// On registers a callback function for events matching the given matcher.
//
// The callback will be invoked when an event matching the matcher is received.
// This method panics if called after listening has started.
//
// Parameters:
//
//	matcher - The matcher that determines which events trigger the callback
//	callback - The function to invoke when a matching event is received
//
// Returns:
//
//	The Subscriber instance for method chaining
func (s *Subscriber) On(matcher Matcher, callback func(Event)) *Subscriber {
	if s.started {
		panic("cannot register callback after listening started")
	}

	s.Lock()
	defer s.Unlock()
	if _, exists := s.registered[matcher]; !exists {
		s.registered[matcher] = make([]func(Event), 0)
	}

	s.registered[matcher] = append(s.registered[matcher], callback)
	return s
}

func (s *Subscriber) listen() {
	for {
		select {
		case <-s.done:
			return
		case event := <-s.events:
			s.RLock()
			for matcher, callbacks := range s.registered {
				if matcher.Match(event) {
					for _, callback := range callbacks {
						callback(event)
					}
				}
			}
			s.RUnlock()
		}
	}
}

// ListenWithWorkers starts multiple worker goroutines to process events.
//
// Each worker runs in its own goroutine and processes events concurrently.
// The number of workers determines the level of parallelism.
//
// Parameters:
//
//	workers - The number of concurrent worker goroutines to start
func (s *Subscriber) ListenWithWorkers(workers int) {
	s.started = true
	for i := 0; i < workers; i++ {
		go s.listen()
	}
}

// ListenNonBlocking starts a single event listener goroutine.
//
// Events are processed asynchronously, and callbacks for matching events
// are executed in separate goroutines to avoid blocking.
func (s *Subscriber) ListenNonBlocking() {
	s.started = true
	go func() {
		for {
			select {
			case <-s.done:
				return
			case event := <-s.events:
				s.RLock()
				for matcher, callbacks := range s.registered {
					if matcher.Match(event) {
						for _, callback := range callbacks {
							go callback(event)
						}
					}
				}
				s.RUnlock()
			}
		}
	}()
}

// Accept checks if the subscriber can accept (handle) the given event.
//
// It returns true if any registered matcher matches the event.
//
// Parameters:
//
//	event - The event to check
//
// Returns:
//
//	true if the event matches any registered matcher, false otherwise
func (s *Subscriber) Accept(event Event) bool {
	s.RLock()
	defer s.RUnlock()
	for matcher := range s.registered {
		if matcher.Match(event) {
			return true
		}
	}
	return false
}

// Detach stops the subscriber and releases its resources.
//
// This method closes the done channel, which signals all listener goroutines to exit.
func (s *Subscriber) Detach() {
	close(s.done)
}

func (s *Subscriber) Closed() bool {
	select {
	case _, ok := <-s.done:
		return !ok
	default:
		return false
	}
}
