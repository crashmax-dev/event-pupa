package subscriber

import "EventManager/event"

type Interface interface {
	LockMutex()
	UnlockMutex()
	GetChannels() []chan int
	AddChannel(chan int)
	GetBase() event.Interface
}
