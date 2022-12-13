package once

import "sync"

// eventSubscriber - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type eventOnce struct {
	sync.Once
}

func NewOnce() Interface {
	return &eventOnce{}
}

func (ev *eventOnce) Do(f func()) {
	ev.Once.Do(f)
}
