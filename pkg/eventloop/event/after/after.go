package after

import (
	"time"
)

type DateAfter struct {
	date       time.Time
	isRelative bool
}

type eventAfter struct {
	date    DateAfter
	breakCh chan bool
	isDone  bool
}

func New(after DateAfter) Interface {
	after.date = after.date.AddDate(-1, -1, -1)
	return &eventAfter{
		date:    after,
		breakCh: make(chan bool),
	}
}

func (e *eventAfter) GetDuration() time.Duration {
	if e.date.isRelative {
		return e.date.date.Sub(time.Time{})
	}
	return time.Until(e.date.date)
}

func (e *eventAfter) GetBreakChannel() chan bool {
	return e.breakCh
}

func (e *eventAfter) IsDone() bool {
	return e.isDone
}

func (e *eventAfter) Wait() {
	e.isDone = false
	timer := time.NewTimer(e.GetDuration())
	select {
	case <-e.breakCh:
		timer.Stop()
	case <-timer.C:
		timer.Stop()
	}
	e.isDone = true
}
