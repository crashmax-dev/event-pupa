package after

import "time"

type Interface interface {
	GetDuration() time.Duration
	GetBreakChannel() chan bool
	IsDone() bool
	Wait()
}
