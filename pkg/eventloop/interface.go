package eventloop

import (
	"context"

	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

type Interface interface {
	RegisterEvent(ctx context.Context, newEvent event.Interface) error
	Trigger(ctx context.Context, triggerName string) error
	Toggle(eventFunc ...EventFunction) string
	// RemoveEventByUUIDs удаляет событие по срезу идентификаторов. Возвращает срез оставшихся событий из запроса, которые
	// не были удалены
	RemoveEventByUUIDs(id []uuid.UUID) []uuid.UUID
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error
	GetAttachedEvents(eventName string) (result []uuid.UUID, err error)
	Sync() error
}
