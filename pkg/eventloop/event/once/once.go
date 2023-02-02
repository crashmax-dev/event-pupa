package once

import "sync"

// eventSubscriber - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type component struct {
	sync.Once
}

func NewOnce() Interface {
	return &component{}
}

func (ev *component) Do(f func()) {
	ev.Once.Do(f)
}
