package interval

import (
	"time"
)

// component - событие, запускаемое с определённым интервалом. Имеет собственный канал, с помощью которого можно
// прервать работу события.
type component struct {
	interval  time.Duration
	isRunning bool
	quit      chan bool
}

func NewIntervalEvent(interval time.Duration) Interface {
	return &component{interval: interval, quit: make(chan bool)}
}

func (e *component) GetDuration() time.Duration {
	return e.interval
}

func (e *component) GetQuitChannel() chan bool {
	return e.quit
}

func (e *component) IsRunning() bool {
	return e.isRunning
}

func (e *component) SetRunning(run bool) {
	e.isRunning = run
}
