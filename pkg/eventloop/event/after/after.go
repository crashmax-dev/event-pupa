package after

import (
	"time"
)

type DateAfterArgs struct {
	Date       time.Time
	IsRelative bool
}

type eventAfter struct {
	date    DateAfterArgs
	breakCh chan bool
	isDone  bool
}

func New(after DateAfterArgs) Interface {
	after.Date = after.Date.AddDate(-1, -1, -1)
	return &eventAfter{
		date:    after,
		breakCh: make(chan bool),
	}
}

func (e *eventAfter) GetDuration() time.Duration {
	if e.date.IsRelative {
		return e.date.Date.Sub(time.Time{})
	}
	return time.Until(e.date.Date)
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
