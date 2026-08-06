package event_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
)

const (
	testTimeout = 5 * time.Second
)

func TestSubscriber(t *testing.T) {
	t.Run("should call the registered handler when an event matching the registered matcher is published", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)

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
		eventest.Wait(t, &wg, testTimeout)

		// Then
		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should accept matching events")
		assert.Equal(t, sub, result, "On should return the subscriber for chaining")
	})

	t.Run("should invoke all matching handlers when multiple handlers are registered with the same matcher", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 1)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)

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

		eventest.Wait(t, &wg1, testTimeout)
		eventest.Wait(t, &wg2, testTimeout)
	})

	t.Run("should not invoke a handler when the matcher does not match the event", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)

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
		sub := event.NewSubscriber(eventChan, event.Is(fakeType))

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
		eventest.Wait(t, &wg, testTimeout)

		assert.Equal(t, int32(2), count.Load(), "only matching events should be processed")
		assert.True(t, sub.Accept(event.New(fakePayload("coucou2"))))
		assert.False(t, sub.Accept(event.New(fakePayload2{})))
	})

	t.Run("should apply OR logic when multiple default matchers are specified", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan,
			event.Is(fakeType),
			event.Is(fakeType2),
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
		eventest.Wait(t, &wg, testTimeout)

		assert.Equal(t, int32(3), count.Load(), "events matching any default matcher should be processed")
		assert.True(t, sub.Accept(event.New(fakePayload("test"))))
		assert.True(t, sub.Accept(event.New(fakePayload2{})))
		assert.False(t, sub.Accept(event.New(fakePayload3{})))
	})

	t.Run("should work concurrently", func(t *testing.T) {
		// Given
		events := make(chan event.Event)
		defer close(events)

		var (
			callCnt1, callCnt2 atomic.Uint64
			n                  uint64 = 100
		)
		sub := event.NewSubscriber(events).
			On(event.Is(fakeType), func(evt event.Event) {
				callCnt1.Add(1)
			}).
			On(event.Is(fakeType2), func(evt event.Event) {
				callCnt2.Add(1)
			})
		sub.ListenWithWorkers(10)

		// When
		for i := range n {
			events <- event.New(fakePayload(fmt.Sprintf("test-%d", i)))
			events <- event.New(fakePayload2{})
		}

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.Equal(ct, n, callCnt1.Load())
			assert.Equal(ct, n, callCnt2.Load())
		}, testTimeout, 10*time.Millisecond)
	})
}

func TestSubscriber_Detach(t *testing.T) {
	t.Run("should no longer invoke a registered handler when detached", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 1)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)

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
		eventest.Wait(t, &wg, testTimeout)
		assert.Equal(t, uint32(1), cnt.Load(), "handler should be called once")

		sub.Detach()
		eventChan <- event.New(fakePayload("test"))
		assert.Equal(t, uint32(1), cnt.Load(), "handler should not be called after detach")

		assert.False(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should not accept matching events after detach")
	})

	t.Run("should clear all registered handlers when Detach is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)
		handler1 := func(evt event.Event) {}
		handler2 := func(evt event.Event) {}

		sub.On(matcher, handler1)
		sub.On(matcher, handler2)

		// When
		sub.Detach()

		// Then
		assert.False(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should not accept events after Detach clears handlers")
	})

	t.Run("should be safe to call Detach multiple times", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When & Then
		sub.Detach()

		assert.NotPanics(t, func() {
			sub.Detach()
		})
		assert.True(t, sub.Detached(), "subscriber should be closed after first Detach")
	})
}

func TestSubscriber_Accept(t *testing.T) {
	t.Run("should returns true if any matcher matches the event", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher1 := event.Is(fakeType)
		matcher2 := event.Is(fakeType2)

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
		defer close(eventChan)

		sub := event.NewSubscriber(eventChan)

		sub.On(event.IsAny(), func(evt event.Event) {
			time.Sleep(200 * time.Millisecond)
		})

		// When & Then
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

		eventest.WaitChan(t, done, testTimeout)

		sub.Detach()
		assert.True(t, sub.Detached())
	})
}

func TestSubscriber_ListenWithWorkers(t *testing.T) {
	t.Run("should panic when a handler is registered after the listening has started", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When
		sub.ListenWithWorkers(1)
		defer func() {
			if r := recover(); r != nil {
				// Expected panic
				assert.Contains(t, r.(string), "cannot register handler after listening started")
			} else {
				assert.Fail(t, "expected panic when registering handler after listening started")
			}
		}()

		// Then
		sub.On(event.IsAny(), func(evt event.Event) {})
	})

	t.Run("should not invoke any handlers after Detach is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)
		var called atomic.Bool
		sub.On(matcher, func(evt event.Event) {
			called.Store(true)
		})

		sub.ListenWithWorkers(1)

		// When
		sub.Detach()
		eventChan <- event.New(fakePayload("test"))

		// Then
		assert.Eventually(t, func() bool {
			return !called.Load()
		}, testTimeout, 10*time.Millisecond, "handler should not be invoked after Detach")
	})
}

func TestSubscriber_OnWithCancel(t *testing.T) {
	t.Run("should return a cancel function when OnWithCancel is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		// When
		cancel := sub.OnWithCancel(event.Is(fakeType), func(evt event.Event) {})

		// Then
		assert.NotNil(t, cancel, "OnWithCancel should return a cancel function")
		assert.NotPanics(t, func() {
			cancel()
		}, "cancel function should not panic when called")
	})

	t.Run("should remove the specific handler when cancel is called", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)
		var (
			called1, called2 atomic.Bool
		)

		cancel1 := sub.OnWithCancel(matcher, func(evt event.Event) {
			called1.Store(true)
		})
		sub.OnWithCancel(matcher, func(evt event.Event) {
			called2.Store(true)
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel1()
		eventChan <- event.New(fakePayload("test"))

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.False(ct, called1.Load(), "first handler should not be called after cancel")
			assert.True(ct, called2.Load(), "second handler should still be called")
		}, testTimeout, 10*time.Millisecond)
	})

	t.Run("should allow multiple OnWithCancel handlers to be independent", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)
		var called1, called2, called3 atomic.Bool

		cancel1 := sub.OnWithCancel(matcher, func(evt event.Event) { called1.Store(true) })
		cancel2 := sub.OnWithCancel(matcher, func(evt event.Event) { called2.Store(true) })
		_ = sub.OnWithCancel(matcher, func(evt event.Event) { called3.Store(true) })

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel1()
		cancel2()
		eventChan <- event.New(fakePayload("test"))
		time.Sleep(10 * time.Millisecond)

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.False(ct, called1.Load(), "first handler should not be called")
			assert.False(ct, called2.Load(), "second handler should not be called")
			assert.True(ct, called3.Load(), "third handler should still be called")
		}, testTimeout, 10*time.Millisecond)
	})

	t.Run("should be safe to call cancel function multiple times", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		var called atomic.Bool
		cancel := sub.OnWithCancel(event.Is(fakeType), func(evt event.Event) {
			called.Store(true)
		})

		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		cancel()
		cancel() // Second call - should be idempotent
		eventChan <- event.New(fakePayload("test"))

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.False(ct, called.Load(), "handler should not be called after cancel")
		}, testTimeout, 10*time.Millisecond)
	})

	t.Run("should be thread-safe when canceling concurrently", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 100)
		sub := event.NewSubscriber(eventChan)

		matcher := event.Is(fakeType)
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

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.Equal(ct, int32(0), callCount.Load(), "no handlers should be called after all are cancelled")
		}, testTimeout, 10*time.Millisecond)
	})

	t.Run("should work correctly when Detach is called after cancel", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		cancel := sub.OnWithCancel(event.Is(fakeType), func(evt event.Event) {})

		// When & Then
		assert.NotPanics(t, func() {
			cancel()
			sub.Detach()
		})
		assert.True(t, sub.Detached(), "subscriber should be closed")
	})

	t.Run("should be safe to call cancel after Detach", func(t *testing.T) {
		// Given
		eventChan := make(chan event.Event, 10)
		sub := event.NewSubscriber(eventChan)

		cancel := sub.OnWithCancel(event.Is(fakeType), func(evt event.Event) {})

		// When & Then
		sub.Detach()
		assert.NotPanics(t, func() {
			cancel()
		}, "cancel should be safe to call after Detach")

		// Then
		assert.True(t, sub.Detached(), "subscriber should be closed")
	})
}

func TestSubscriber_DetachOn(t *testing.T) {
	t.Run("should detach the subscriber when a matching is received", func(t *testing.T) {
		// Given
		events := make(chan event.Event)
		defer close(events)
		sub := event.NewSubscriber(events)

		sub.DetachOn(event.Is(fakeType))
		sub.ListenWithWorkers(1)
		require.False(t, sub.Detached())

		// When
		events <- event.New(fakePayload("test"))

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.True(ct, sub.Detached(), "subscriber should be closed")
		}, 1*time.Second, 10*time.Millisecond)
	})

	t.Run("should trigger the other handler before detaching when multiple matchers are set", func(t *testing.T) {
		// Given
		events := make(chan event.Event)
		defer close(events)

		var (
			called1, called2 atomic.Bool
		)

		sub := event.NewSubscriber(events)

		sub.
			DetachOn(event.Is(fakeType)).
			On(event.Is(fakeType), func(evt event.Event) {
				called1.Store(true)
			}).
			OnWithCancel(event.Is(fakeType), func(evt event.Event) {
				called2.Store(true)
			})
		sub.ListenWithWorkers(1)
		require.False(t, sub.Detached())

		// When
		events <- event.New(fakePayload("test"))

		// Then
		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.True(ct, called1.Load(), "first handler should be called")
			assert.True(ct, called2.Load(), "second handler should be called")
			assert.True(ct, sub.Detached(), "subscriber should be closed")
		}, 1*time.Second, 10*time.Millisecond)
	})

	t.Run("should detach concurrently", func(t *testing.T) {
		// Given
		events := make(chan event.Event)
		defer close(events)

		var (
			callCnt atomic.Int64
		)

		sub := event.NewSubscriber(events).
			DetachOn(event.Is(fakeType2)).
			On(event.Is(fakeType), func(evt event.Event) {
				callCnt.Add(1)
			})
		sub.ListenWithWorkers(10)

		// When
		for i := range 9 {
			events <- event.New(fakePayload(fmt.Sprintf("test-%d", i)))
		}
		events <- event.New(fakePayload2{})

		assert.EventuallyWithT(t, func(ct *assert.CollectT) {
			assert.True(ct, sub.Detached(), "subscriber should be detached")
		}, 1*time.Second, 10*time.Millisecond)
		t.Logf("%s handler called %d times", fakeType, callCnt.Load())
	})
}
