package eventslist

import (
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

type Interface interface {
	EventName(eventName string) Priority
	RemoveEventByUUIDs(id []uuid.UUID) []uuid.UUID
	GetEventIdsByName(eventName string) (result []uuid.UUID, err error)
}

type Priority interface {
	Priority(priority int) *EventIdsList
	Len() int
}

type EventID interface {
	List() EventIdsList
	EventID(eventID string) event.Interface
	AddEvent(newEvent event.Interface)
}
