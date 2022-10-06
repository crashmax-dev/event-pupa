package eventloop

import (
	"EventManager/event"
	"EventManager/event/schedule"
	"EventManager/event/subscriber"
	"context"
)

type EventLoop interface {
	On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- int)
	Trigger(ctx context.Context, eventName string)
	Toggle(eventFunc ...EventFunction)
	ScheduleEvent(ctx context.Context, newEvent schedule.Interface, out chan<- int)
	StartScheduler(ctx context.Context)
	StopScheduler()
	RemoveEvent(id int) bool
	Subscribe(ctx context.Context, triggers []subscriber.Interface, listeners []schedule.Interface)
}
