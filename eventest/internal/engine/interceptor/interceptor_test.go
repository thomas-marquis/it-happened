package interceptor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/gomockevent"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/clock"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/interceptor"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/timeline"
	"github.com/thomas-marquis/it-happened/inmemory"
	mocksevent "github.com/thomas-marquis/it-happened/internal/mocks/event"
	"go.uber.org/mock/gomock"
)

func TestInterceptor(t *testing.T) {
	t.Run("should pass when all events are published in the right order", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		gomock.InOrder(
			mockBus.EXPECT().Publish(gomockevent.PayloadEq(runtime.DefaultPayload("a"))).Times(1),
			mockBus.EXPECT().Publish(gomockevent.PayloadEq(runtime.DefaultPayload("b"))).Times(1),
			mockBus.EXPECT().Publish(gomockevent.PayloadEq(runtime.DefaultPayload("c"))).Times(1),
		)

		clock := clock.NewClock()
		tt := &testing.T{}
		it := interceptor.New(tt, mockBus, clock)

		// When
		it.EXPECT().FromMarble("^abc")

		clock.Start()

		// Tick 0: initEvent (no actual event published, just a marker)
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 1: event "a"
		it.Publish(event.New(runtime.DefaultPayload("a")))
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 2: event "b"
		it.Publish(event.New(runtime.DefaultPayload("b")))
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 3: event "c"
		it.Publish(event.New(runtime.DefaultPayload("c")))
		clock.Forward(timeline.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.False(t, tt.Failed())
	})

	t.Run("should fail when an event is missing", func(t *testing.T) {
		// Given
		clock := clock.NewClock()
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done)

		tt := &testing.T{}
		it := interceptor.New(tt, bus, clock)

		// When
		it.EXPECT().FromMarble("^abc")

		clock.Start()

		it.Publish(event.New(runtime.DefaultPayload("a")))
		clock.Forward(timeline.DefaultTickDuration)

		it.Publish(event.New(runtime.DefaultPayload("b")))
		clock.Forward(timeline.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.True(t, tt.Failed())
	})

	t.Run("should pass with ordered and unordered groups", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		mockBus.EXPECT().Publish(gomock.Any()).AnyTimes()

		clock := clock.NewClock()
		tt := &testing.T{}
		it := interceptor.New(tt, mockBus, clock)

		// When
		it.EXPECT().FromMarble("^[ab](cd)")

		clock.Start()

		// Tick 0: initEvent (no actual event)
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 1: [ab]
		it.Publish(event.New(runtime.DefaultPayload("a")))
		it.Publish(event.New(runtime.DefaultPayload("b")))
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 2: (cd)
		it.Publish(event.New(runtime.DefaultPayload("d")))
		it.Publish(event.New(runtime.DefaultPayload("c")))
		clock.Forward(timeline.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.False(t, tt.Failed())
	})

	t.Run("should pass with nested groups", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		mockBus.EXPECT().Publish(gomock.Any()).AnyTimes()

		clock := clock.NewClock()
		tt := &testing.T{}
		it := interceptor.New(tt, mockBus, clock)

		// When
		it.EXPECT().FromMarble("^[a(bc)]")

		clock.Start()

		// Tick 0: initEvent (no actual event)
		clock.Forward(timeline.DefaultTickDuration)

		// Tick 1: [a(bc)]
		it.Publish(event.New(runtime.DefaultPayload("a")))
		it.Publish(event.New(runtime.DefaultPayload("c")))
		it.Publish(event.New(runtime.DefaultPayload("b")))
		clock.Forward(timeline.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.False(t, tt.Failed())
	})
}
