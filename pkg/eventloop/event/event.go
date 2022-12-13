package event

import (
	"context"
	"errors"
	"eventloop/pkg/eventloop/event/schedule"
	"eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
	"sync"
	"time"
)

// event - обычное событие, которое может иметь свойства других событий (одноразовых, интервальных, зависимых)
type event struct {
	id       uuid.UUID
	priority int
	fun      EventFunc

	isOnce bool
	sync.Once

	mx sync.Mutex

	subscriber subscriber.Interface
	schedule   schedule.Interface
}

type EventFunc func(ctx context.Context) string

func NewEvent(fun EventFunc) Interface {
	return &event{id: uuid.New(), fun: fun}
}

func NewIntervalEvent(fun EventFunc, interval time.Duration) Interface {
	return &event{id: uuid.New(), fun: fun, schedule: schedule.NewScheduleEvent(interval)}
}

func NewOnceEvent(fun EventFunc) Interface {
	return &event{id: uuid.New(), fun: fun, isOnce: true}
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

func (ev *event) GetSubscriber() subscriber.Interface {
	if ev.subscriber == nil {
		return nil, errors.New("it is not a subscriber event")
	}
	return ev.subscriber, nil
}

func (ev *event) GetSchedule() (schedule.Interface, error) {
	if ev.schedule == nil {
		return nil, errors.New("it is not an interval event")
	}
	return ev.schedule, nil
}

func (ev *event) IsOnce() bool {
	return ev.isOnce
}

func (ev *event) GetOnce() *sync.Once {
	return &ev.Once
}
