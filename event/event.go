package event

import (
	"EventManager/event/schedule"
	"EventManager/event/subscriber"
	"EventManager/helpers"
	"context"
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
	return ev.subscriber
}

func (ev *event) GetSchedule() schedule.Interface {
	return ev.schedule
}
