package runtime

import (
	"slices"
	"time"
)

type Clock interface {
	Now() time.Time
	Start()
	Started() bool
	StartTime() time.Time
	Stop()
	Elapsed() time.Duration
	Forward(duration time.Duration)
	Schedule(delay time.Duration, f func())
}

type clockImpl struct {
	current   time.Time
	started   bool
	startTime time.Time
	scheduled map[time.Duration]func()
}

var (
	_ Clock = (*clockImpl)(nil)
)

func NewClock() Clock {
	return &clockImpl{
		scheduled: make(map[time.Duration]func()),
	}
}

func (c *clockImpl) Now() time.Time {
	return c.current
}

func (c *clockImpl) Start() {
	c.started = true
	c.startTime = time.Now()
	c.current = c.startTime
	for delay, f := range c.scheduled {
		if delay <= 0 {
			f()
			delete(c.scheduled, delay)
		}
	}
}

func (c *clockImpl) Started() bool {
	return c.started
}

func (c *clockImpl) StartTime() time.Time {
	return c.startTime
}

func (c *clockImpl) Stop() {
	c.started = false
	c.startTime = time.Time{}
}

func (c *clockImpl) Elapsed() time.Duration {
	if !c.started {
		return 0
	}
	return c.Now().Sub(c.startTime)
}

type scheduledFunc struct {
	f         func()
	startTime time.Time
}

func (c *clockImpl) Forward(duration time.Duration) {
	c.current = c.current.Add(duration)
	elapsed := c.Elapsed()
	toBeRun := make([]scheduledFunc, 0)
	for delay, f := range c.scheduled {
		if delay <= elapsed {
			toBeRun = append(toBeRun, scheduledFunc{
				f:         f,
				startTime: c.startTime.Add(delay),
			})
			delete(c.scheduled, delay)
		}
	}

	slices.SortFunc(toBeRun, func(a, b scheduledFunc) int {
		return a.startTime.Compare(b.startTime)
	})
	for _, f := range toBeRun {
		f.f()
	}
}

func (c *clockImpl) Schedule(delay time.Duration, f func()) {
	if c.started {
		panic("cannot schedule event when clock is started")
	}
	if c.scheduled == nil {
		c.scheduled = make(map[time.Duration]func())
	}
	c.scheduled[delay] = f
}
