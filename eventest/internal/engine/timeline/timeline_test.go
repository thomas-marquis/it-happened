package timeline_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/engine/timeline"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestTimeline(t *testing.T) {
	t.Run("should build a simple timeline with one tick per event", func(t *testing.T) {
		// Given
		node, _ := marble.ParseAsNode("abc")

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "b"},
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "c"},
				},
			},
		}

		// Then
		tl := timeline.NewTimeline(node)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should gather grouped events within the same tick", func(t *testing.T) {
		// Given
		node, _ := marble.ParseAsNode("[ab]")

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 3}, // 0
					marble.EventOp{Name: "a"},             // 1
					marble.EventOp{Name: "b"},             // 2
					marble.OrderedGroupEndOp{StartPos: 0}, // 3
				},
			},
		}

		// When
		tl := timeline.NewTimeline(node)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should gather nested grouped events within the same tick", func(t *testing.T) {
		// Given
		node, _ := marble.ParseAsNode("[a[xy]b][lm]")

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 7}, // 0
					marble.EventOp{Name: "a"},             // 1
					marble.OrderedGroupStartOp{EndPos: 5}, // 2
					marble.EventOp{Name: "x"},             // 3
					marble.EventOp{Name: "y"},             // 4
					marble.OrderedGroupEndOp{StartPos: 2}, // 5
					marble.EventOp{Name: "b"},             // 6
					marble.OrderedGroupEndOp{StartPos: 0}, // 7
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 3}, // 0
					marble.EventOp{Name: "l"},             // 1
					marble.EventOp{Name: "m"},             // 2
					marble.OrderedGroupEndOp{StartPos: 0}, // 3
				},
			},
		}

		// When
		tl := timeline.NewTimeline(node)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should shuffle unordered grouped events within the same tick", func(t *testing.T) {
		// Given
		node, _ := marble.ParseAsNode("(abcde)(x)[qr[nm]s]")

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.UnorderedGroupStartOp{EndPos: 6}, // 0
					marble.EventOp{Name: "e"},               // 1
					marble.EventOp{Name: "a"},               // 2
					marble.EventOp{Name: "c"},               // 3
					marble.EventOp{Name: "b"},               // 4
					marble.EventOp{Name: "d"},               // 5
					marble.UnorderedGroupEndOp{StartPos: 0}, // 6
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.UnorderedGroupStartOp{EndPos: 2}, // 0
					marble.EventOp{Name: "x"},               // 1
					marble.UnorderedGroupEndOp{StartPos: 0}, // 2
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 8}, // 0
					marble.EventOp{Name: "q"},             // 1
					marble.EventOp{Name: "r"},             // 2
					marble.OrderedGroupStartOp{EndPos: 6}, // 3
					marble.EventOp{Name: "n"},             // 4
					marble.EventOp{Name: "m"},             // 5
					marble.OrderedGroupEndOp{StartPos: 3}, // 6
					marble.EventOp{Name: "s"},             // 7
					marble.OrderedGroupEndOp{StartPos: 0}, // 8
				},
			},
		}

		// When
		tl := timeline.NewTimeline(node, timeline.WithSeed(42))
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should keep order for ordered grouped events within the same tick", func(t *testing.T) {
		// Given
		node, _ := marble.ParseAsNode("[abcde]")

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 6}, // 0
					marble.EventOp{Name: "a"},             // 1
					marble.EventOp{Name: "b"},             // 2
					marble.EventOp{Name: "c"},             // 3
					marble.EventOp{Name: "d"},             // 4
					marble.EventOp{Name: "e"},             // 5
					marble.OrderedGroupEndOp{StartPos: 0}, // 6
				},
			},
		}

		// When
		tl := timeline.NewTimeline(node)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should handle event with followup", func(t *testing.T) {
		// Given
		node, err := marble.ParseAsNode("ab<-/prev [c d<-/prev2 e]-")
		assert.NoError(t, err)

		expected := []timeline.Tick{
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventWithFollowupOp{NewEvent: "b", OfEvent: "prev"},
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.OrderedGroupStartOp{EndPos: 4},                       // 0
					marble.EventOp{Name: "c"},                                   // 1
					marble.EventWithFollowupOp{NewEvent: "d", OfEvent: "prev2"}, // 2
					marble.EventOp{Name: "e"},                                   // 3
					marble.OrderedGroupEndOp{StartPos: 0},                       // 4
				},
			},
			{
				Duration: timeline.DefaultTickDuration,
				Ops: []marble.Op{
					marble.WaitOp{},
				},
			},
		}

		// When
		tl := timeline.NewTimeline(node)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})
}
