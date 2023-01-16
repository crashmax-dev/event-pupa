package interval

import (
	"time"
)

// eventInterval - событие, запускаемое с определённым интервалом. Имеет собственный канал, с помощью которого можно
// прервать работу события.
type eventInterval struct {
	interval  time.Duration
	isRunning bool
	quit      chan bool
}

func NewIntervalEvent(interval time.Duration) Interface {
	return &eventInterval{interval: interval, quit: make(chan bool)}
}

func (e *eventInterval) GetDuration() time.Duration {
	return e.interval
}

func (e *eventInterval) GetQuitChannel() chan bool {
	return e.quit
}

func (e *eventInterval) IsRunning() bool {
	return e.isRunning
}

func (e *eventInterval) SetRunning(run bool) {
	e.isRunning = run
}
