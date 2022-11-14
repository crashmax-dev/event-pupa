package schedule

import (
	"time"
)

// eventSchedule - событие, запускаемое с определённым интервалом. Имеет собственный канал, с помощью которого можно
// прервать работу события.
type eventSchedule struct {
	interval time.Duration
	quit     chan bool
}

func NewScheduleEvent(interval time.Duration) Interface {
	return &eventSchedule{interval: interval, quit: make(chan bool)}
}

func (e eventSchedule) GetInterval() time.Duration {
	return e.interval
}

func (e eventSchedule) GetQuitChannel() chan bool {
	return e.quit
}
