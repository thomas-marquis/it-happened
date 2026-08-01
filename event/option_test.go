package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

func TestWithContext(t *testing.T) {
	t.Run("should set custom context", func(t *testing.T) {
		// Given
		ctx := context.Background()

		// When
		evt := event.New(fakePayload("test"),
			event.WithContext(ctx),
		)

		// Then
		assert.Equal(t, ctx, evt.Context(), "context should be set")
	})

	t.Run("should return a default background context by default", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))

		// When
		ctx := evt.Context()

		// Then
		assert.Equal(t, context.Background(), ctx, "default context should be background")
	})

	t.Run("should return a default background context when a nil context is set", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"),
			event.WithContext(nil))

		// When
		ctx := evt.Context()

		// Then
		assert.Equal(t, context.Background(), ctx, "default context should be background")
	})
}

func TestWithRef(t *testing.T) {
	t.Run("should set custom ref", func(t *testing.T) {
		// Given
		ref := "test-ref"

		// When
		evt := event.New(fakePayload("test"),
			event.WithRef(ref),
		)

		// Then
		assert.Equal(t, ref, evt.ChainRef(), "chain ref should be set")
	})

	t.Run("should create a default ref equal to the event ID when event is the first of its chain", func(t *testing.T) {
		// When
		evt := event.New(fakePayload("test"))

		// Then
		assert.Equal(t, evt.ID(), evt.ChainRef())
	})
}

func TestWithPriority(t *testing.T) {
	t.Run("should return a 0 priority by default", func(t *testing.T) {
		// When
		evt := event.New(fakePayload2{})

		// Then
		assert.Equal(t, 0, evt.Priority())
	})

	t.Run("should set a custom priority", func(t *testing.T) {
		// When
		evt := event.New(fakePayload2{},
			event.WithPriority(10))

		// Then
		assert.Equal(t, 10, evt.Priority())
	})
}
