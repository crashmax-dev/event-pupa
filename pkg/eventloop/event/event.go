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

type event struct {
	id       uuid.UUID
	priority int
	fun      func(ctx context.Context) string

	isOnce bool
	sync.Once

	mx sync.Mutex

	subscriber subscriber.Interface
	schedule   schedule.Interface
}

func NewEvent(fun func(ctx context.Context) string) Interface {
	return &event{id: uuid.New(), fun: fun}
}

func NewIntervalEvent(fun func(ctx context.Context) string, interval time.Duration) Interface {
	return &event{id: uuid.New(), fun: fun, schedule: schedule.NewScheduleEvent(interval)}
}

func NewOnceEvent(fun func(ctx context.Context) string) Interface {
	return &event{id: uuid.New(), fun: fun, isOnce: true}
}

func NewPriorityEvent(fun func(ctx context.Context) string, priority int) Interface {
	return &event{id: uuid.New(), fun: fun, priority: priority}
}

func (ev *event) GetId() uuid.UUID {
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
		ev.subscriber = subscriber.NewSubscriber()
	}
	return ev.subscriber
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
