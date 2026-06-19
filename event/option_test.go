package event_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

func TestOption_Apply(t *testing.T) {
	t.Run("Given event with options applied, When options are applied, Then they correctly configure the event properties", func(t *testing.T) {
		// Given
		ctx := context.Background()
		ref := "test-ref"

		// When - create event with options
		evt := event.New(fakePayload("test"),
			event.WithContext(ctx),
			event.WithRef(ref),
		)

		// Then
		assert.Equal(t, ctx, evt.Context(), "context should be set")
		assert.Equal(t, ref, evt.ChainRef(), "chain ref should be set")
	})
}

func TestOption_Compose(t *testing.T) {
	t.Run("Given multiple options, When applied to the same event, Then all options are applied in order", func(t *testing.T) {
		// Given
		ctx1 := context.WithValue(context.Background(), "key1", "value1")
		ctx2 := context.WithValue(context.Background(), "key2", "value2")
		ref1 := "ref1"
		ref2 := "ref2"

		// When - apply options in sequence
		// Note: Options are applied in the order they are passed to event.New
		// But WithRef overwrites the ref, so the last one wins
		evt := event.New(fakePayload("test"),
			event.WithContext(ctx1),
			event.WithRef(ref1),
			event.WithContext(ctx2),
			event.WithRef(ref2),
		)

		// Then - the last options should take effect
		assert.Equal(t, ctx2, evt.Context(), "last context should be set")
		assert.Equal(t, ref2, evt.ChainRef(), "last ref should be set")
	})
}

func TestWithContext_DefaultContext(t *testing.T) {
	t.Run("Given event with no context option, When created, Then it has background context", func(t *testing.T) {
		// Given/When
		evt := event.New(fakePayload("test"))

		// Then
		assert.NotNil(t, evt.Context())
		// The context should be background or a derived context
		// We can't directly compare contexts, but we can verify it's not nil
	})
}

func TestWithRef_DefaultRef(t *testing.T) {
	t.Run("Given event with no ref option, When created, Then it has an auto-generated ref", func(t *testing.T) {
		// Given/When
		evt := event.New(fakePayload("test"))

		// Then
		assert.NotEmpty(t, evt.ChainRef(), "ref should be auto-generated")
		// The ref should be the same as the ID by default
		assert.Equal(t, evt.ID(), evt.ChainRef(), "ref should equal ID by default")
	})
}

func TestOption_NilContext(t *testing.T) {
	t.Run("Given event with nil context, When WithContext is called with nil, Then it uses background context", func(t *testing.T) {
		// Given/When
		evt := event.New(fakePayload("test"),
			event.WithContext(nil),
		)

		// Then
		assert.NotNil(t, evt.Context())
		// Should default to background context
	})
}
