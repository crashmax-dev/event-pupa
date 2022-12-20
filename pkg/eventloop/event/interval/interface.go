package interval

import "time"

type Interface interface {
	GetDuration() time.Duration
	GetQuitChannel() chan bool
	IsRunning() bool
	SetRunning(run bool)
}
