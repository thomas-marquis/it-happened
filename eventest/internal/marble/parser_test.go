package marble_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

const (
	fakeDuration = 10 * time.Millisecond
)

func TestMarbleParser(t *testing.T) {
	for _, tc := range []struct {
		marble string
		ops    []marble.Op
	}{
		{
			marble: "-",
			ops: []marble.Op{
				marble.WaitOp{Duration: fakeDuration},
			},
		},
		{
			marble: " -	_________       -",
			ops: []marble.Op{
				marble.WaitOp{Duration: fakeDuration},
				marble.WaitOp{Duration: fakeDuration},
				marble.WaitOp{Duration: fakeDuration},
			},
		},
		{
			marble: "____ _____",
			ops: []marble.Op{
				marble.WaitOp{Duration: fakeDuration},
				marble.WaitOp{Duration: fakeDuration},
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
		}, {
			marble: "a-b/toto/lolo- c",
			ops: []marble.Op{
				marble.EventOp{Name: "a"},
				marble.WaitOp{Duration: fakeDuration},
				marble.EventOp{Name: "b"},
				marble.EventOp{Name: "toto"},
				marble.EventOp{Name: "lolo"},
				marble.WaitOp{Duration: fakeDuration},
				marble.EventOp{Name: "c"},
			},
		},
		{
			marble: "(abc)",
			ops: []marble.Op{
				marble.UnorderedGroupOp{Ops: []marble.Op{
					marble.EventOp{Name: "a"},
					marble.EventOp{Name: "b"},
					marble.EventOp{Name: "c"},
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
			ops: []marble.Op{
				marble.OrderedGroupOp{Ops: []marble.Op{
					marble.UnorderedGroupOp{Ops: []marble.Op{
						marble.UnorderedGroupOp{Ops: []marble.Op{
							marble.EventOp{Name: "a"},
							marble.EventOp{Name: "b"},
							marble.EventOp{Name: "c"},
							marble.OrderedGroupOp{Ops: []marble.Op{
								marble.EventOp{Name: "x"},
								marble.EventOp{Name: "y"},
							}},
						}},
						marble.EventOp{Name: "d"},
					}},
					marble.WaitOp{Duration: fakeDuration},
				}},
				marble.EventOp{Name: "f"},
			},
		},
		{
			marble: "a<-b",
			ops: []marble.Op{
				marble.EventWithFollowupOp{EventName: "a", From: "b"},
			},
		},
	} {
		t.Run(tc.marble, func(t *testing.T) {
			res, err := marble.Parse(tc.marble, fakeDuration)
			assert.NoError(t, err)
			assert.Equal(t, tc.ops, res)
		})
	}

	t.Run("should return an error when the sequence is empty", func(t *testing.T) {
		_, err := marble.Parse("", fakeDuration)
		assert.ErrorIs(t, err, marble.ErrEmptyMarble)
		assert.ErrorIs(t, err, marble.ErrMarbleSyntax)
	})
}
