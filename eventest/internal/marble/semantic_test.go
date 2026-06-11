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

func TestStartEventAtBeginningRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		// We use a manual sequence node since ParseAsNode might fail on ^^
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.StartNode{},
				&marble.StartNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if group contains multiple start events", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.GroupNode{
					Ordered: true,
					Children: []marble.Node{
						&marble.StartNode{},
						&marble.StartNode{},
					},
				},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should fail if start event is not at the beginning", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.StartNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "the start event must be at the beginning of the timeline")
	})

	t.Run("should pass if start event is at the beginning", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		node, _ := marble.ParseAsNode("^a")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if start event is inside a group at the beginning", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		node, _ := marble.ParseAsNode("[^]a")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAtBeginningRule{}
		node, _ := marble.ParseAsNode("a")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})
}

func TestStartEventAnywhereRule(t *testing.T) {
	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.StartNode{},
				&marble.StartNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline can have at most one start event")
	})

	t.Run("should pass if start event is anywhere", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.StartNode{},
				&marble.EventNode{Name: "b"},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})

	t.Run("should pass if no start event", func(t *testing.T) {
		// Given
		rule := marble.StartEventAnywhereRule{}
		node, _ := marble.ParseAsNode("a")

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})
}

func TestUniqueStartEventRule(t *testing.T) {
	t.Run("should fail if no start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		node, _ := marble.ParseAsNode("a")

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should fail if more than one start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.StartNode{},
				&marble.StartNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.ErrorIs(t, err, marble.ErrSemantic)
		assert.Contains(t, err.Error(), "a timeline must have exactly one start event")
	})

	t.Run("should pass if exactly one start event", func(t *testing.T) {
		// Given
		rule := marble.UniqueStartEventRule{}
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.StartNode{},
			},
		}

		// When
		err := rule.Validate(node)

		// Then
		assert.NoError(t, err)
	})
}
