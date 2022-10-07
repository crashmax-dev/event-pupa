package event

import (
	"context"
	"eventloop/event/schedule"
	"eventloop/event/subscriber"
	"eventloop/helpers"
	"time"
)

type event struct {
	//name string
	id       int
	priority int
	fun      func(ctx context.Context)
	isOnce   bool

	subscriber subscriber.Interface
	schedule   schedule.Interface
}

func NewEvent(fun func(ctx context.Context)) Interface {
	return &event{id: helpers.GenerateIdFromNow(), fun: fun}
}

func NewIntervalEvent(fun func(ctx context.Context), interval time.Duration) Interface {
	return &event{id: helpers.GenerateIdFromNow(), fun: fun, schedule: schedule.NewScheduleEvent(interval)}
}

func NewOnceEvent(fun func(ctx context.Context)) Interface {
	return &event{id: helpers.GenerateIdFromNow(), fun: fun, isOnce: true}
}

func NewPriorityEvent(fun func(ctx context.Context), priority int) Interface {
	return &event{id: helpers.GenerateIdFromNow(), fun: fun, priority: priority}
}

func (ev *event) GetId() int {
	return ev.id
}

func (ev *event) GetPriority() int {
	return ev.priority
}

func (ev *event) SetPriority(prior int) {
	ev.priority = prior
}

func (ev *event) RunFunction(ctx context.Context) {
	ev.fun(ctx)
}

func (ev *event) GetSubscriber() subscriber.Interface {
	if ev.subscriber == nil {
		ev.subscriber = subscriber.NewSubscriber()
	}
	return ev.subscriber
}

func (ev *event) GetSchedule() schedule.Interface {
	return ev.schedule
}

func (ev *event) IsOnce() bool {
	return ev.isOnce
}
