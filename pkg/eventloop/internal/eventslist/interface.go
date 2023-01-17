package eventslist

import (
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

type Interface interface {
	EventName(eventName string) Priority
	RemoveEventByUUIDs(ids ...uuid.UUID) []uuid.UUID
	GetEventIdsByTriggerName(triggerName string) (result []uuid.UUID, err error)
}

type Priority interface {
	Priority(priority int) *EventsByUuidString
	Len() int
	GetKeys() (keys []int)
}

type EventID interface {
	List() EventsByUuidString
	EventID(eventID string) event.Interface
	AddEvent(newEvent event.Interface)
}
