package eventest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSingleTickGroupRule(t *testing.T) {
	t.Run("should fail when wait op is inside a group", func(t *testing.T) {
		// Given
		rule := singleTickGroupRule{}
		ops := []op{
			orderedGroupOp{Ops: []op{waitOp{Duration: time.Second}}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a group is a single tick operation so a wait operator can't be used here")
	})

	t.Run("should fail when wait op is inside a nested group", func(t *testing.T) {
		// Given
		rule := singleTickGroupRule{}
		ops := []op{
			unorderedGroupOp{Ops: []op{
				orderedGroupOp{Ops: []op{waitOp{Duration: time.Second}}},
			}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
	})

	t.Run("should pass when no wait op is inside groups", func(t *testing.T) {
		// Given
		rule := singleTickGroupRule{}
		ops := []op{
			waitOp{Duration: time.Second},
			orderedGroupOp{Ops: []op{eventOp{Name: "a"}}},
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
		rule := notEmptyRule{}
		ops := []op{}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline cannot be empty")
	})

	t.Run("should pass on non-empty sequence", func(t *testing.T) {
		// Given
		rule := notEmptyRule{}
		ops := []op{eventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestStartEventAtBeginningRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := startEventAtBeginningRule{}
		ops := []op{startEventOp{}, startEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if group contains multiple start events", func(t *testing.T) {
		// Given
		rule := startEventAtBeginningRule{}
		ops := []op{
			orderedGroupOp{Ops: []op{startEventOp{}, startEventOp{}}},
		}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if start event is not at the beginning", func(t *testing.T) {
		// Given
		rule := startEventAtBeginningRule{}
		ops := []op{eventOp{Name: "a"}, startEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "the start event must be at the beginning of the timeline")
	})

	t.Run("should pass if start event is at the beginning", func(t *testing.T) {
		// Given
		rule := startEventAtBeginningRule{}
		ops := []op{startEventOp{}, eventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := startEventAtBeginningRule{}
		ops := []op{eventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestStartEventAnywhereRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := startEventAnywhereRule{}
		ops := []op{startEventOp{}, startEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should pass if start event is anywhere", func(t *testing.T) {
		// Given
		rule := startEventAnywhereRule{}
		ops := []op{eventOp{Name: "a"}, startEventOp{}, eventOp{Name: "b"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := startEventAnywhereRule{}
		ops := []op{eventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}

func TestUniqueStartEventRule(t *testing.T) {
	t.Run("should fail if no start event", func(t *testing.T) {
		// Given
		rule := uniqueStartEventRule{}
		ops := []op{eventOp{Name: "a"}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := uniqueStartEventRule{}
		ops := []op{startEventOp{}, startEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.ErrorIs(t, err, ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should pass if exactly one start event", func(t *testing.T) {
		// Given
		rule := uniqueStartEventRule{}
		ops := []op{eventOp{Name: "a"}, startEventOp{}}

		// When
		err := rule.Validate(ops)

		// Then
		assert.NoError(t, err)
	})
}
