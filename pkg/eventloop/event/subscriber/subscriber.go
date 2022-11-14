package subscriber

import (
	"sync"
)

// eventSubscriber - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type eventSubscriber struct {
	channels []chan int
	mx       sync.Mutex
}

func NewSubscriber() Interface {
	return &eventSubscriber{}
}

func (ev *eventSubscriber) LockMutex() {
	ev.mx.Lock()
}

func (ev *eventSubscriber) UnlockMutex() {
	ev.mx.Unlock()
}

func (ev *eventSubscriber) GetChannels() []chan int {
	return ev.channels
}

// AddChannel добавляет каналы, по которым тригеррится событие, и по этим же каналам событие-триггер триггерит
// события слушатели.
func (ev *eventSubscriber) AddChannel(ch chan int) {
	ev.channels = append(ev.channels, ch)
}
