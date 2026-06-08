package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/runtime"
	"github.com/thomas-marquis/it-happened/inmemory"
	mockruntime "github.com/thomas-marquis/it-happened/internal/mocks/runtime"
	"go.uber.org/mock/gomock"
)

func TestInterceptor(t *testing.T) {
	t.Run("should pass when all events are published in the right order", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)

		mockClock := mockruntime.NewMockClock(ctrl)

		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, nil)

		tt := &testing.T{}
		it := runtime.NewInterceptor(tt, bus, mockClock)

		// When
		it.EXPECT().FromMarble("abc")

		bus.Publish(event.New(fakePayload("aaa")))
		bus.Publish(event.New(fakePayload("bbb")))
		bus.Publish(event.New(fakePayload("ccc")))

		it.Finish()

		// Then
		assert.False(t, tt.Failed())
	})

	t.Run("should fail when an event is missing", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)

		mockClock := mockruntime.NewMockClock(ctrl)

		done := make(chan struct{})
		defer close(done)

		bus := inmemory.NewBus(done, nil)

		tt := &testing.T{}
		it := runtime.NewInterceptor(tt, bus, mockClock)

		// When
		it.EXPECT().FromMarble("abc")
		bus.Publish(event.New(fakePayload("aaa")))
		bus.Publish(event.New(fakePayload("bbb")))

		it.Finish()

		// Then
		assert.True(t, tt.Failed())
	})
}
