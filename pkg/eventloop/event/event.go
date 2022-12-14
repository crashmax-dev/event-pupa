package event

import (
	"context"
	"errors"
	"sync"
	"time"

	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
)

// event - обычное событие, которое может иметь свойства других событий (одноразовых, интервальных, зависимых)
type event struct {
	id       uuid.UUID
	priority int
	fun      EventFunc

	mx sync.Mutex

	subscriber subscriber.Interface
	interval   interval.Interface
	once       once.Interface
}

type EventFunc func(ctx context.Context) string

func NewEvent(fun EventFunc) Interface {
	return &event{id: uuid.New(), fun: fun}
}

func NewIntervalEvent(fun EventFunc, intervalTime time.Duration) Interface {
	return &event{id: uuid.New(), fun: fun, interval: interval.NewIntervalEvent(intervalTime)}
}

func NewOnceEvent(fun EventFunc) Interface {
	return &event{id: uuid.New(), fun: fun, once: once.NewOnce()}
}

func NewPriorityEvent(fun EventFunc, priority int) Interface {
	return &event{id: uuid.New(), fun: fun, priority: priority}
}

func (ev *event) GetID() uuid.UUID {
	return ev.id
}

func (ev *event) GetPriority() int {
	return ev.priority
}

func (ev *event) SetPriority(prior int) {
	ev.priority = prior
}

func (ev *event) RunFunction(ctx context.Context) string {
	return ev.fun(ctx)
}

// Subscriber
func (ev *event) Subscriber() subscriber.Interface {
	if ev.subscriber == nil {
		ev.subscriber = subscriber.NewSubscriber()
	}
	return ev.subscriber
}

func (ev *event) Interval() (interval.Interface, error) {
	if ev.interval == nil {
		return nil, errors.New("it is not an interval event")
	}
	return ev.interval, nil
}

func (ev *event) Once() (once.Interface, error) {
	if ev.once == nil {
		return nil, errors.New("it is not an once event")
	}
	return ev.once, nil
}
