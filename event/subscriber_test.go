package event_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
)

// Test payload types
type fakePayload3 struct{}

func (fakePayload3) EventType() event.Type {
	return "different.payload"
}

func TestSubscriber(t *testing.T) {
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
		eventest.Wait(t, &wg, time.Second)

		// Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")
		assert.Equal(t, sub, result, "On should return the subscriber for chaining")
	})

	t.Run("should invoke all matching handlers when multiple handlers are registered with the same matcher", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 1)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var (
			wg1, wg2 sync.WaitGroup
		)

		wg1.Add(1)
		sub.On(matcher, func(evt event.Event) {
			wg1.Done()
			println("handler 1")
		})
		wg2.Add(1)
		sub.On(matcher, func(evt event.Event) {
			wg2.Done()
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")
		eventChan <- event.New(fakePayload("test"))

		eventest.Wait(t, &wg1, time.Second)
		eventest.Wait(t, &wg2, time.Second)
	})

	t.Run("should not invoke a callback when the matcher does not match the event", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")

		var called atomic.Bool
		sub.On(matcher, func(evt event.Event) {
			called.Store(true)
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		nonMatchingEvent := event.New(fakePayload2{})

		// Then
		assert.False(t, sub.Accept(nonMatchingEvent), "subscriber should not accept non-matching events")
		assert.False(t, called.Load(), "handler should not be called")
	})

	t.Run("should apply default matcher before all", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan, event.Is("fake.payload"))

		var (
			count atomic.Int32
			wg    sync.WaitGroup
		)

		sub.On(event.IsAny(), func(evt event.Event) {
			if _, ok := evt.Payload().(fakePayload2); ok {
				assert.Fail(t, "subscriber should not accept non-matching events")
				return
			}
			count.Add(1)
			wg.Done()
		})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		wg.Add(2)
		eventChan <- event.New(fakePayload("match1"))
		eventChan <- event.New(fakePayload2{}) // Should be ignored by default matcher
		eventChan <- event.New(fakePayload("match2"))

		// Then
		// If match2 is processed, we are sure fakePayload2 was processed (and ignored) before it
		// because we have only 1 worker and channels are FIFO.
		eventest.Wait(t, &wg, time.Second)

		assert.Equal(t, int32(2), count.Load(), "only matching events should be processed")
		assert.True(t, sub.Accept(event.New(fakePayload("coucou2"))))
		assert.False(t, sub.Accept(event.New(fakePayload2{})))
	})

	t.Run("should apply OR logic when multiple default matchers are specified", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan,
			event.Is("fake.payload"),
			event.Is("fake.payload.2"),
		)

		var (
			count atomic.Int32
			wg    sync.WaitGroup
		)

		sub.On(event.IsAny(), func(evt event.Event) {
			if _, ok := evt.Payload().(fakePayload3); ok {
				assert.Fail(t, "subscriber should not accept non-matching events")
				return
			}
			count.Add(1)
			wg.Done()
		})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		wg.Add(3)
		eventChan <- event.New(fakePayload("match1"))
		eventChan <- event.New(fakePayload2{})
		eventChan <- event.New(fakePayload3{}) // Should be ignored
		eventChan <- event.New(fakePayload("match2"))

		// Then
		eventest.Wait(t, &wg, time.Second)

		assert.Equal(t, int32(3), count.Load(), "events matching any default matcher should be processed")
		assert.True(t, sub.Accept(event.New(fakePayload("test"))))
		assert.True(t, sub.Accept(event.New(fakePayload2{})))
		assert.False(t, sub.Accept(event.New(fakePayload3{})))
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
		eventest.Wait(t, &wg, time.Second)
		assert.Equal(t, uint32(1), cnt.Load(), "handler should be called once")

		sub.Detach()
		eventChan <- event.New(fakePayload("test"))
		assert.Equal(t, uint32(1), cnt.Load(), "handler should not be called after detach")

		assert.False(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should not accept matching events after detach")
	})

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

func TestSubscriber_ListenWithWorkers(t *testing.T) {
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

	t.Run("should not invoke any callbacks after Detach is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var called atomic.Bool
		sub.On(matcher, func(evt event.Event) {
			called.Store(true)
		})

		sub.ListenWithWorkers(1)

		// When
		sub.Detach()
		eventChan <- event.New(fakePayload("test"))

		// Then
		// Give some time for the event to be processed
		time.Sleep(10 * time.Millisecond)
		assert.False(t, called.Load(), "callback should not be invoked after Detach")
	})
}

func TestSubscriber_OnWithCancel(t *testing.T) {
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

	t.Run("should allow multiple OnWithCancel callbacks to be independent", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var called1, called2, called3 atomic.Bool

		cancel1 := sub.OnWithCancel(matcher, func(evt event.Event) { called1.Store(true) })
		cancel2 := sub.OnWithCancel(matcher, func(evt event.Event) { called2.Store(true) })
		_ = sub.OnWithCancel(matcher, func(evt event.Event) { called3.Store(true) })

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When - cancel only the first two
		cancel1()
		cancel2()
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		assert.False(t, called1.Load(), "first callback should not be called")
		assert.False(t, called2.Load(), "second callback should not be called")
		assert.True(t, called3.Load(), "third callback should still be called")
	})

	t.Run("should be safe to call cancel function multiple times", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		var called atomic.Bool
		cancel := sub.OnWithCancel(event.Is("fake.payload"), func(evt event.Event) {
			called.Store(true)
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel()
		cancel() // Second call - should be idempotent
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		assert.False(t, called.Load(), "callback should not be called after cancel")
	})

	t.Run("should be thread-safe when canceling concurrently", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 100)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is("fake.payload")
		var wg sync.WaitGroup
		callCount := atomic.Int32{}

		cancels := make([]func(), 100)
		for i := 0; i < 100; i++ {
			cancels[i] = sub.OnWithCancel(matcher, func(evt event.Event) {
				callCount.Add(1)
			})
		}

		sub.ListenWithWorkers(4)
		defer sub.Detach()

		// When
		wg.Add(100)
		for i := 0; i < 100; i++ {
			go func(idx int) {
				defer wg.Done()
				cancels[idx]()
			}(i)
		}
		wg.Wait()

		eventChan <- event.New(fakePayload("test"))
		time.Sleep(50 * time.Millisecond)

		// Then
		assert.Equal(t, int32(0), callCount.Load(), "no callbacks should be called after all are cancelled")
	})

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
