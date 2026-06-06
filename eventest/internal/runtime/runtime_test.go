package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/eventest/internal/runtime"
	mocksevent "github.com/thomas-marquis/it-happened/mocks/event"
	"go.uber.org/mock/gomock"
)

type fakePayload string

func (fakePayload) Type() event.Type {
	return "fake"
}

func TestRuntime(t *testing.T) {
	t.Run("should publish events according to the timeline", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		call1 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("abc")))
		call2 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("def")))
		call3 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("ghi")))

		gomock.InOrder(call1, call2, call3)

		tl := runtime.NewRuntime(mockBus, map[string]event.Payload{
			"a": fakePayload("abc"),
			"b": fakePayload("def"),
			"c": fakePayload("ghi"),
		})

		// When
		err := tl.Run("abc")

		// Then
		assert.NoError(t, err)
	})

	t.Run("should publish events with grouped events", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		call1 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("abc")))
		call2 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("def")))
		call3 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("ghi")))
		call4 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("de")))

		callX := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("xy")))
		callY := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("yz")))

		gomock.InOrder(call1, call2, callX, call3, call4)
		gomock.InOrder(call1, call2, callY, call3, call4)

		tl := runtime.NewRuntime(mockBus, map[string]event.Payload{
			"a": fakePayload("abc"),
			"b": fakePayload("def"),
			"c": fakePayload("ghi"),
			"d": fakePayload("de"),
			"x": fakePayload("xy"),
			"y": fakePayload("yz"),
		})

		// When
		err := tl.Run("[ab(xy)cd]")

		// Then
		assert.NoError(t, err)
	})

	t.Run("should use a default payload when none is provided", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		call1 := mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("a")))
		call2 := mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("b")))
		call3 := mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("c")))

		gomock.InOrder(call1, call2, call3)

		tl := runtime.NewRuntime(mockBus, nil)

		// When
		err := tl.Run("abc")

		// Then
		assert.NoError(t, err)
	})
}
