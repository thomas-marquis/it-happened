package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/runtime"
	"github.com/thomas-marquis/it-happened/internal/mocks/event"
	"go.uber.org/mock/gomock"
)

func TestWithEventsMapping(t *testing.T) {
	t.Run("should panic when an event as already be set in the payloads mapping", func(t *testing.T) {
		// Given
		plMap := map[string]event.Payload{
			"a": fakePayload("a"),
		}
		evtMap := map[string]event.Event{
			"a": event.New(fakePayload("a")),
		}

		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		// When & Then
		assert.PanicsWithValue(t, "the event 'a' has already been defined as a payload", func() {
			runtime.NewRuntime(mockBus,
				runtime.WithPayloadsMapping(plMap),
				runtime.WithEventsMapping(evtMap), // <- panic!!
			)
		})
	})
}

func TestWithPayloadsMapping(t *testing.T) {
	t.Run("should panic when an event as already be set in the events mapping", func(t *testing.T) {
		// Given
		evtMap := map[string]event.Event{
			"a": event.New(fakePayload("a")),
		}
		plMap := map[string]event.Payload{
			"a": fakePayload("a"),
		}

		ctrl := gomock.NewController(t)
		mockBus := mocksevent.NewMockBus(ctrl)

		// When & Then
		assert.PanicsWithValue(t, "the payload corresponding to the event 'a' has already been defined as an event", func() {
			runtime.NewRuntime(mockBus,
				runtime.WithEventsMapping(evtMap),
				runtime.WithPayloadsMapping(plMap), // <- panic!!
			)
		})
	})
}
