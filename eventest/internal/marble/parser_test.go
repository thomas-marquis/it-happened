package marble_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestMarbleParser(t *testing.T) {
	for _, tc := range []struct {
		marble string
		ops    []marble.Op
	}{
		{
			marble: "-",
			ops: []marble.Op{
				marble.WaitOp{},
			},
		},
		{
			marble: " -	_________       -",
			ops: []marble.Op{
				marble.WaitOp{},
				marble.WaitOp{},
				marble.WaitOp{},
			},
		},
		{
			marble: "____ _____",
			ops: []marble.Op{
				marble.WaitOp{},
				marble.WaitOp{},
			},
		},
		{
			marble: "^abc",
			ops: []marble.Op{
				marble.StartEventOp{},
				marble.EventOp{Name: "a"},
				marble.EventOp{Name: "b"},
				marble.EventOp{Name: "c"},
			},
		},
		{
			marble: "a-b/toto/lolo- c",
			ops: []marble.Op{
				marble.EventOp{Name: "a"},
				marble.WaitOp{},
				marble.EventOp{Name: "b"},
				marble.EventOp{Name: "toto"},
				marble.EventOp{Name: "lolo"},
				marble.WaitOp{},
				marble.EventOp{Name: "c"},
			},
		},
		{
			marble: "(abc)",
			ops: []marble.Op{
				marble.UnorderedGroupStartOp{EndPos: 4}, // 0 (
				marble.EventOp{Name: "a"},               // 1 a
				marble.EventOp{Name: "b"},               // 2 b
				marble.EventOp{Name: "c"},               // 3 c
				marble.UnorderedGroupEndOp{StartPos: 0}, // 4 )
			},
		},
		{
			marble: `[
						(
							(
								abc
								[x y]
							)
							d
						)
						-
					]
					/f`,
			ops: []marble.Op{
				marble.OrderedGroupStartOp{EndPos: 14},   // 0  [
				marble.UnorderedGroupStartOp{EndPos: 12}, // 1  (
				marble.UnorderedGroupStartOp{EndPos: 10}, // 2  (
				marble.EventOp{Name: "a"},                // 3  a
				marble.EventOp{Name: "b"},                // 4  b
				marble.EventOp{Name: "c"},                // 5  c
				marble.OrderedGroupStartOp{EndPos: 9},    // 6  [
				marble.EventOp{Name: "x"},                // 7  x
				marble.EventOp{Name: "y"},                // 8  y
				marble.OrderedGroupEndOp{StartPos: 6},    // 9  ]
				marble.UnorderedGroupEndOp{StartPos: 2},  // 10 )
				marble.EventOp{Name: "d"},                // 11 d
				marble.UnorderedGroupEndOp{StartPos: 1},  // 12 )
				marble.WaitOp{},                          // 13 - <- ok, it's likely a semantic error, but I'm a parser, not a semantic validator, so I don't care...
				marble.OrderedGroupEndOp{StartPos: 0},    // 14 ]
				marble.EventOp{Name: "f"},                // 15 /f
			},
		},
		{
			marble: "a<-b",
			ops: []marble.Op{
				marble.EventWithFollowupOp{NewEvent: "a", OfEvent: "b"},
			},
		},
	} {
		t.Run(tc.marble, func(t *testing.T) {
			res, err := marble.Parse(tc.marble)
			assert.NoError(t, err)
			assert.Equal(t, tc.ops, res)
		})
	}

	t.Run("should return an error when the sequence is empty", func(t *testing.T) {
		_, err := marble.Parse("")
		assert.ErrorIs(t, err, marble.ErrEmptyMarble)
		assert.ErrorIs(t, err, marble.ErrMarbleSyntax)
	})
}
