package event

import (
	"context"
	"errors"
	schedule2 "eventloop/pkg/eventloop/event/schedule"
	subscriber2 "eventloop/pkg/eventloop/event/subscriber"
	"github.com/google/uuid"
	"sync"
	"time"
)

type event struct {
	//name string
	id       uuid.UUID
	priority int
	fun      func(ctx context.Context) string
	isOnce   bool
	mx       sync.Mutex

	subscriber subscriber2.Interface
	schedule   schedule2.Interface
}

func NewEvent(fun func(ctx context.Context) string) Interface {
	return &event{id: uuid.New(), fun: fun}
}

func NewIntervalEvent(fun func(ctx context.Context) string, interval time.Duration) Interface {
	return &event{id: uuid.New(), fun: fun, schedule: schedule2.NewScheduleEvent(interval)}
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

func (ev *event) GetSubscriber() subscriber2.Interface {
	if ev.subscriber == nil {
		ev.subscriber = subscriber2.NewSubscriber()
	}
	return ev.subscriber
}

func (ev *event) GetSchedule() (schedule2.Interface, error) {
	if ev.schedule == nil {
		return nil, errors.New("It is not an interval event")
	}
	return ev.schedule, nil
}

func (ev *event) IsOnce() bool {
	return ev.isOnce
}
