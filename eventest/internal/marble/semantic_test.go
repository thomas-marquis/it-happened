package marble_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestSingleTickGroupRule(t *testing.T) {
	t.Run("should fail when wait op is inside a group", func(t *testing.T) {
		// Given
		rule := marble.SingleTickGroupRule{}
		ops := []marble.Op{
			marble.OrderedGroupOp{Ops: []marble.Op{marble.WaitOp{Duration: time.Second}}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a group is a single tick operation so a wait operator can't be used here")
	})

	t.Run("should fail when wait op is inside a nested group", func(t *testing.T) {
		// Given
		rule := marble.SingleTickGroupRule{}
		ops := []marble.Op{
			marble.UnorderedGroupOp{Ops: []marble.Op{
				marble.OrderedGroupOp{Ops: []marble.Op{marble.WaitOp{Duration: time.Second}}},
			}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
	})

	t.Run("should pass when no wait op is inside groups", func(t *testing.T) {
		// Given
		rule := marble.SingleTickGroupRule{}
		ops := []marble.Op{
			marble.WaitOp{Duration: time.Second},
			marble.OrderedGroupOp{Ops: []marble.Op{marble.EventOp{Name: "a"}}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestNotEmptyRule(t *testing.T) {
	t.Run("should fail on empty sequence", func(t *testing.T) {
		// Given
		rule := marble.NotEmptyRule{}
		ops := []marble.Op{}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline cannot be empty")
	})

	t.Run("should pass on non-empty sequence", func(t *testing.T) {
		// Given
		rule := marble.NotEmptyRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestStartEventAtBeginningRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		ops := []marble.Op{marble.StartEventOp{}, marble.StartEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if group contains multiple start events", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		ops := []marble.Op{
			marble.OrderedGroupOp{Ops: []marble.Op{marble.StartEventOp{}, marble.StartEventOp{}}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if start event is not at the beginning", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}, marble.StartEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "the start event must be at the beginning of the timeline")
	})

	t.Run("should pass if start event is at the beginning", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		ops := []marble.Op{marble.StartEventOp{}, marble.EventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestStartEventAnywhereRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		ops := []marble.Op{marble.StartEventOp{}, marble.StartEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should pass if start event is anywhere", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}, marble.StartEventOp{}, marble.EventOp{Name: "b"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestUniqueStartEventRule(t *testing.T) {
	t.Run("should fail if no start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		ops := []marble.Op{marble.StartEventOp{}, marble.StartEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should pass if exactly one start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		ops := []marble.Op{marble.EventOp{Name: "a"}, marble.StartEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}
