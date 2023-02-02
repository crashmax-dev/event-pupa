package eventsContainer

import (
	"eventloop/pkg/eventloop/event"
)

type Interface interface {
	AddEvent(newEvent event.Interface)
	EventsByTrigger(triggerName string) []event.Interface
	GetTriggers() []string
	GetAll() []event.Interface
	GetEventsByType(eventType string) []event.Interface
	RemoveEventByUUIDs(uuids ...string) []string
	RemoveTriggers(triggers ...string) []string
	GetPrioritySortedEventsByTrigger(triggerName string) []event.Interface
	ToggleTrigger(triggerName string, enable bool)
	IsTriggerEnabled(triggerName string) bool
}
