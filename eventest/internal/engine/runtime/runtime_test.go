package runtime_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/internal/mocks/event"
	"go.uber.org/mock/gomock"
)

type fakePayload string

func (fakePayload) Type() event.Type {
	return "fake"
}

func TestRuntime_RunAll(t *testing.T) {
	t.Run("should publish events according to the timeline", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		call1 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("abc")))
		call2 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("def")))
		call3 := mockBus.EXPECT().Publish(eventest.PayloadEq(fakePayload("ghi")))

		gomock.InOrder(call1, call2, call3)

		rt := runtime.NewRuntime(mockBus, runtime.WithPayloadsMapping(map[string]event.Payload{
			"a": fakePayload("abc"),
			"b": fakePayload("def"),
			"c": fakePayload("ghi"),
		}))

		// When
		err := rt.RunAll("abc")

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

		rt := runtime.NewRuntime(mockBus,
			runtime.WithEventsMapping(map[string]event.Event{
				"c": event.New(fakePayload("ghi")),
			}),
			runtime.WithPayloadsMapping(map[string]event.Payload{
				"a": fakePayload("abc"),
				"b": fakePayload("def"),
				"d": fakePayload("de"),
				"x": fakePayload("xy"),
				"y": fakePayload("yz"),
			}))

		// When
		err := rt.RunAll("[ab(xy)cd]")

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

		rt := runtime.NewRuntime(mockBus)

		// When
		err := rt.RunAll("abc")

		// Then
		assert.NoError(t, err)
	})
}

func TestRuntime_Run(t *testing.T) {
	t.Run("should publish events through multiple time ticks", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		rt := runtime.NewRuntime(mockBus, runtime.WithBaseTickDuration(1*time.Second))

		// When & Then
		sess, err := rt.Run("ab-[cd]-e")
		assert.NoError(t, err)

		clock := sess.Clock()
		assert.Equal(t, 0*time.Second, clock.Elapsed())

		// a
		mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("a")))
		assert.NoError(t, sess.Next())
		assert.Equal(t, 1*time.Second, clock.Elapsed())

		// b
		mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("b")))
		assert.NoError(t, sess.Next())
		assert.Equal(t, 2*time.Second, clock.Elapsed())

		// -
		mockBus.EXPECT().Publish(gomock.Any()).Times(0)
		assert.NoError(t, sess.Next())
		assert.Equal(t, 3*time.Second, clock.Elapsed())

		// [cd]
		gomock.InOrder(
			mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("c"))),
			mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("d"))),
		)
		assert.NoError(t, sess.Next())
		assert.Equal(t, 4*time.Second, clock.Elapsed())

		// -
		mockBus.EXPECT().Publish(gomock.Any()).Times(0)
		assert.NoError(t, sess.Next())
		assert.Equal(t, 5*time.Second, clock.Elapsed())

		// e
		mockBus.EXPECT().Publish(eventest.PayloadEq(runtime.DefaultPayload("e")))
		assert.NoError(t, sess.Next())
		assert.Equal(t, 6*time.Second, clock.Elapsed())

		assert.ErrorIs(t, sess.Next(), runtime.SessionEnded)
	})

	t.Run("should publish events with followup", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		evtA := event.New(fakePayload("abc"))
		evtX := event.New(fakePayload("xyz"))

		rt := runtime.NewRuntime(mockBus,
			runtime.WithEventsMapping(map[string]event.Event{
				"a": evtA,
				"x": evtX,
			}),
			runtime.WithBaseTickDuration(1*time.Second))

		// When & Then
		sess, err := rt.Run("b<-ay<-x")
		assert.NoError(t, err)

		// b<-a
		mockBus.EXPECT().Publish(
			gomock.All(
				eventest.IsFollowupOf(evtA),
				eventest.PayloadEq(runtime.DefaultPayload("b"))),
		)
		assert.NoError(t, sess.Next())

		mockBus.EXPECT().Publish(
			gomock.All(
				eventest.IsFollowupOf(evtX),
				eventest.PayloadEq(runtime.DefaultPayload("y"))),
		)
		assert.NoError(t, sess.Next())

		assert.ErrorIs(t, sess.Next(), runtime.SessionEnded)
	})

	t.Run("should publish the start event when provided", func(t *testing.T) {
		// Given
		//ctrl := gomock.NewController(t)
		//mockBus := mocksevent.NewMockBus(ctrl)

		//rt := runtime.NewRuntime(mockBus)
	})

	t.Run("should return an error when the start event is missing", func(t *testing.T) {

	})
}
