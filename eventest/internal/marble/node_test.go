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

func TestSequenceNodeFromOps(t *testing.T) {
	t.Run("should convert simple ops to SequenceNode", func(t *testing.T) {
		ops := []marble.Op{
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
			marble.EventOp{Name: "c"},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 3)
		assert.IsType(t, &marble.EventNode{}, node.Children[0])
		assert.Equal(t, "a", node.Children[0].(*marble.EventNode).Name)
		assert.IsType(t, &marble.EventNode{}, node.Children[1])
		assert.Equal(t, "b", node.Children[1].(*marble.EventNode).Name)
		assert.IsType(t, &marble.EventNode{}, node.Children[2])
		assert.Equal(t, "c", node.Children[2].(*marble.EventNode).Name)
	})

	t.Run("should convert ops with wait to SequenceNode", func(t *testing.T) {
		ops := []marble.Op{
			marble.EventOp{Name: "a"},
			marble.WaitOp{},
			marble.EventOp{Name: "b"},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 3)
		assert.IsType(t, &marble.EventNode{}, node.Children[0])
		assert.IsType(t, &marble.WaitNode{}, node.Children[1])
		assert.IsType(t, &marble.EventNode{}, node.Children[2])
	})

	t.Run("should convert ordered group ops to GroupNode", func(t *testing.T) {
		ops := []marble.Op{
			marble.OrderedGroupStartOp{EndPos: 3},
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
			marble.OrderedGroupEndOp{StartPos: 0},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 1)
		group := node.Children[0].(*marble.GroupNode)
		assert.True(t, group.Ordered)
		assert.Len(t, group.Children, 2)
	})

	t.Run("should convert unordered group ops to GroupNode", func(t *testing.T) {
		ops := []marble.Op{
			marble.UnorderedGroupStartOp{EndPos: 3},
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
			marble.UnorderedGroupEndOp{StartPos: 0},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 1)
		group := node.Children[0].(*marble.GroupNode)
		assert.False(t, group.Ordered)
		assert.Len(t, group.Children, 2)
	})

	t.Run("should handle multiple groups", func(t *testing.T) {
		// Test with multiple separate groups
		ops := []marble.Op{
			marble.OrderedGroupStartOp{EndPos: 3},
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
			marble.OrderedGroupEndOp{StartPos: 0},
			marble.UnorderedGroupStartOp{EndPos: 7},
			marble.EventOp{Name: "c"},
			marble.EventOp{Name: "d"},
			marble.UnorderedGroupEndOp{StartPos: 4},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 2)
		group1 := node.Children[0].(*marble.GroupNode)
		assert.True(t, group1.Ordered)
		assert.Len(t, group1.Children, 2)
		group2 := node.Children[1].(*marble.GroupNode)
		assert.False(t, group2.Ordered)
		assert.Len(t, group2.Children, 2)
	})

	t.Run("should handle start event", func(t *testing.T) {
		ops := []marble.Op{
			marble.StartEventOp{},
			marble.EventOp{Name: "a"},
			marble.EventOp{Name: "b"},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 3)
		assert.IsType(t, &marble.StartNode{}, node.Children[0])
	})

	t.Run("should handle followup event", func(t *testing.T) {
		ops := []marble.Op{
			marble.EventWithFollowupOp{NewEvent: "b", OfEvent: "a"},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 1)
		followup := node.Children[0].(*marble.FollowupNode)
		assert.Equal(t, "b", followup.NewEvent)
		assert.Equal(t, "a", followup.OfEvent)
	})

	t.Run("should handle mixed sequence", func(t *testing.T) {
		ops := []marble.Op{
			marble.StartEventOp{},
			marble.EventOp{Name: "a"},
			marble.WaitOp{},
			marble.OrderedGroupStartOp{EndPos: 6},
			marble.EventOp{Name: "b"},
			marble.EventOp{Name: "c"},
			marble.OrderedGroupEndOp{StartPos: 3},
			marble.EventWithFollowupOp{NewEvent: "d", OfEvent: "c"},
		}
		node := marble.SequenceNodeFromOps(ops)
		assert.Len(t, node.Children, 5)
		assert.IsType(t, &marble.StartNode{}, node.Children[0])
		assert.IsType(t, &marble.EventNode{}, node.Children[1])
		assert.IsType(t, &marble.WaitNode{}, node.Children[2])
		assert.IsType(t, &marble.GroupNode{}, node.Children[3])
		assert.IsType(t, &marble.FollowupNode{}, node.Children[4])
	})
}
