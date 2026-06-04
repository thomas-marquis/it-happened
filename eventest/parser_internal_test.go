package eventest

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	fakeDuration = 10 * time.Millisecond
)

func TestMarbleParser(t *testing.T) {
	for _, tc := range []struct {
		marble string
		ops    []op
	}{
		{
			marble: "-",
			ops: []op{
				waitOp{Duration: fakeDuration},
			},
		},
		{
			marble: " -	_________       -",
			ops: []op{
				waitOp{Duration: fakeDuration},
				waitOp{Duration: fakeDuration},
				waitOp{Duration: fakeDuration},
			},
		},
		{
			marble: "____ _____",
			ops: []op{
				waitOp{Duration: fakeDuration},
				waitOp{Duration: fakeDuration},
			},
		},
		{
			marble: "^abc",
			ops: []op{
				startEventOp{},
				eventOp{Name: "a"},
				eventOp{Name: "b"},
				eventOp{Name: "c"},
			},
		}, {
			marble: "a-b/toto/lolo- c",
			ops: []op{
				eventOp{Name: "a"},
				waitOp{Duration: fakeDuration},
				eventOp{Name: "b"},
				eventOp{Name: "toto"},
				eventOp{Name: "lolo"},
				waitOp{Duration: fakeDuration},
				eventOp{Name: "c"},
			},
		},
		{
			marble: "(abc)",
			ops: []op{
				unorderedGroupOp{Ops: []op{
					eventOp{Name: "a"},
					eventOp{Name: "b"},
					eventOp{Name: "c"},
				}},
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
			ops: []op{
				orderedGroupOp{Ops: []op{
					unorderedGroupOp{Ops: []op{
						unorderedGroupOp{Ops: []op{
							eventOp{Name: "a"},
							eventOp{Name: "b"},
							eventOp{Name: "c"},
							orderedGroupOp{Ops: []op{
								eventOp{Name: "x"},
								eventOp{Name: "y"},
							}},
						}},
						eventOp{Name: "d"},
					}},
					waitOp{Duration: fakeDuration},
				}},
				eventOp{Name: "f"},
			},
		},
	} {
		t.Run(tc.marble, func(t *testing.T) {
			res, err := parseMarble(tc.marble, fakeDuration)
			assert.NoError(t, err)
			assert.Equal(t, tc.ops, res)
		})
	}

	t.Run("should return an error when the sequence is empty", func(t *testing.T) {
		_, err := parseMarble("", fakeDuration)
		assert.ErrorIs(t, err, ErrEmptyMarble)
		assert.ErrorIs(t, err, ErrMarbleSyntax)
	})
}
