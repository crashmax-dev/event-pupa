package eventloop

import (
	"context"
	"eventloop/pkg/channelEx"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal/scheduler"
	"github.com/google/uuid"
)

type Interface interface {
	addEvent(eventName string, newEvent event.Interface)
	On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- uuid.UUID) error
	OnBefore(ctx context.Context, eventName string, elFunction event.EventFunc)
	OnAfter(ctx context.Context, eventName string, elFunction event.EventFunc)
	Trigger(ctx context.Context, eventName string, ch channelEx.Interface[string]) error
	Toggle(eventFunc ...EventFunction) string
	ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID) error
	StartScheduler(ctx context.Context) error
	Scheduler() scheduler.Interface
	StopScheduler()
	RemoveEventByUUIDs(id []uuid.UUID) []uuid.UUID
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error
	GetAttachedEvents(eventName string) (result []uuid.UUID, err error)
}
