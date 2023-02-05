package eventloop

import (
	"context"

	"gitlab.com/YSX/eventloop/pkg/eventloop/event"
)

type Interface interface {
	RegisterEvent(ctx context.Context, newEvent ...event.Interface) error
	RegisterMiddlewareEvent(ev event.Interface, triggerName string) error
	Trigger(ctx context.Context, triggerName string) error
	ToggleEventLoopFuncs(eventFunc ...EventFunction) string
	ToggleTriggers(triggerNames ...string) string
	// RemoveEventByUUIDs удаляет событие по срезу идентификаторов. Возвращает срез оставшихся событий из запроса, которые
	// не были удалены
	RemoveEventByUUIDs(UUIDs ...string) []string
	RemoveTriggers(triggers ...string) []string
	Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error
	GetAttachedEvents(triggerName string) (result map[string][]event.Interface)
	GetTriggerNames() AllTriggers
}
