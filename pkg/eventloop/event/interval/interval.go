package interval

import (
	"time"
)

// eventSchedule - событие, запускаемое с определённым интервалом. Имеет собственный канал, с помощью которого можно
// прервать работу события.
type eventSchedule struct {
	interval  time.Duration
	isRunning bool
	quit      chan bool
}

func NewIntervalEvent(interval time.Duration) Interface {
	return &eventSchedule{interval: interval, quit: make(chan bool)}
}

func (e *eventSchedule) GetDuration() time.Duration {
	return e.interval
}

func (e *eventSchedule) GetQuitChannel() chan bool {
	return e.quit
}

func (e *eventSchedule) IsRunning() bool {
	return e.isRunning
}

func (e *eventSchedule) SetRunning(run bool) {
	e.isRunning = run
}
