package schedule

import "time"

type Interface interface {
	GetInterval() time.Duration
	GetQuitChannel() chan bool
}
