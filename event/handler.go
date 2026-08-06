package event

import "sync"

type HandlerFunc func(evt Event)

// HarvesterHandler can be used to build a handler that collects all received events in a slice.
// This struct is thread-safe.
type HarvesterHandler struct {
	mu     sync.Mutex
	events []Event
}

// Handle is the handle function that can be passed to the On method.
func (h *HarvesterHandler) Handle(evt Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = append(h.events, evt)
}

// HandleWithWaitGroup returns a handler that both collects received events in a slice and decrements a WaitGroup.
func (h *HarvesterHandler) HandleWithWaitGroup(wg *sync.WaitGroup) HandlerFunc {
	return func(evt Event) {
		defer wg.Done()
		h.mu.Lock()
		defer h.mu.Unlock()
		h.events = append(h.events, evt)
	}
}

// Events returns all harvested events.
func (h *HarvesterHandler) Events() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.events
}
