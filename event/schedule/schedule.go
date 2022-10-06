package schedule

import (
	"time"
)

type eventschedule struct {
	interval time.Duration
	quit     chan bool
}

func NewScheduleEvent(interval time.Duration) Interface {
	return &eventschedule{interval: interval}
}
