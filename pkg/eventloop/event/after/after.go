package after

import (
	"time"
)

type Args struct {
	Date       time.Time
	IsRelative bool
}

type component struct {
	date    Args
	breakCh chan bool
	isDone  bool
}

func New(after Args) Interface {
	after.Date = after.Date.AddDate(-1, -1, -1)
	return &component{
		date:    after,
		breakCh: make(chan bool),
	}
}

func (e *component) GetDuration() time.Duration {
	if e.date.IsRelative {
		return e.date.Date.Sub(time.Time{})
	}
	return time.Until(e.date.Date)
}

func (e *component) GetBreakChannel() chan bool {
	return e.breakCh
}

func (e *component) IsDone() bool {
	return e.isDone
}

func (e *component) Wait() {
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
