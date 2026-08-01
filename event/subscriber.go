package event

import (
	"sync"
	"sync/atomic"
)

// Subscriber manages event subscriptions and callback execution.
// It matches incoming events against registered matchers and invokes the corresponding callbacks.
type Subscriber struct {
	mu sync.RWMutex

	registered      map[Matcher][]func(Event)
	cancellable     map[Matcher][]*cancellableCallback
	detachOn        map[Matcher]struct{}
	events          chan Event
	started         bool
	done            chan struct{}
	detached        bool
	nextCancelID    uint64
	defaultMatchers []Matcher
}

// cancellableCallback wraps a callback with a unique ID for cancellation
type cancellableCallback struct {
	id       uint64
	callback func(Event)
}

// NewSubscriber creates a new Subscriber that listens on the given event channel.
// To be processed, the event must match at least one of the default matchers, if any.
//
// Parameters:
//
//	event - The channel through which events will be received
//	defaultMatchers - Optional matchers that will be applied to all registered callbacks
//
// Returns:
//
//	A new Subscriber instance ready to register callbacks
func NewSubscriber(event chan Event, defaultMatchers ...Matcher) *Subscriber {
	return &Subscriber{
		registered:      make(map[Matcher][]func(Event)),
		cancellable:     make(map[Matcher][]*cancellableCallback),
		detachOn:        make(map[Matcher]struct{}),
		events:          event,
		done:            make(chan struct{}),
		defaultMatchers: defaultMatchers,
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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		panic("cannot register callback after listening started")
	}

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
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		panic("cannot register callback after listening started")
	}

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
		s.mu.Lock()
		defer s.mu.Unlock()
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

// DetachOn registers a matcher that will detach the subscriber at the first matching event is received.
// If many matchers can match the same event, the detaching will happen after all other registered callbacks are executed.
func (s *Subscriber) DetachOn(matcher Matcher) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		panic("cannot register callback after listening started")
	}

	if _, exists := s.detachOn[matcher]; exists {
		panic("only one DetachOn can be registered per matcher")
	}
	s.detachOn[matcher] = struct{}{}
	return s
}

func (s *Subscriber) listen(execute func(cb func(Event), evt Event)) {
	for {
		select {
		case <-s.done:
			return
		case event := <-s.events:
			if event == nil {
				continue
			}
			s.mu.RLock()
			if !s.matchDefault(event) {
				s.mu.RUnlock()
				continue
			}

			for matcher, callbacks := range s.registered {
				if matcher.Match(event) {
					for _, callback := range callbacks {
						execute(callback, event)
					}
				}
			}
			for matcher, cancellables := range s.cancellable {
				if matcher.Match(event) {
					for _, cc := range cancellables {
						execute(cc.callback, event)
					}
				}
			}
			var shouldDetach bool
			for matcher := range s.detachOn {
				if matcher.Match(event) {
					shouldDetach = true
					break
				}
			}
			s.mu.RUnlock()

			if shouldDetach {
				s.mu.Lock()
				s.doDetach()
				s.mu.Unlock()
				return
			}
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
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	for i := 0; i < workers; i++ {
		go s.listen(func(cb func(Event), evt Event) {
			cb(evt)
		})
	}
}

// ListenNonBlocking starts a single event listener goroutine.
//
// Events are processed asynchronously, and callbacks for matching events
// are executed in separate goroutines to avoid blocking.
func (s *Subscriber) ListenNonBlocking() {
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	go s.listen(func(cb func(Event), evt Event) {
		go cb(evt)
	})
}

// Accept checks if the subscriber can accept (handle) the given event.
//
// It returns true if any registered matcher matches the event.
// Default matchers are applied before any registered matchers.
//
// Parameters:
//
//	event - The event to check
//
// Returns:
//
//	true if the event matches any registered matcher, false otherwise
func (s *Subscriber) Accept(event Event) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.matchDefault(event) {
		return false
	}

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
	for matcher := range s.detachOn {
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
	if s.Detached() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.doDetach()
}

func (s *Subscriber) doDetach() {
	clear(s.registered)
	clear(s.cancellable)
	close(s.done)
	s.detached = true
}

func (s *Subscriber) Detached() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.detached
}

func (s *Subscriber) matchDefault(evt Event) bool {
	if len(s.defaultMatchers) == 0 {
		return true
	}
	for _, matcher := range s.defaultMatchers {
		if matcher.Match(evt) {
			return true
		}
	}
	return false
}
