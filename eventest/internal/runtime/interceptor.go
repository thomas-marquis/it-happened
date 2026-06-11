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
	expectedSeq  string
	expectedNode marble.Node
	timeline     *Timeline
	it           *Interceptor
	matchers     map[string]event.Matcher
}

func (r *InterceptorRecorder) FromMarble(seq string) *InterceptorRecorder {
	if r.expectedSeq != "" {
		panic("already expecting a marble sequence")
	}
	r.expectedSeq = seq

	node, err := marble.ParseAsNode(seq)
	if err != nil {
		panic(err)
	}
	r.expectedNode = node

	if err := marble.Validate(node,
		marble.WaitlessGroupsRule{}); err != nil {
		panic(err)
	}

	tl := NewTimeline(node)
	r.timeline = tl

	for _, tick := range tl.Ticks() {
		for _, op := range tick.Ops {
			switch o := op.(type) {
			case marble.EventOp:
				if _, ok := r.matchers[o.Name]; !ok {
					r.matchers[o.Name] = event.HasPayload(DefaultPayload(o.Name))
				}

			case marble.EventWithFollowupOp:
				if _, ok := r.matchers[o.NewEvent]; !ok {
					r.matchers[o.NewEvent] = event.HasPayload(DefaultPayload(o.NewEvent))
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
	if r.expectedNode == nil {
		panic("no timeline defined: please specify a marble sequence using FromMarble()")
	}

	if r.it.clock.Started() {
		return []error{fmt.Errorf("clock has not been stopped")}
	}

	validator := NewInterceptorValidator(r.timeline, r.it.actualActivityEntries, r.matchers)
	return validator.Validate(r.expectedNode)
}

type activityEntry struct {
	elapsedFromStart time.Duration
	event            event.Event
}

func sortActivityEntries(entries []activityEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].elapsedFromStart < entries[j].elapsedFromStart
	})
}

func selectActivityEntriesByRange(entries []activityEntry, start, end time.Duration) []activityEntry {
	var res []activityEntry
	for _, e := range entries {
		if e.elapsedFromStart >= start && e.elapsedFromStart < end {
			res = append(res, e)
		}
	}
	return res
}
