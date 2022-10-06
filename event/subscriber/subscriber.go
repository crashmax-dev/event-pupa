package subscriber

import (
	"sync"
)

type eventsubscriber struct {
	channels []chan int
	mx       sync.Mutex
}

func NewSubscriber() Interface {
	return &eventsubscriber{}
}

func (ev *eventsubscriber) LockMutex() {
	ev.mx.Lock()
}

func (ev *eventsubscriber) UnlockMutex() {
	ev.mx.Unlock()
}

func (ev *eventsubscriber) GetChannels() []chan int {
	return ev.channels
}

func (ev *eventsubscriber) AddChannel(ch chan int) {
	ev.channels = append(ev.channels, ch)
}
