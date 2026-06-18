package marble_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestSingleTickGroupRule(t *testing.T) {
	t.Run("should fail when wait op is inside a group", func(t *testing.T) {
		// Given
		rule := marble.WaitlessGroupsRule{}
		node, _ := marble.ParseAsNode("[-]")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a group is a single tick operation so a wait operator can't be used here")
	})

	t.Run("should fail when wait op is inside a nested group", func(t *testing.T) {
		// Given
		rule := marble.WaitlessGroupsRule{}
		node, _ := marble.ParseAsNode("([-])")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
	})

	t.Run("should pass when no wait op is inside groups", func(t *testing.T) {
		// Given
		rule := marble.WaitlessGroupsRule{}
		node, _ := marble.ParseAsNode("-[a]")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})
}

func TestNotEmptyRule(t *testing.T) {
	t.Run("should fail on empty sequence", func(t *testing.T) {
		// Given
		rule := marble.NotEmptyRule{}
		node := &marble.SequenceNode{}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline cannot be empty")
	})

	t.Run("should pass on non-empty sequence", func(t *testing.T) {
		// Given
		rule := marble.NotEmptyRule{}
		node, _ := marble.ParseAsNode("a")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})
}

func TestMandatoryInitEventRule(t *testing.T) {
	t.Run("should pass if expectation has exactly one initEvent at the beginning", func(t *testing.T) {
		// Given
		rule := marble.MandatoryInitEventRule{}
		node, _ := marble.ParseAsNode("^abc")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if initEvent is inside a group at the beginning", func(t *testing.T) {
		// Given
		rule := marble.MandatoryInitEventRule{}
		node, _ := marble.ParseAsNode("(^abc)")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should fail if no initEvent", func(t *testing.T) {
		// Given
		rule := marble.MandatoryInitEventRule{}
		node, _ := marble.ParseAsNode("abc")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "expectation must contain exactly one initEvent (^) at the beginning")
	})

	t.Run("should fail if more than one initEvent", func(t *testing.T) {
		// Given
		rule := marble.MandatoryInitEventRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.InitEventNode{},
				&marble.InitEventNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "expectation must contain exactly one initEvent (^)")
	})

	t.Run("should fail if initEvent is not at the beginning", func(t *testing.T) {
		// Given
		rule := marble.MandatoryInitEventRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.InitEventNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "initEvent (^) must be the first element in the expectation")
	})
}

func TestNoInitEventInSideEffectRule(t *testing.T) {
	t.Run("should pass if side effect has no initEvent", func(t *testing.T) {
		// Given
		rule := marble.NoInitEventInSideEffectRule{}
		node, _ := marble.ParseAsNode("abc")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should fail if side effect contains initEvent", func(t *testing.T) {
		// Given
		rule := marble.NoInitEventInSideEffectRule{}
		node, _ := marble.ParseAsNode("^abc")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "side effect must not contain initEvent (^)")
	})

	t.Run("should fail if side effect contains initEvent in group", func(t *testing.T) {
		// Given
		rule := marble.NoInitEventInSideEffectRule{}
		node, _ := marble.ParseAsNode("(^ab)")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "side effect must not contain initEvent (^)")
	})
}

func TestSideEffectDurationRule(t *testing.T) {
	t.Run("should pass if side effect duration <= expectation duration", func(t *testing.T) {
		// Given
		rule := marble.SideEffectDurationRule{ExpectedDuration: 5}
		node, _ := marble.ParseAsNode("abc") // 3 ticks

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should fail if side effect duration > expectation duration", func(t *testing.T) {
		// Given
		rule := marble.SideEffectDurationRule{ExpectedDuration: 2}
		node, _ := marble.ParseAsNode("abc") // 3 ticks

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "side effect duration (3 ticks) exceeds expectation duration (2 ticks)")
	})
}
