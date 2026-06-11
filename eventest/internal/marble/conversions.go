package marble

// SequenceNodeFromOps converts []Op to SequenceNode, reconstructing group hierarchy from position markers
func SequenceNodeFromOps(ops []Op) *SequenceNode {
	children := parseOps(ops, 0)
	return &SequenceNode{Children: children}
}

// parseOps parses a slice of Ops starting from index start, returning the converted Nodes.
// It stops when it encounters an end marker that matches the current group.
func parseOps(ops []Op, start int) []Node {
	var children []Node
	pos := start

	for pos < len(ops) {
		op := ops[pos]

		switch o := op.(type) {
		case OrderedGroupStartOp:
			// Find the matching end first
			endPos := -1
			for i := pos + 1; i < len(ops); i++ {
				if endOp, ok := ops[i].(OrderedGroupEndOp); ok && endOp.StartPos == pos {
					endPos = i
					break
				}
			}
			if endPos == -1 {
				// No matching end found, treat as regular op
				children = append(children, o.ToNode())
				pos++
				continue
			}
			// Parse group contents (between start+1 and endPos-1)
			groupChildren := parseOps(ops[pos+1:endPos], 0)
			children = append(children, &GroupNode{
				Children: groupChildren,
				Ordered:  true,
			})
			// Skip past the end marker
			pos = endPos + 1

		case UnorderedGroupStartOp:
			// Find the matching end first
			endPos := -1
			for i := pos + 1; i < len(ops); i++ {
				if endOp, ok := ops[i].(UnorderedGroupEndOp); ok && endOp.StartPos == pos {
					endPos = i
					break
				}
			}
			if endPos == -1 {
				// No matching end found, treat as regular op
				children = append(children, o.ToNode())
				pos++
				continue
			}
			// Parse group contents (between start+1 and endPos-1)
			groupChildren := parseOps(ops[pos+1:endPos], 0)
			children = append(children, &GroupNode{
				Children: groupChildren,
				Ordered:  false,
			})
			// Skip past the end marker
			pos = endPos + 1

		case OrderedGroupEndOp, UnorderedGroupEndOp:
			// End markers are handled by their corresponding start markers
			// This should not be reached at the top level
			return children

		default:
			// Regular op - convert directly
			children = append(children, opToNode(op))
			pos++
		}
	}

	return children
}

// opToNode converts a single Op to Node
func opToNode(op Op) Node {
	switch o := op.(type) {
	case EventOp:
		return o.ToNode()
	case WaitOp:
		return o.ToNode()
	case StartEventOp:
		return o.ToNode()
	case EventWithFollowupOp:
		return o.ToNode()
	}
	return nil
}
