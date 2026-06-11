package runtime

import (
	"fmt"
	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
	"time"
)

type InterceptorValidator struct {
	timeline    *Timeline
	activity    []activityEntry
	matchers    map[string]event.Matcher
	errors      []error
	currentTick int
}

func NewInterceptorValidator(tl *Timeline, activity []activityEntry, matchers map[string]event.Matcher) *InterceptorValidator {
	return &InterceptorValidator{
		timeline: tl,
		activity: activity,
		matchers: matchers,
	}
}

func (v *InterceptorValidator) Validate(root marble.Node) []error {
	root.Accept(v)
	return v.errors
}

func (v *InterceptorValidator) VisitSequence(n *marble.SequenceNode) {
	for i, child := range n.Children {
		v.currentTick = i
		child.Accept(v)
	}
}

func (v *InterceptorValidator) VisitEvent(n *marble.EventNode) {
	v.validateSingleEvent(n.Name, v.currentTick)
}

func (v *InterceptorValidator) VisitWait(n *marble.WaitNode) {
	v.validateEmptyTick(v.currentTick)
}

func (v *InterceptorValidator) VisitStart(n *marble.StartNode) {
	// Start event is usually not verified in the same way, but we can check if it was published
	// For now, we skip it as it's often used for initialization
}

func (v *InterceptorValidator) VisitFollowup(n *marble.FollowupNode) {
	v.validateSingleEvent(n.NewEvent, v.currentTick)
}

func (v *InterceptorValidator) VisitGroup(n *marble.GroupNode) {
	tick := v.timeline.Ticks()[v.currentTick]

	// Calculate tick range
	tickStart := time.Duration(0)
	for i := 0; i < v.currentTick; i++ {
		tickStart += v.timeline.Ticks()[i].Duration
	}
	tickEnd := tickStart + tick.Duration

	tickActivity := selectActivityEntriesByRange(v.activity, tickStart, tickEnd)

	if n.Ordered {
		v.validateOrderedGroup(n, tickActivity)
	} else {
		v.validateUnorderedGroup(n, tickActivity)
	}
}

func (v *InterceptorValidator) validateSingleEvent(name string, tickIdx int) {
	if tickIdx >= len(v.timeline.Ticks()) {
		v.errors = append(v.errors, fmt.Errorf("extra event %s found (no corresponding tick %d)", name, tickIdx))
		return
	}
	tick := v.timeline.Ticks()[tickIdx]
	tickStart := time.Duration(0)
	for i := 0; i < tickIdx; i++ {
		tickStart += v.timeline.Ticks()[i].Duration
	}
	tickEnd := tickStart + tick.Duration

	tickActivity := selectActivityEntriesByRange(v.activity, tickStart, tickEnd)

	if len(tickActivity) != 1 {
		v.errors = append(v.errors, fmt.Errorf("tick %d: expected exactly one event (%s), got %d", tickIdx, name, len(tickActivity)))
		return
	}

	if m, ok := v.matchers[name]; ok {
		if !m.Match(tickActivity[0].event) {
			v.errors = append(v.errors, fmt.Errorf("tick %d: event %s does not match expected pattern", tickIdx, name))
		}
	}
}

func (v *InterceptorValidator) validateEmptyTick(tickIdx int) {
	if tickIdx >= len(v.timeline.Ticks()) {
		return
	}
	tickStart := time.Duration(0)
	for i := 0; i < tickIdx; i++ {
		tickStart += v.timeline.Ticks()[i].Duration
	}
	tickEnd := tickStart + v.timeline.Ticks()[tickIdx].Duration

	tickActivity := selectActivityEntriesByRange(v.activity, tickStart, tickEnd)
	if len(tickActivity) > 0 {
		v.errors = append(v.errors, fmt.Errorf("tick %d: nothing is supposed to happen, but %d events were published", tickIdx, len(tickActivity)))
	}
}

func (v *InterceptorValidator) validateOrderedGroup(n *marble.GroupNode, activity []activityEntry) {
	v.validateNode(n, activity)
}

func (v *InterceptorValidator) validateUnorderedGroup(n *marble.GroupNode, activity []activityEntry) {
	v.validateNode(n, activity)
}

func (v *InterceptorValidator) validateNode(n marble.Node, activity []activityEntry) {
	switch node := n.(type) {
	case *marble.EventNode:
		if len(activity) != 1 {
			v.errors = append(v.errors, fmt.Errorf("tick %d: expected 1 event for %s, got %d", v.currentTick, node.Name, len(activity)))
			return
		}
		if m, ok := v.matchers[node.Name]; ok {
			if !m.Match(activity[0].event) {
				v.errors = append(v.errors, fmt.Errorf("tick %d: event %s mismatch", v.currentTick, node.Name))
			}
		}

	case *marble.FollowupNode:
		if len(activity) != 1 {
			v.errors = append(v.errors, fmt.Errorf("tick %d: expected 1 event for followup %s, got %d", v.currentTick, node.NewEvent, len(activity)))
			return
		}
		if m, ok := v.matchers[node.NewEvent]; ok {
			if !m.Match(activity[0].event) {
				v.errors = append(v.errors, fmt.Errorf("tick %d: followup event %s mismatch", v.currentTick, node.NewEvent))
			}
		}

	case *marble.GroupNode:
		if node.Ordered {
			v.validateOrderedGroupInternal(node, activity)
		} else {
			v.validateUnorderedGroupInternal(node, activity)
		}
	}
}

func (v *InterceptorValidator) validateOrderedGroupInternal(n *marble.GroupNode, activity []activityEntry) {
	expectedCount := v.countEvents(n)
	if expectedCount != len(activity) {
		v.errors = append(v.errors, fmt.Errorf("tick %d (ordered group): expected %d events, got %d", v.currentTick, expectedCount, len(activity)))
		return
	}

	currentActivityIdx := 0
	for _, child := range n.Children {
		childEventsCount := v.countEvents(child)
		if currentActivityIdx+childEventsCount > len(activity) {
			// This shouldn't happen because of the check above, but for safety
			v.errors = append(v.errors, fmt.Errorf("tick %d: out of activity range", v.currentTick))
			return
		}

		v.validateNode(child, activity[currentActivityIdx:currentActivityIdx+childEventsCount])
		currentActivityIdx += childEventsCount
	}
}

func (v *InterceptorValidator) validateUnorderedGroupInternal(n *marble.GroupNode, activity []activityEntry) {
	expectedCount := v.countEvents(n)
	if expectedCount != len(activity) {
		v.errors = append(v.errors, fmt.Errorf("tick %d (unordered group): expected %d events, got %d", v.currentTick, expectedCount, len(activity)))
		return
	}

	// For unordered groups, it's more complex because children can be subgroups.
	// We'll simplify: collect all leaf events and match them.
	// This might lose some structure if we have (a[bc]), but the current marble language
	// doesn't really have clear semantics for unordered groups containing ordered ones
	// in terms of "partial order".

	var expected []string
	v.collectExpectedNames(n, &expected)

	matched := make(map[int]bool)
	for _, act := range activity {
		found := false
		for i, name := range expected {
			if matched[i] {
				continue
			}
			if m, ok := v.matchers[name]; ok {
				if m.Match(act.event) {
					matched[i] = true
					found = true
					break
				}
			}
		}
		if !found {
			v.errors = append(v.errors, fmt.Errorf("tick %d (unordered group): unexpected event %v", v.currentTick, act.event))
		}
	}
}

func (v *InterceptorValidator) countEvents(n marble.Node) int {
	switch node := n.(type) {
	case *marble.EventNode, *marble.FollowupNode:
		return 1
	case *marble.GroupNode:
		count := 0
		for _, child := range node.Children {
			count += v.countEvents(child)
		}
		return count
	case *marble.SequenceNode:
		count := 0
		for _, child := range node.Children {
			count += v.countEvents(child)
		}
		return count
	default:
		return 0
	}
}

func (v *InterceptorValidator) collectExpectedNames(n marble.Node, names *[]string) {
	switch node := n.(type) {
	case *marble.EventNode:
		*names = append(*names, node.Name)
	case *marble.FollowupNode:
		*names = append(*names, node.NewEvent)
	case *marble.GroupNode:
		for _, child := range node.Children {
			v.collectExpectedNames(child, names)
		}
	case *marble.SequenceNode:
		for _, child := range node.Children {
			v.collectExpectedNames(child, names)
		}
	}
}
