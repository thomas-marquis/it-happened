package carrier_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/carrier"
	"github.com/thomas-marquis/it-happened/event"
)

// TestCompletedOnFollowupReceived tests the default completion condition.
func TestCompletedOnFollowupReceived(t *testing.T) {
	t.Run("should returns true when called with followup events", func(t *testing.T) {
		// Given
		sent := event.New(testPayload("sent"))
		received := event.New(testPayload("received"))

		// When
		result := carrier.CompletedOnFollowupReceived(sent, received)

		// Then
		assert.True(t, result)
	})
}
