package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/eventest/internal/runtime"
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
			mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("aaa"))).Times(1),
			mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("bbb"))).Times(1),
			mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("ccc"))).Times(1),
		)

		clock := runtime.NewClock()
		tt := &testing.T{}
		it := runtime.NewInterceptor(tt, mockBus, clock)

		// When
		it.EXPECT().FromMarble("abc")

		clock.Start()

		it.Publish(event.New(fakePayload("aaa")))
		clock.Forward(runtime.DefaultTickDuration)

		it.Publish(event.New(fakePayload("bbb")))
		clock.Forward(runtime.DefaultTickDuration)

		it.Publish(event.New(fakePayload("ccc")))
		clock.Forward(runtime.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.False(t, tt.Failed())
	})

	t.Run("should fail when an event is missing", func(t *testing.T) {
		// Given
		clock := runtime.NewClock()
		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, nil)

		tt := &testing.T{}
		it := runtime.NewInterceptor(tt, bus, clock)

		// When
		it.EXPECT().FromMarble("abc")

		clock.Start()

		it.Publish(event.New(fakePayload("aaa")))
		clock.Forward(runtime.DefaultTickDuration)

		it.Publish(event.New(fakePayload("bbb")))
		clock.Forward(runtime.DefaultTickDuration)

		clock.Stop()
		it.Finish()

		// Then
		assert.True(t, tt.Failed())
	})
}
