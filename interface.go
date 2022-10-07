package LearnProject

import (
	"context"
	"eventloop/event"
)

type Interface interface {
	On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- int)
	Trigger(ctx context.Context, eventName string)
	Toggle(eventFunc ...EventFunction)
	ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- int)
	StartScheduler(ctx context.Context)
	StopScheduler()
	RemoveEvent(id int) bool
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface)
	LockMutex()
	UnlockMutex()
}
