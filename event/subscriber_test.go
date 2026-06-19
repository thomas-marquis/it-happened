package event_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

// Test payload types
type fakePayload3 struct{}

func (fakePayload3) EventType() event.Type {
	return "different.payload"
}

func TestSubscriber_Register(t *testing.T) {
	t.Run("should call the registered callback when an event matching the registered matcher is published", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var (
			cnt atomic.Uint32
			wg  sync.WaitGroup
		)
		wg.Add(1)
		result := sub.On(matcher, func(evt event.Event) {
			cnt.Add(1)
			wg.Done()
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		eventChan <- event.New(fakePayload("test"))
		wg.Wait()

		// Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")
		assert.Equal(t, sub, result, "On should return the subscriber for chaining")
	})
}

func TestSubscriber_Detach(t *testing.T) {
	t.Run("should no longer invoke a registered callback when detached", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 1)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var (
			cnt atomic.Uint32
			wg  sync.WaitGroup
		)
		wg.Add(1)
		sub.On(matcher, func(evt event.Event) {
			cnt.Add(1)
			wg.Done()
		})

		sub.ListenWithWorkers(1)

		// When & Then
		eventChan <- event.New(fakePayload("test"))
		wg.Wait()
		assert.Equal(t, uint32(1), cnt.Load(), "handler should be called once")

		sub.Detach()
		eventChan <- event.New(fakePayload("test"))
		assert.Equal(t, uint32(1), cnt.Load(), "handler should not be called after detach")

		assert.False(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should not accept matching events after detach")
	})
}

func TestSubscriber_MultipleHandlers(t *testing.T) {
	t.Run("should invoke all matching handlers when multiple handlers are registered with the same matcher", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 1)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var (
			wg1, wg2 sync.WaitGroup
			w1Done   = make(chan struct{})
			w2Done   = make(chan struct{})
		)

		wg1.Add(1)
		sub.On(matcher, func(evt event.Event) {
			wg1.Done()
			close(w1Done)
			println("handler 1")
		})
		wg2.Add(1)
		sub.On(matcher, func(evt event.Event) {
			wg2.Done()
			close(w2Done)
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")
		eventChan <- event.New(fakePayload("test"))

		wg1.Wait()
		select {
		case <-w1Done:
			println("handler 1")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for handler 1")
		}

		wg2.Wait()
		select {
		case <-w2Done:
			println("handler 2")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for handler 2")
		}
	})
}

func TestSubscriber_NonMatching(t *testing.T) {
	t.Run("should not invoke a callback when the matcher does not match the event", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var called bool
		sub.On(matcher, func(evt event.Event) {
			called = true
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		nonMatchingEvent := event.New(fakePayload2{})

		// Then
		assert.False(t, sub.Accept(nonMatchingEvent), "subscriber should not accept non-matching events")
		assert.False(t, called, "handler should not be called")
	})
}

func TestSubscriber_Accept(t *testing.T) {
	t.Run("should returns true if any matcher matches the event", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher1 := event.Is("fake.payload")
		matcher2 := event.Is("fake.payload.2")

		sub.On(matcher1, func(evt event.Event) {})
		sub.On(matcher2, func(evt event.Event) {})

		// When & Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "should accept fake.payload events")
		assert.True(t, sub.Accept(event.New(fakePayload2{})), "should accept fake.payload.2 events")
		assert.False(t, sub.Accept(event.New(fakePayload3{})), "should not accept different event types")
	})
}

func TestSubscriber_ListenNonBlocking(t *testing.T) {
	t.Run("Given subscriber, When ListenNonBlocking is called, Then it starts listening in a goroutine", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event)
		//defer close(eventChan)

		sub := event.NewSubscriber(eventChan)

		sub.On(event.IsAny(), func(evt event.Event) {
			time.Sleep(200 * time.Millisecond)
		})

		// When/Then
		assert.NotPanics(t, func() {
			sub.ListenNonBlocking()
		})

		done := make(chan struct{})
		go func() {
			eventChan <- event.New(fakePayload("test"))
			eventChan <- event.New(fakePayload("test"))
			eventChan <- event.New(fakePayload("test"))
			eventChan <- event.New(fakePayload2{})
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for events")
		}

		sub.Detach()
		assert.True(t, sub.Detached())
	})
}

func TestSubscriber_PanicAfterListenStarted(t *testing.T) {
	t.Run("should panic when a callback is registered after the listening has started", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When
		sub.ListenWithWorkers(1)
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
				assert.Contains(t, r.(string), "cannot register callback after listening started")
			} else {
				assert.Fail(t, "expected panic when registering callback after listening started")
			}
		}()

		// Then
		sub.On(event.IsAny(), func(evt event.Event) {})
	})
}

func TestSubscriber_DetachClearsCallbacks(t *testing.T) {
	t.Run("should clear all registered callbacks when Detach is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		callback1 := func(evt event.Event) {}
		callback2 := func(evt event.Event) {}

		sub.On(matcher, callback1)
		sub.On(matcher, callback2)

		// When
		sub.Detach()

		// Then
		// After Detach(), the registered map should be empty
		// This will fail initially until implementation is added
		assert.False(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should not accept events after Detach clears callbacks")
	})
}

func TestSubscriber_NoCallbacksAfterDetach(t *testing.T) {
	t.Run("should not invoke any callbacks after Detach is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var called bool
		sub.On(matcher, func(evt event.Event) {
			called = true
		})

		sub.ListenWithWorkers(1)

		// When
		sub.Detach()
		eventChan <- event.New(fakePayload("test"))

		// Then
		// Give some time for the event to be processed
		time.Sleep(10 * time.Millisecond)
		assert.False(t, called, "callback should not be invoked after Detach")
	})
}

func TestSubscriber_DetachIdempotent(t *testing.T) {
	t.Run("should be safe to call Detach multiple times", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When
		sub.Detach()
		sub.Detach() // Second call

		// Then
		assert.True(t, sub.Detached(), "subscriber should be closed after first Detach")
		// Should not panic
	})
}

func TestSubscriber_OnWithCancel_ReturnsCancelFunction(t *testing.T) {
	t.Run("should return a cancel function when OnWithCancel is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When
		cancel := sub.OnWithCancel(event.Is("fake.payload"), func(evt event.Event) {})

		// Then
		assert.NotNil(t, cancel, "OnWithCancel should return a cancel function")
		assert.NotPanics(t, func() {
			cancel()
		}, "cancel function should not panic when called")
	})
}

func TestSubscriber_OnWithCancel_RemovesSpecificCallback(t *testing.T) {
	t.Run("should remove the specific callback when cancel is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var (
			called1, called2 bool
			mu               sync.Mutex
		)

		cancel1 := sub.OnWithCancel(matcher, func(evt event.Event) {
			mu.Lock()
			defer mu.Unlock()
			called1 = true
		})
		sub.OnWithCancel(matcher, func(evt event.Event) {
			mu.Lock()
			defer mu.Unlock()
			called2 = true
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel1()
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		mu.Lock()
		assert.False(t, called1, "first callback should not be called after cancel")
		assert.True(t, called2, "second callback should still be called")
		mu.Unlock()
	})
}

func TestSubscriber_OnWithCancel_MultipleIndependent(t *testing.T) {
	t.Run("should allow multiple OnWithCancel callbacks to be independent", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var called1, called2, called3 bool

		cancel1 := sub.OnWithCancel(matcher, func(evt event.Event) { called1 = true })
		cancel2 := sub.OnWithCancel(matcher, func(evt event.Event) { called2 = true })
		_ = sub.OnWithCancel(matcher, func(evt event.Event) { called3 = true })

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When - cancel only the first two
		cancel1()
		cancel2()
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		assert.False(t, called1, "first callback should not be called")
		assert.False(t, called2, "second callback should not be called")
		assert.True(t, called3, "third callback should still be called")
	})
}

func TestSubscriber_OnWithCancel_CancelIdempotent(t *testing.T) {
	t.Run("should be safe to call cancel function multiple times", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		var called bool
		cancel := sub.OnWithCancel(event.Is("fake.payload"), func(evt event.Event) {
			called = true
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel()
		cancel() // Second call - should be idempotent
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		assert.False(t, called, "callback should not be called after cancel")
	})
}

func TestSubscriber_OnWithCancel_ConcurrentCancellation(t *testing.T) {
	t.Run("should be thread-safe when canceling concurrently", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 100)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var wg sync.WaitGroup
		callCount := atomic.Int32{}

		// Create many callbacks
		cancels := make([]func(), 100)
		for i := 0; i < 100; i++ {
			cancels[i] = sub.OnWithCancel(matcher, func(evt event.Event) {
				callCount.Add(1)
			})
		}

		sub.ListenWithWorkers(4)
		defer sub.Detach()

		// When - cancel all concurrently
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				defer wg.Done()
				cancels[idx]()
			}(i)
		}
		wg.Wait()

		// Send an event
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(50 * time.Millisecond)

		// Then - no callbacks should have been called
		assert.Equal(t, int32(0), callCount.Load(), "no callbacks should be called after all are cancelled")
	})
}

func TestSubscriber_OnWithCancel_DetachAfterCancel(t *testing.T) {
	t.Run("should work correctly when Detach is called after cancel", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		cancel := sub.OnWithCancel(event.Is("fake.payload"), func(evt event.Event) {
			// callback
		})

		// When
		cancel()
		sub.Detach()

		// Then
		assert.True(t, sub.Detached(), "subscriber should be closed")
		// Should not panic
	})
}

func TestSubscriber_OnWithCancel_CancelAfterDetach(t *testing.T) {
	t.Run("should be safe to call cancel after Detach", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		cancel := sub.OnWithCancel(event.Is("fake.payload"), func(evt event.Event) {
			// callback
		})

		// When
		sub.Detach()
		// Should not panic
		assert.NotPanics(t, func() {
			cancel()
		}, "cancel should be safe to call after Detach")

		// Then
		assert.True(t, sub.Detached(), "subscriber should be closed")
	})
}
