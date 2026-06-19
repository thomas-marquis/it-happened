package event

import (
	"sync"
	"sync/atomic"
)

// Subscriber manages event subscriptions and callback execution.
// It matches incoming events against registered matchers and invokes the corresponding callbacks.
type Subscriber struct {
	sync.RWMutex

	registered   map[Matcher][]func(Event)
	cancellable  map[Matcher][]*cancellableCallback
	events       chan Event
	started      bool
	done         chan struct{}
	detached     bool
	nextCancelID uint64
}

// cancellableCallback wraps a callback with a unique ID for cancellation
type cancellableCallback struct {
	id       uint64
	callback func(Event)
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
		registered:  make(map[Matcher][]func(Event)),
		cancellable: make(map[Matcher][]*cancellableCallback),
		events:      event,
		done:        make(chan struct{}),
	}
}

// On registers a callback function for events matching the given matcher.
//
// The callback will be invoked when an event matching the matcher is received.
// This method panics if called after listening has started.
//
// Note: Callbacks registered via On() persist until the Subscriber is detached.
// For subscriptions requiring individual cleanup, use OnWithCancel() instead.
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

// OnWithCancel registers a callback for events matching the given matcher
// and returns a function to cancel/unregister that specific callback.
//
// The callback will be invoked when an event matching the matcher is received.
// Unlike On(), this method allows fine-grained removal of individual callbacks
// without detaching the entire subscriber.
//
// Parameters:
//
//	matcher - The matcher that determines which events trigger the callback
//	callback - The function to invoke when a matching event is received
//
// Returns:
//
//	A function that, when called, removes this specific callback
func (s *Subscriber) OnWithCancel(matcher Matcher, callback func(Event)) func() {
	if s.started {
		panic("cannot register callback after listening started")
	}

	s.Lock()
	defer s.Unlock()

	// Generate a unique ID for this callback
	id := atomic.AddUint64(&s.nextCancelID, 1)
	cc := &cancellableCallback{
		id:       id,
		callback: callback,
	}

	// Add to cancellable map
	if _, exists := s.cancellable[matcher]; !exists {
		s.cancellable[matcher] = make([]*cancellableCallback, 0)
	}
	s.cancellable[matcher] = append(s.cancellable[matcher], cc)

	// Return cancellation function
	return func() {
		s.Lock()
		defer s.Unlock()
		if s.detached {
			// Subscriber has been detached, all callbacks are already cleared
			return
		}
		if callbacks, exists := s.cancellable[matcher]; exists {
			for i, cc := range callbacks {
				if cc.id == id {
					// Remove by swapping with last element and slicing
					callbacks[i] = callbacks[len(callbacks)-1]
					s.cancellable[matcher] = callbacks[:len(callbacks)-1]
					// Clean up empty matcher entries
					if len(s.cancellable[matcher]) == 0 {
						delete(s.cancellable, matcher)
					}
					break
				}
			}
		}
	}
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
			for matcher, cancellables := range s.cancellable {
				if matcher.Match(event) {
					for _, cc := range cancellables {
						cc.callback(event)
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
				for matcher, cancellables := range s.cancellable {
					if matcher.Match(event) {
						for _, cc := range cancellables {
							go cc.callback(event)
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
	for matcher := range s.cancellable {
		if matcher.Match(event) {
			return true
		}
	}
	return false
}

// Detach stops the subscriber and releases its resources.
//
// This method closes the done channel, which signals all listener goroutines to exit,
// and clears all registered callbacks to prevent memory leaks.
// This method is idempotent and safe to call multiple times.
func (s *Subscriber) Detach() {
	s.Lock()
	defer s.Unlock()

	if s.detached {
		return
	}

	s.registered = make(map[Matcher][]func(Event))
	s.cancellable = make(map[Matcher][]*cancellableCallback)
	close(s.done)
	s.detached = true
}

func (s *Subscriber) Detached() bool {
	return s.detached
}
