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

		assert.True(t, sub.Accept(event.New(fakePayload("test"))), "subscriber should still accept matching events after detach")
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
		assert.True(t, sub.Closed())
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
