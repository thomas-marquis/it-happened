package timeline_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/timeline"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestTimelineBuilder_BuildsCorrectTicks(t *testing.T) {
	t.Run("simple sequence", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("abc")
		assert.NoError(t, err)

		builder := timeline.NewTimelineBuilder(time.Millisecond, nil)

		// When
		ticks, err := builder.Build(node)
		assert.NoError(t, err)

		// Then
		assert.Len(t, ticks, 3)
		assert.Len(t, ticks[0].Ops, 1)
		assert.Len(t, ticks[1].Ops, 1)
		assert.Len(t, ticks[2].Ops, 1)
	})

	t.Run("ordered group", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("[abc]")
		assert.NoError(t, err)

		builder := timeline.NewTimelineBuilder(time.Millisecond, nil)

		// When
		ticks, err := builder.Build(node)
		assert.NoError(t, err)

		// Then
		assert.Len(t, ticks, 1)
		assert.Len(t, ticks[0].Ops, 5) // start, a, b, c, end
	})

	t.Run("mixed sequence and groups", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("a-[bc]-(de)")
		assert.NoError(t, err)

		builder := timeline.NewTimelineBuilder(time.Millisecond, nil)

		// When
		ticks, err := builder.Build(node)
		assert.NoError(t, err)

		// Then - should have: a, -, [bc], -, (de) = 5 ticks
		assert.Len(t, ticks, 5)
	})

	t.Run("nested groups", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("[a(bc)d]")
		assert.NoError(t, err)

		builder := timeline.NewTimelineBuilder(time.Millisecond, nil)

		// When
		ticks, err := builder.Build(node)
		assert.NoError(t, err)

		// Then
		assert.Len(t, ticks, 1)
		// The group [a(bc)d] should be in one tick
		assert.True(t, len(ticks[0].Ops) > 3, "Expected multiple ops in group tick")
	})

	t.Run("followup events", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("a<-b")
		assert.NoError(t, err)

		builder := timeline.NewTimelineBuilder(time.Millisecond, nil)

		// When
		ticks, err := builder.Build(node)
		assert.NoError(t, err)

		// Then
		assert.Len(t, ticks, 1)
		// Should contain a followup op
		assert.IsType(t, marble.EventWithFollowupOp{}, ticks[0].Ops[0])
	})
}

func TestTimelineBuilder_EquivalenceWithNewTimeline(t *testing.T) {
	t.Run("should produce equivalent structure for simple sequence", func(t *testing.T) {
		// Given
		marbleStr := "abc"

		// When
		ops, _ := marble.Parse(marbleStr)
		tl1 := timeline.NewTimelineFromOps(ops)

		node, _ := marble.ParseAsNode(marbleStr)
		tl2 := timeline.NewTimeline(node)

		// Then - check both produce same number of ticks
		ticks1 := tl1.Ticks()
		ticks2 := tl2.Ticks()
		assert.Len(t, ticks1, 3)
		assert.Len(t, ticks2, 3)
	})

	t.Run("should produce equivalent structure for ordered group", func(t *testing.T) {
		// Given
		marbleStr := "[abc]"

		// When
		ops, _ := marble.Parse(marbleStr)
		tl1 := timeline.NewTimelineFromOps(ops)

		node, _ := marble.ParseAsNode(marbleStr)
		tl2 := timeline.NewTimeline(node)

		// Then - check both produce 1 tick with group
		ticks1 := tl1.Ticks()
		ticks2 := tl2.Ticks()
		assert.Len(t, ticks1, 1)
		assert.Len(t, ticks2, 1)
		// Both should have group start and end markers
		assert.Equal(t, 5, len(ticks1[0].Ops)) // start, a, b, c, end
		assert.Equal(t, 5, len(ticks2[0].Ops))
	})

	t.Run("should produce equivalent structure for mixed sequence", func(t *testing.T) {
		// Given
		marbleStr := "a-[bc]-(de)"

		// When
		ops, _ := marble.Parse(marbleStr)
		tl1 := timeline.NewTimelineFromOps(ops)

		node, _ := marble.ParseAsNode(marbleStr)
		tl2 := timeline.NewTimeline(node)

		// Then - check both produce same number of ticks
		ticks1 := tl1.Ticks()
		ticks2 := tl2.Ticks()
		assert.Len(t, ticks1, 5)
		assert.Len(t, ticks2, 5)
	})

	t.Run("should produce equivalent structure for nested groups", func(t *testing.T) {
		// Given
		marbleStr := "[a(bc)d]"

		// When
		ops, _ := marble.Parse(marbleStr)
		tl1 := timeline.NewTimelineFromOps(ops)

		node, _ := marble.ParseAsNode(marbleStr)
		tl2 := timeline.NewTimeline(node)

		// Then - check both produce 1 tick (group with nested group inside)
		ticks1 := tl1.Ticks()
		ticks2 := tl2.Ticks()
		assert.Len(t, ticks1, 1)
		assert.Len(t, ticks2, 1)
	})
}
