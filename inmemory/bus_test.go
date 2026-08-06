package inmemory_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/inmemory"
	mocks_events "github.com/thomas-marquis/it-happened/internal/mocks/events"
	goMock "go.uber.org/mock/gomock"
)

const (
	testType  event.Type = "test.payload"
	testType2 event.Type = "test.payload.2"
)

type testPayload string

func (testPayload) EventType() event.Type {
	return testType
}

type testPayload2 struct {
	Value string
}

func (testPayload2) EventType() event.Type {
	return testType2
}

func setupBus(t *testing.T) (func(), event.Bus) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	bus := inmemory.NewBus(ctx)
	return cancel, bus
}

type busFixture struct {
	ctx       context.Context
	cancel    context.CancelFunc
	bus       event.Bus
	harvester *event.HarvesterNotifier
	t         *testing.T
}

func setupBusFixture(t *testing.T) *busFixture {
	t.Helper()
	fxt := &busFixture{t: t}

	fxt.ctx, fxt.cancel = context.WithCancel(context.Background())
	fxt.bus = inmemory.NewBus(fxt.ctx,
		inmemory.WithNotifier(fxt.Harvester()))

	t.Cleanup(fxt.teardown)

	return fxt
}

func (f *busFixture) Harvester() *event.HarvesterNotifier {
	f.t.Helper()
	if f.harvester == nil {
		f.harvester = &event.HarvesterNotifier{}
	}
	return f.harvester
}

func (f *busFixture) Bus() event.Bus {
	f.t.Helper()
	return f.bus
}

func (f *busFixture) teardown() {
	f.cancel()
}

func TestInmemoryBus(t *testing.T) {
	t.Run("should deliver published event to one subscriber", func(t *testing.T) {
		// Given
		fxt := setupBusFixture(t)

		wg := sync.WaitGroup{}
		wg.Add(1)

		sub := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
		})
		sub.ListenWithWorkers(1)

		testEvent := event.New(testPayload("test"))

		// When
		fxt.Bus().Publish(testEvent)

		// Then
		eventest.Wait(t, &wg, time.Second)

		assert.Len(t, fxt.Harvester().Events(), 1)
		assert.Equal(t, testEvent, fxt.Harvester().Events()[0])
	})

	t.Run("should deliver published event to all subscribers", func(t *testing.T) {
		// Given
		fxt := setupBusFixture(t)

		testEvent := event.New(testPayload("test"))

		var wg sync.WaitGroup
		wg.Add(3)

		sub1 := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			assert.Equal(t, testEvent, evt)
		})
		sub1.ListenWithWorkers(1)

		sub2 := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			assert.Equal(t, testEvent, evt)
		})
		sub2.ListenWithWorkers(1)

		sub3 := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			assert.Equal(t, testEvent, evt)
		})
		sub3.ListenWithWorkers(1)

		// When
		fxt.Bus().Publish(testEvent)

		// Then
		eventest.Wait(t, &wg, time.Second)

		assert.Len(t, fxt.Harvester().Events(), 1)
		assert.Equal(t, testEvent, fxt.Harvester().Events()[0])
	})

	t.Run("should handle concurrent publish without data races", func(t *testing.T) {
		// Given
		fxt := setupBusFixture(t)

		numEvents := 100
		var wg sync.WaitGroup
		wg.Add(numEvents)

		sub := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
		})
		sub.ListenWithWorkers(16)

		// When
		for i := 0; i < numEvents; i++ {
			go func(idx int) {
				evt := event.New(testPayload2{Value: "event"})
				fxt.Bus().Publish(evt)
			}(i)
		}

		// Then
		eventest.Wait(t, &wg, 2*time.Second)
		require.Len(t, fxt.Harvester().Events(), numEvents)

		idSet := make(map[string]struct{})
		for _, evt := range fxt.Harvester().Events() {
			idSet[evt.ID()] = struct{}{}
		}
		assert.Len(t, idSet, numEvents, "all events should have unique IDs")
	})

	t.Run("should deliver events only to subscribers with matching criteria", func(t *testing.T) {
		// Given
		fxt := setupBusFixture(t)

		event1 := event.New(testPayload("test1"))

		var wg sync.WaitGroup
		wg.Add(2)
		sub1 := fxt.Bus().Subscribe().On(event.Is(testType), func(evt event.Event) {
			defer wg.Done()
			assert.Equal(t, event1, evt)
		})
		sub1.ListenWithWorkers(1)

		sub2 := fxt.Bus().Subscribe().On(event.Is(testType2), func(evt event.Event) {
			assert.Failf(t, "subscriber2", "%s is not supposed to be received", testType2)
		})
		sub2.ListenWithWorkers(1)

		sub3 := fxt.Bus().Subscribe().On(event.IsAny(), func(evt event.Event) {
			defer wg.Done()
			assert.Equal(t, event1, evt)
		})
		sub3.ListenWithWorkers(1)

		// When
		fxt.Bus().Publish(event1)

		// Then
		eventest.Wait(t, &wg, time.Second)
		assert.Len(t, fxt.Harvester().Events(), 1)
		assert.Equal(t, event1, fxt.Harvester().Events()[0])
	})

	t.Run("should handle concurrent publish and subscribe without race conditions", func(t *testing.T) {
		// Given
		closeBus, bus := setupBus(t)
		defer closeBus()

		numPublishers := 10
		numEventsPerPublisher := 10
		totalEvents := numPublishers * numEventsPerPublisher

		hh := event.HarvesterHandler{}
		var wg sync.WaitGroup
		wg.Add(totalEvents)

		sub := bus.Subscribe().On(event.IsAny(), hh.HandleWithWaitGroup(&wg))
		sub.ListenWithWorkers(16)

		var opWg sync.WaitGroup

		// Publishers
		for i := 0; i < numPublishers; i++ {
			opWg.Add(1)
			go func(publisherID int) {
				defer opWg.Done()
				for j := 0; j < numEventsPerPublisher; j++ {
					bus.Publish(event.New(testPayload2{Value: "event"}))
				}
			}(i)
		}

		// Concurrent subscribers
		for i := 0; i < 5; i++ {
			opWg.Add(1)
			go func() {
				defer opWg.Done()
				bus.Subscribe().
					On(event.IsAny(), func(evt event.Event) {}).
					ListenWithWorkers(1)
			}()
		}

		opWg.Wait()

		// Then
		eventest.Wait(t, &wg, 2*time.Second)
		require.Len(t, hh.Events(), totalEvents)

		idSet := make(map[string]struct{})
		for _, evt := range hh.Events() {
			idSet[evt.ID()] = struct{}{}
		}
		assert.Len(t, idSet, totalEvents, "all events should have unique IDs")
	})
}

func TestInmemoryBus_Notifier(t *testing.T) {
	t.Run("should notify notifier when configured", func(t *testing.T) {
		// Given & Then
		ctrl := goMock.NewController(t)
		mockNotifier := mocks_events.NewMockNotifier(ctrl)

		evt := event.New(testPayload("test"))

		var mu sync.Mutex
		var sub *event.Subscriber

		mockNotifier.EXPECT().
			NotifyPublished(goMock.Eq(evt)).
			Times(1)

		mockNotifier.EXPECT().
			NotifySubscribed(goMock.Cond(func(s *event.Subscriber) bool {
				mu.Lock()
				defer mu.Unlock()
				sub = s
				return true
			})).
			Times(1)

		mockNotifier.EXPECT().
			NotifyUnsubscribed(goMock.Cond(func(s *event.Subscriber) bool {
				mu.Lock()
				defer mu.Unlock()
				return assert.Equal(t, sub, s)
			})).
			Times(1)

		ctx, cancel := context.WithCancel(context.Background())
		bus := inmemory.NewBus(ctx, inmemory.WithNotifier(mockNotifier))
		defer cancel()

		// When
		bus.Publish(evt)
		res := bus.Subscribe()
		bus.Unsubscribe(sub)

		mu.Lock()
		mu.Unlock()
		assert.Equal(t, res, sub)
	})

	t.Run("should work with multiple events and notifications", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockCtrl := goMock.NewController(t)

		mockNotifier := mocks_events.NewMockNotifier(mockCtrl)
		bus := inmemory.NewBus(ctx, inmemory.WithNotifier(mockNotifier))

		mockNotifier.EXPECT().NotifySubscribed(goMock.Any()).Times(2)

		sub1 := bus.Subscribe()
		_ = bus.Subscribe()

		event1 := event.New(testPayload("event1"))
		event2 := event.New(testPayload("event2"))

		mockNotifier.EXPECT().NotifyPublished(event1).Times(1)
		mockNotifier.EXPECT().NotifyPublished(event2).Times(1)

		mockNotifier.EXPECT().NotifyUnsubscribed(goMock.Any()).Times(1)

		// When

		bus.Publish(event1)
		bus.Publish(event2)
		bus.Unsubscribe(sub1)
	})
}
