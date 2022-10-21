package eventloop

import (
	"context"
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

type Interface interface {
	On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- uuid.UUID)
	Trigger(ctx context.Context, eventName string, ch *Channel[string])
	Toggle(eventFunc ...EventFunction)
	ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID)
	StartScheduler(ctx context.Context)
	StopScheduler()
	RemoveEvent(id uuid.UUID) bool
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface)
	LockMutex()
	UnlockMutex()
}
