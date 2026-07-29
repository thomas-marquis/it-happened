package eventest

import (
	"sync"
	"testing"
	"time"
)

// Wait waits for the WaitGroup to be done within the given timeout.
// If the timeout is reached, it fails the test.
func Wait(t *testing.T, wg *sync.WaitGroup, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Errorf("timeout waiting for WaitGroup after %v", timeout)
		t.FailNow()
	}
}

// WaitChan waits for the channel to be closed or receive a value within the given timeout.
// If the timeout is reached, it fails the test.
func WaitChan(t *testing.T, ch <-chan struct{}, timeout time.Duration) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(timeout):
		t.Errorf("timeout waiting for channel after %v", timeout)
		t.FailNow()
	}
}
