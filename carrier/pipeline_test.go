package carrier_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/inmemory"
)

const (
	testTimeout = 5 * time.Second
)

func TestPipelineCarrier_Dispatch(t *testing.T) {
	t.Run("should dispatch all events sequentially when carrier is published", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eg := &event.HarvesterNotifier{}
		bus := inmemory.NewBus(ctx, inmemory.WithNotifier(eg))

		var wg sync.WaitGroup

		initEvent := event.New(testPayload("event1"))
		timeoutEvent := event.New(testPayload("timeout"))
		event2 := event.New(testPayload("event2"))
		event3 := event.New(testPayload("event3"))

		expectedIDs := map[string]bool{
			initEvent.ID(): true,
			event2.ID():    true,
			event3.ID():    true,
		}

		wg.Add(3)
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if !expectedIDs[evt.ID()] {
					return
				}
				bus.Publish(evt.NewFollowup("followup" + evt.Payload().(testPayload)))
				wg.Done()
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		stages := []carrier.PipelineStage{
			func(prev event.Event) event.Event {
				assert.Equal(t, "followupevent1", string(prev.Payload().(testPayload)))
				return event2
			},
			func(prev event.Event) event.Event {
				assert.Equal(t, "followupevent2", string(prev.Payload().(testPayload)))
				return event3
			},
		}

		carrierEvent := carrier.NewPipeline(
			initEvent,
			stages,
			timeoutEvent,
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, testTimeout):
			publishedEvents := eg.Events
			require.Len(t, publishedEvents, 7)

			// Verify all original events were received
			idSet := make(map[string]struct{})
			for _, evt := range publishedEvents {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, 7)
			assert.Equal(t, "event1", string(publishedEvents[1].Payload().(testPayload)))
			assert.Equal(t, "followupevent1", string(publishedEvents[2].Payload().(testPayload)))
			assert.Equal(t, "event2", string(publishedEvents[3].Payload().(testPayload)))
			assert.Equal(t, "followupevent2", string(publishedEvents[4].Payload().(testPayload)))
			assert.Equal(t, "event3", string(publishedEvents[5].Payload().(testPayload)))
			assert.Equal(t, "followupevent3", string(publishedEvents[6].Payload().(testPayload)))

		case <-time.After(testTimeout):
			assert.Fail(t, "timeout waiting for all events")
		}
	})

	t.Run("should publish timeout event when sequence processing exceeds timeout duration", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		bus := inmemory.NewBus(ctx)

		var timeoutReceived bool
		var mu sync.Mutex

		initEvent := event.New(testPayload("event1"))
		event2 := event.New(testPayload("event2"))
		event3 := event.New(testPayload("event3"))

		timeoutEvent := event.New(testPayload("timeout"))

		stages := []carrier.PipelineStage{
			func(prev event.Event) event.Event {
				return event2
			},
			func(prev event.Event) event.Event {
				return event3
			},
		}

		carrierEvent := carrier.NewPipeline(
			initEvent,
			stages,
			timeoutEvent,
			carrier.WithTimeout(50*time.Millisecond),
		)

		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if evt.ID() == timeoutEvent.ID() {
					mu.Lock()
					timeoutReceived = true
					mu.Unlock()
				}
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		// When
		bus.Publish(carrierEvent)

		// Then
		time.Sleep(200 * time.Millisecond)
		mu.Lock()
		assert.True(t, timeoutReceived, "timeout event should be published")
		mu.Unlock()
	})

	t.Run("should interrupt the pipeline when a stop event is published", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		eg := &event.HarvesterNotifier{}
		bus := inmemory.NewBus(ctx, inmemory.WithNotifier(eg))

		var wg sync.WaitGroup

		initEvent := event.New(testPayload("event1"))
		timeoutEvent := event.New(testPayload("timeout"))
		event2 := event.New(testPayload("event2"))
		event3 := event.New(testPayload("event3"))
		event4 := event.New(testPayload("event4"))

		expectedIDs := map[string]bool{
			initEvent.ID(): true,
			event2.ID():    true,
			event3.ID():    true,
		}

		wg.Add(3)
		sub := bus.Subscribe().
			On(event.Is("test.payload"), func(evt event.Event) {
				if !expectedIDs[evt.ID()] {
					return
				}
				bus.Publish(evt.NewFollowup("followup" + evt.Payload().(testPayload)))
				wg.Done()
			})
		sub.ListenWithWorkers(1)
		defer sub.Detach()

		stages := []carrier.PipelineStage{
			func(prev event.Event) event.Event {
				return event2
			},
			func(prev event.Event) event.Event {
				return carrier.StopPipelineWithEvent(event3)
			},
			func(prev event.Event) event.Event {
				assert.Fail(t, "event4 should never be published")
				return event4
			},
		}

		carrierEvent := carrier.NewPipeline(
			initEvent,
			stages,
			timeoutEvent,
		)

		// When
		bus.Publish(carrierEvent)

		// Then
		select {
		case <-waitForEvents(t, &wg, testTimeout):
			publishedEvents := eg.Events
			require.Len(t, publishedEvents, 7)

			// Verify all original events were received
			idSet := make(map[string]struct{})
			for _, evt := range publishedEvents {
				idSet[evt.ID()] = struct{}{}
			}
			assert.Len(t, idSet, 7)
			assert.Equal(t, "followupevent2", string(publishedEvents[4].Payload().(testPayload)))
			assert.Equal(t, "event3", string(publishedEvents[5].Payload().(testPayload)))
			assert.Equal(t, "followupevent3", string(publishedEvents[6].Payload().(testPayload)))

		case <-time.After(testTimeout):
			assert.Fail(t, "timeout waiting for all events")
		}
	})
}
