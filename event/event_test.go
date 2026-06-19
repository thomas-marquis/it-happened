package event_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/event"
)

func TestType_String(t *testing.T) {
	t.Run("Given event Type, When String is called, Then it returns the string representation", func(t *testing.T) {
		// Given
		eventType := event.Type("test.payload")

		// When
		str := eventType.String()

		// Then
		assert.Equal(t, "test.payload", str)
	})
}

func TestEvent_Payload(t *testing.T) {
	t.Run("Given event, When Payload is called, Then it returns the event payload", func(t *testing.T) {
		// Given
		payload := fakePayload("test")
		evt := event.New(payload)

		// When
		retrieved := evt.Payload()

		// Then
		assert.Equal(t, payload, retrieved)
	})
}

func TestEvent_NewFollowup(t *testing.T) {
	t.Run("Given parent event, When NewFollowup is called, Then it creates a followup event", func(t *testing.T) {
		// Given
		parent := event.New(fakePayload("parent"))

		// When
		followup := parent.NewFollowup(fakePayload("followup"))

		// Then
		assert.NotNil(t, followup)
		assert.Equal(t, parent.ChainRef(), followup.ChainRef())
		assert.Equal(t, uint(1), followup.ChainPosition())
		assert.NotEqual(t, parent.ID(), followup.ID())
	})
}

func TestEvent_Type(t *testing.T) {
	t.Run("Given event, When Type is called, Then it returns the event type", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))

		// When
		eventType := evt.Type()

		// Then
		assert.Equal(t, event.Type("fake.payload"), eventType)
	})
}

func TestEvent_ChainRef(t *testing.T) {
	t.Run("Given event, When ChainRef is called, Then it returns the chain reference", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))

		// When
		chainRef := evt.ChainRef()

		// Then
		assert.NotEmpty(t, chainRef)
		// ChainRef should be the same as ID for non-followup events
		assert.Equal(t, evt.ID(), chainRef)
	})
}

func TestEvent_ChainPosition(t *testing.T) {
	t.Run("Given regular event, When ChainPosition is called, Then it returns 0", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))

		// When
		position := evt.ChainPosition()

		// Then
		assert.Equal(t, uint(0), position)
	})

	t.Run("Given followup event, When ChainPosition is called, Then it returns position > 0", func(t *testing.T) {
		// Given
		parent := event.New(fakePayload("parent"))
		followup := parent.NewFollowup(fakePayload("followup"))

		// When
		position := followup.ChainPosition()

		// Then
		assert.Equal(t, uint(1), position)
	})
}

func TestEvent_ID(t *testing.T) {
	t.Run("Given event, When ID is called, Then it returns a unique identifier", func(t *testing.T) {
		// Given
		evt1 := event.New(fakePayload("test1"))
		evt2 := event.New(fakePayload("test2"))

		// When
		id1 := evt1.ID()
		id2 := evt2.ID()

		// Then
		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
	})
}

func TestEvent_Context(t *testing.T) {
	t.Run("Given event with default context, When Context is called, Then it returns background context", func(t *testing.T) {
		// Given
		evt := event.New(fakePayload("test"))

		// When
		ctx := evt.Context()

		// Then
		assert.NotNil(t, ctx)
	})
}
