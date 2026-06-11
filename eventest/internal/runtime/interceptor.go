package runtime

import (
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/thomas-marquis/it-happened/event"
	"github.com/thomas-marquis/it-happened/eventest/internal/marble"
)

type Interceptor struct {
	actualBus event.Bus
	clock     Clock
	t         *testing.T
	recorders []*InterceptorRecorder

	actualActivityEntries []activityEntry

	expectedOps []marble.Op
}

var (
	_ event.Bus = (*Interceptor)(nil)
)

func NewInterceptor(t *testing.T, bus event.Bus, clock Clock) *Interceptor {
	it := &Interceptor{
		actualBus:             bus,
		t:                     t,
		clock:                 clock,
		actualActivityEntries: make([]activityEntry, 0),
	}

	t.Cleanup(func() {
		it.finish(true)
	})
	return it
}

func (i *Interceptor) Publish(evt event.Event) {
	i.actualBus.Publish(evt)
	i.actualActivityEntries = append(i.actualActivityEntries,
		activityEntry{elapsedFromStart: i.clock.Elapsed(), event: evt})
}

func (i *Interceptor) Subscribe() *event.Subscriber {
	return i.actualBus.Subscribe()
}

func (i *Interceptor) EXPECT() *InterceptorRecorder {
	r := &InterceptorRecorder{
		it:       i,
		matchers: make(map[string]event.Matcher),
	}
	i.recorders = append(i.recorders, r)
	return r
}

func (i *Interceptor) finish(cleanup bool) {
	i.t.Helper()

	sortActivityEntries(i.actualActivityEntries)

	var errs []error
	for _, rec := range i.recorders {
		errs = append(errs, rec.Failures()...)
	}

	if len(errs) == 0 {
		return
	}

	for _, err := range errs {
		i.t.Error(err)
	}

	if !cleanup {
		i.t.Fail()
	}
}

// Finish forces to terminate the test and perform assertions.
// Within a classic unit test, this method is unlikely useful: the finalization is automatically performed by the test framework.
func (i *Interceptor) Finish() {
	i.t.Helper()
	i.finish(false)
}

type InterceptorRecorder struct {
	expectedSeq string
	timeline    *Timeline
	it          *Interceptor
	matchers    map[string]event.Matcher
}

func (r *InterceptorRecorder) FromMarble(seq string) *InterceptorRecorder {
	if r.expectedSeq != "" {
		panic("already expecting a marble sequence")
	}
	r.expectedSeq = seq

	ops, err := marble.Parse(seq)
	if err != nil {
		panic(err)
	}

	if err := marble.Validate(ops,
		marble.WaitlessGroupsRule{}); err != nil {
		panic(err)
	}

	r.it.expectedOps = ops

	tl := NewTimeline(ops)
	r.timeline = tl

	for _, tick := range tl.Ticks() {
		for _, op := range tick.Ops {
			switch o := op.(type) {
			case marble.EventOp:
				if _, ok := r.matchers[o.Name]; !ok {
					r.matchers[o.Name] = event.HasPayload(DefaultPayload(o.Name)) // TODO: wrong way...
				}

			case marble.EventWithFollowupOp:
				if _, ok := r.matchers[o.EventName]; !ok {
					r.matchers[o.EventName] = event.HasPayload(DefaultPayload(o.EventName)) // TODO: wrong way...
				}
			}
		}
	}

	return r
}

func (r *InterceptorRecorder) ShouldMatch(matchers map[string]event.Matcher) {
	for label, matcher := range matchers {
		r.matchers[label] = matcher
	}
}

func (r *InterceptorRecorder) Failures() []error {
	expectedTicks := r.timeline.Ticks()
	if len(expectedTicks) == 0 {
		panic("no timeline defined: please specify a marble sequence using FromMarble()")
	}

	var (
		errs            []error
		currentTickIds  int
		currentTick     = expectedTicks[0]
		eventInCurrTick []event.Event
	)

	nextTickStart := currentTick.Duration

	if r.it.clock.Started() {
		return []error{fmt.Errorf("clock has not been stopped")}
	}

	var tickStart time.Duration
	for i, tick := range r.timeline.Ticks() {
		tickEnd := tickStart + tick.Duration

		tickActivity := selectActivityEtriesByRange(r.it.actualActivityEntries, nextTickStart, tickEnd)

		// mode strict
		if len(tick.Ops) == 0 && len(tickActivity) > 0 {
			errs = append(errs, fmt.Errorf("nothing is supposed to happen in the tick %d", i))
		} else if len(tick.Ops) == 1 {
			if len(tickActivity) != 1 {
				errs = append(errs, fmt.Errorf("expected exactly one event in the tick %d", i))
				continue
			}
			switch op := tick.Ops[0].(type) {
			case marble.EventOp:
				if m := r.matchers[op.Name]; !m.Match(tickActivity[0].event) {
					errs = append(errs, fmt.Errorf("expected event %s to match %v, got %v",
						op.Name, m, tickActivity[0].event))
				}
			case marble.EventWithFollowupOp:
				if m := r.matchers[op.EventName]; !m.Match(tickActivity[0].event) {
					errs = append(errs, fmt.Errorf("expected event %s to match %v, got %v",
						op.EventName, m, tickActivity[0].event))
				}
			default:
				panic("implementation error: unexpected op type for matching")
			}
		} else {
			var grpPos int
			errs = append(errs, r.failuresFromGroup(tick, tickActivity, tick.Ops, &grpPos)...)
		}

		// mode lenient
		if len(tick.Ops) == 0 {
			continue
		} else if len(tick.Ops) == 1 {
			for _, act := range tickActivity {
				switch op := tick.Ops[0].(type) {
				case marble.EventOp:
					if m := r.matchers[op.Name]; !m.Match(act.event) {
						errs = append(errs, fmt.Errorf("expected event %s to match %v, got %v",
							op.Name, m, act.event))
					}
				case marble.EventWithFollowupOp:
					if m := r.matchers[op.EventName]; !m.Match(act.event) {
						errs = append(errs, fmt.Errorf("expected event %s to match %v, got %v",
							op.EventName, m, act.event))
					}
				default:
					panic("implementation error: unexpected op type for matching")
				}
			}
		} else {

		}

		tickStart = tickEnd
	}

	return errs
}

func (r *InterceptorRecorder) failuresFromGroup(tick Tick, activity []activityEntry, ops []marble.Op, posOp, posEvt *int, strict bool) []error {
	if len(ops) <= 2 {
		return nil
	}

	var (
		grpOpStart  = *posOp
		grpOpEnd    int
		grpEvtStart = *posEvt
		ordered     bool
		errs        []error
	)

	switch o := ops[0].(type) {
	case marble.OrderedGroupStartOp:
		grpOpEnd = o.EndPos
		ordered = true
	case marble.UnorderedGroupStartOp:
		grpOpEnd = o.EndPos
		ordered = false
	}

	grpOps := ops[grpOpStart+1 : grpOpEnd]
	grpActEntries := activity[grpEvtStart : grpEvtStart+len(grpOps)]

	if strict && len(grpActEntries) != len(grpOps) {
		return append(errs, fmt.Errorf("expected %d events in group, got %d", len(grpOps), len(grpActEntries)))
	}

	if !strict && len(grpActEntries) < len(grpOps) {
		return append(errs, fmt.Errorf("expected at least %d events in group, got %d", len(grpOps), len(grpActEntries)))
	}

	if ordered {

		for *posOp < grpOpEnd {
			op := grpOps[*posOp]
			if op.Type() == marble.OrderedGroupStartType || op.Type() == marble.UnorderedGroupStartType {
				errs = append(errs, r.failuresFromGroup(tick, activity, ops, posOp, posEvt, strict)...)
				*posOp++
				continue
			}

			label := getOpLabel(op)

			if strict {
				ae := grpActEntries[*posEvt]
				evt := ae.event
				m := r.matchers[label]

				if !m.Match(evt) {
					errs = append(errs, fmt.Errorf("event %s (%s) does not match op %s", evt.ID, evt.Type(), label))
				}
			} else {
				// TODO
			}
		}
	} else {
		// TODO
	}

	return errs
}

func (r *InterceptorRecorder) failuresFromOrderedGroup(tick Tick, activity []activityEntry, ops []marble.Op, posOp, posEvt *int) []error {
	return nil
}

func (r *InterceptorRecorder) failuresFromUnorderedGroup(tick Tick, activity []activityEntry, ops []marble.Op, posOp, posEvt *int) []error {
	return nil
}

//func (r *InterceptorRecorder) matchesEventsWithOps(events []event.Event, ops []marble.Op, pos *int) bool {
//	for *pos < len(ops) {
//		currOp := ops[*pos]
//		switch op := currOp.(type) {
//		case marble.EventOp:
//			//if !event.HasPayload(DefaultPayload(op.Name)).Match(events[*pos]) {
//			//	return false
//			//}
//			//*pos++
//		case marble.EventWithFollowupOp:
//			//if !event.HasPayload(DefaultPayload(op.EventName)).Match(events[*pos]) {
//			//	return false
//			//}
//		}
//	}
//
//}

//func countOps(ops []marble.Op) int {
//	var cnt int
//	for _, op := range ops {
//		if op.Type() == marble.EventOpType || op.Type() == marble.EventWithFollowupOpType {
//			cnt++
//		}
//	}
//
//	return cnt
//}

type activityEntry struct {
	elapsedFromStart time.Duration
	event            event.Event
}

func sortActivityEntries(entries []activityEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].elapsedFromStart < entries[j].elapsedFromStart
	})
}

func selectActivityEtriesByRange(entries []activityEntry, start, end time.Duration) []activityEntry {
	var res []activityEntry
	for _, e := range entries {
		if e.elapsedFromStart >= start && e.elapsedFromStart < end {
			res = append(res, e)
		}
	}
	return res
}

func getOpLabel(op marble.Op) string {
	switch o := op.(type) {
	case marble.EventOp:
		return o.Name
	case marble.EventWithFollowupOp:
		return o.EventName
	default:
		panic("implementation error: unexpected op type")
	}
}

func recordOrderedGroup(ops []marble.Op, tickPos, grpPos *int) {

}

func extractGroupParts(ops []marble.Op) [][]marble.Op {
	var (
		parts [][]marble.Op
		i     int
	)

	for i < len(ops) {
		op := ops[i]
		if op.Type() == marble.O
	}
}

func isGroup(ops []marble.Op) (isGrp bool, grpEnd int) {
	n := len(ops)
	if n <= 2 {
		return
	}
	isGrp = (ops[0].Type() == marble.OrderedGroupStartType && ops[n-1].Type() == marble.UnorderedGroupEndType) ||
		(ops[0].Type() == marble.UnorderedGroupStartType && ops[n-1].Type() == marble.OrderedGroupEndType)
	return
}

//
//type interceptEngine struct {
//	currPos *int
//	ops []marble.Op
//}
//
//func (e *interceptEngine)