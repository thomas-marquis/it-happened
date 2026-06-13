package marble_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

func TestNodeCreation(t *testing.T) {
	t.Run("should create EventNode", func(t *testing.T) {
		node := &marble.EventNode{Name: "test"}
		assert.NotNil(t, node)
		assert.Equal(t, "test", node.Name)
	})

	t.Run("should create GroupNode", func(t *testing.T) {
		node := &marble.GroupNode{
			Ordered: true,
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.EventNode{Name: "b"},
			},
		}
		assert.NotNil(t, node)
		assert.True(t, node.Ordered)
		assert.Len(t, node.Children, 2)
	})
}

func TestOpToNodeConversion(t *testing.T) {
	t.Run("should convert EventOp to EventNode", func(t *testing.T) {
		op := marble.EventOp{Name: "test"}
		node := op.ToNode()
		assert.IsType(t, &marble.EventNode{}, node)
		assert.Equal(t, "test", node.(*marble.EventNode).Name)
	})
}

func TestToOpList(t *testing.T) {
	t.Run("should convert SequenceNode to Op list", func(t *testing.T) {
		node := &marble.SequenceNode{
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.WaitNode{},
				&marble.EventNode{Name: "b"},
			},
		}
		ops := marble.ToOpList(node)
		assert.Len(t, ops, 3)
		assert.Equal(t, marble.EventOp{Name: "a"}, ops[0])
		assert.Equal(t, marble.WaitOp{}, ops[1])
		assert.Equal(t, marble.EventOp{Name: "b"}, ops[2])
	})

	t.Run("should convert GroupNode to Op list", func(t *testing.T) {
		node := &marble.GroupNode{
			Ordered: true,
			Children: []marble.Node{
				&marble.EventNode{Name: "a"},
				&marble.EventNode{Name: "b"},
			},
		}
		ops := marble.ToOpList(node)
		assert.Len(t, ops, 4)
		assert.Equal(t, marble.OrderedGroupStartOp{EndPos: 3}, ops[0])
		assert.Equal(t, marble.EventOp{Name: "a"}, ops[1])
		assert.Equal(t, marble.EventOp{Name: "b"}, ops[2])
		assert.Equal(t, marble.OrderedGroupEndOp{StartPos: 0}, ops[3])
	})
}
