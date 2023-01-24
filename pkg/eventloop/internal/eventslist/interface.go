package eventslist

import (
	"eventloop/pkg/eventloop/event"
)

type Interface interface {
	TriggerName(triggerName string) Priority
	RemoveEventByUUIDs(uuids ...string) []string
	GetEventIdsByTriggerName(triggerName string) (result []string, err error)
}

type Priority interface {
	Priority(priority int) *EventsByUUIDString
	Len() int
	GetKeys() (keys []int)
}

type EventID interface {
	List() EventsByUUIDString
	EventID(eventID string) event.Interface
	AddEvent(newEvent event.Interface)
}
