package carrier_test

import (
	"sync"
	"testing"
	"time"

	"github.com/thomas-marquis/it-happened/event"
)

// testPayload is a test payload for creating test events.
type testPayload string

// EventType implements the Payload interface for testPayload.
func (testPayload) EventType() event.Type {
	return "test.payload"
}

// testPayload2 is another test payload type.
type testPayload2 struct {
	Value string
}

// EventType implements the Payload interface for testPayload2.
func (testPayload2) EventType() event.Type {
	return "test.payload.2"
}

// slowPayload is a test payload type for testing timeout scenarios.
type slowPayload struct {
	Value string
}

// EventType implements the Payload interface for slowPayload.
func (slowPayload) EventType() event.Type {
	return "slow.payload"
}

// waitForEvents waits for the waitgroup and returns a channel that will be closed when done.
func waitForEvents(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) chan struct{} {
	t.Helper()
	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	return doneCh
}
