package runtime

import (
	"fmt"
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

	publishedEvents   []event.Event // TODO: to remove
	publishedTimeline []Tick        // TODO: to remove
	publishedTicks    []Tick        // TODO: to remove
	eventsOverTime    map[time.Duration][]event.Event

	expectedOps []marble.Op
}

var (
	_ event.Bus = (*Interceptor)(nil)
)

// features:
// - intercepts all published events before publishing them back to the actual bus
// - able to make assertions from a marble sequence

func NewInterceptor(t *testing.T, bus event.Bus, clock Clock) *Interceptor {
	it := &Interceptor{
		actualBus:      bus,
		t:              t,
		clock:          clock,
		eventsOverTime: make(map[time.Duration][]event.Event),
	}

	t.Cleanup(func() {
		it.finish(true)
	})
	return it
}

func (i *Interceptor) Publish(evt event.Event) {
	i.actualBus.Publish(evt)
	elapsed := i.clock.Elapsed()
	if _, ok := i.eventsOverTime[elapsed]; !ok {
		i.eventsOverTime[elapsed] = make([]event.Event, 0)
	}
	i.eventsOverTime[elapsed] = append(i.eventsOverTime[elapsed], evt)
}

func (i *Interceptor) Subscribe() *event.Subscriber {
	return i.actualBus.Subscribe()
}

func (i *Interceptor) EXPECT() *InterceptorRecorder {
	r := &InterceptorRecorder{it: i}
	i.recorders = append(i.recorders, r)
	return r
}

func (i *Interceptor) finish(cleanup bool) {
	i.t.Helper()

	if !cleanup {
		i.t.Fail()
		//i.t.Fatalf("expected %d events, got %d", len(i.expectedOps), len(i.publishedEvents))
		//return
	}
	i.t.Errorf("expected %d events, got %d", len(i.expectedOps), len(i.publishedEvents))

	//if len(i.expectedOps) == 0 {
	//	// TODO:
	//	return
	//}

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
					r.matchers[o.Name] = event.HasPayload(DefaultPayload(o.Name))
				}

			case marble.EventWithFollowupOp:
				if _, ok := r.matchers[o.EventName]; !ok {
					r.matchers[o.EventName] = event.HasPayload(DefaultPayload(o.EventName))
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
		startTime       = r.it.clock.StartTime()
		eventInCurrTick []event.Event
	)

	nextTickStart := currentTick.Duration

	if r.it.clock.Started() {
		return []error{fmt.Errorf("clock has not been stopped")}
	}
	if startTime.IsZero() {
		return []error{fmt.Errorf("clock has never been started")}
	}

	for pubElapsed, publishedEvt := range r.it.eventsOverTime {
		if currentTickIds >= len(expectedTicks) {
			errs = append(errs, fmt.Errorf("unexpected event at time %s: %v", pubElapsed, publishedEvt))
			continue
		}
		if startTime.Add(pubElapsed).Before(startTime.Add(nextTickStart)) {
			eventInCurrTick = append(eventInCurrTick, publishedEvt...)

		} else {
			if len(eventInCurrTick) != len(currentTick.Ops) {
				errs = append(errs, fmt.Errorf("at tick %d: expected %d events, got %d", currentTickIds, len(currentTick.Ops), len(eventInCurrTick)))
				continue
			}

			for range len(eventInCurrTick) {

			}

			currentTickIds++
			currentTick = expectedTicks[currentTickIds]
			nextTickStart += currentTick.Duration
		}

		// marble:      a   bc  -   d
		// elapsed:     ---|---|---|---|
		// evtOverTime: -a- bc- --- -d-
	}

	return errs
}
