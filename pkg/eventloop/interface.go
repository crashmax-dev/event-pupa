package eventloop

import (
	"context"

	"eventloop/pkg/eventloop/event"
)

type Interface interface {
	RegisterEvent(ctx context.Context, newEvent ...event.Interface) error
	Trigger(ctx context.Context, triggerName string) error
	ToggleEventLoopFuncs(eventFunc ...EventFunction) string
	ToggleTriggers(triggerNames ...string) string
	// RemoveEventByUUIDs удаляет событие по срезу идентификаторов. Возвращает срез оставшихся событий из запроса, которые
	// не были удалены
	RemoveEventByUUIDs(UUIDs ...string) []string
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error
	GetAttachedEvents(triggerName string) (result []string, err error)
	Sync() error
}
