package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
	"github.com/thomas-marquis/it-happened/eventest/internal/runtime"
)

func TestTimeline(t *testing.T) {
	t.Run("should build a simple timeline with one tick per event", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
			marble.EventOp{Name: "c"},
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
				},
			},
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "b"},
				},
			},
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "c"},
				},
			},
		}

		// Then
		tl := runtime.NewTimeline(ops)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should gather grouped events within the same tick", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.OrderedGroupStartOp{EndPos: 3}, // 0
			marble.EventOp{Name: "a"},             // 1
			marble.EventOp{Name: "b"},             // 2
			marble.OrderedGroupEndOp{StartPos: 0}, // 3
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
					marble.EventOp{Name: "b"},
				},
			},
		}

		// When
		tl := runtime.NewTimeline(ops)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should gather nested grouped events within the same tick", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.OrderedGroupStartOp{EndPos: 7},  // 0
			marble.EventOp{Name: "a"},              // 1
			marble.OrderedGroupStartOp{EndPos: 5},  // 2
			marble.EventOp{Name: "x"},              // 3
			marble.EventOp{Name: "y"},              // 4
			marble.OrderedGroupEndOp{StartPos: 2},  // 5
			marble.EventOp{Name: "b"},              // 6
			marble.OrderedGroupEndOp{StartPos: 0},  // 7
			marble.OrderedGroupStartOp{EndPos: 11}, // 8
			marble.EventOp{Name: "l"},              // 9
			marble.EventOp{Name: "m"},              // 10
			marble.OrderedGroupEndOp{StartPos: 8},  // 11
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
					marble.EventOp{Name: "x"},
					marble.EventOp{Name: "y"},
					marble.EventOp{Name: "b"},
				},
			},
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "l"},
					marble.EventOp{Name: "m"},
				},
			},
		}

		// When
		tl := runtime.NewTimeline(ops)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should shuffle unordered grouped events within the same tick", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.UnorderedGroupStartOp{EndPos: 6}, // 0
			marble.EventOp{Name: "a"},               // 1
			marble.EventOp{Name: "b"},               // 2
			marble.EventOp{Name: "c"},               // 3
			marble.EventOp{Name: "d"},               // 4
			marble.EventOp{Name: "e"},               // 5
			marble.UnorderedGroupEndOp{StartPos: 0}, // 6
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "e"},
					marble.EventOp{Name: "a"},
					marble.EventOp{Name: "c"},
					marble.EventOp{Name: "b"},
					marble.EventOp{Name: "d"},
				},
			},
		}

		// When
		tl := runtime.NewTimeline(ops, runtime.TimelineWithSeed(42))
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should keep order for ordered grouped events within the same tick", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.OrderedGroupStartOp{EndPos: 6}, // 0
			marble.EventOp{Name: "a"},             // 1
			marble.EventOp{Name: "b"},             // 2
			marble.EventOp{Name: "c"},             // 3
			marble.EventOp{Name: "d"},             // 4
			marble.EventOp{Name: "e"},             // 5
			marble.OrderedGroupEndOp{StartPos: 0}, // 6
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
					marble.EventOp{Name: "b"},
					marble.EventOp{Name: "c"},
					marble.EventOp{Name: "d"},
					marble.EventOp{Name: "e"},
				},
			},
		}

		// When
		tl := runtime.NewTimeline(ops)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})

	t.Run("should handle event with followup", func(t *testing.T) {
		// Given
		ops := []marble.Op{
			marble.EventOp{Name: "a"},                                 // 0
			marble.EventWithFollowupOp{EventName: "b", From: "prev"},  // 1
			marble.OrderedGroupStartOp{EndPos: 6},                     // 2
			marble.EventOp{Name: "c"},                                 // 3
			marble.EventWithFollowupOp{EventName: "d", From: "prev2"}, // 4
			marble.EventOp{Name: "e"},                                 // 5
			marble.OrderedGroupEndOp{StartPos: 2},                     // 6
		}

		expected := []runtime.Tick{
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "a"},
				},
			},
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventWithFollowupOp{EventName: "b", From: "prev"},
				},
			},
			{
				Duration: runtime.DefaultTickDuration,
				Ops: []marble.Op{
					marble.EventOp{Name: "c"},
					marble.EventWithFollowupOp{EventName: "d", From: "prev2"},
					marble.EventOp{Name: "e"},
				},
			},
		}

		// When
		tl := runtime.NewTimeline(ops)
		res := tl.Ticks()

		// Then
		assert.Equal(t, expected, res)
	})
}
