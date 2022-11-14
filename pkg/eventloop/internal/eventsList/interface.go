package eventsList

import (
	"eventloop/pkg/eventloop/event"
	"github.com/google/uuid"
)

type Interface interface {
	EventName(eventName string) Priority
	RemoveEvent(id uuid.UUID) bool
	GetEventIdsByName(eventName string) (result []uuid.UUID, err error)
}

type Priority interface {
	Priority(priority int) *EventIdsList
	Len() int
}

type EventId interface {
	List() EventIdsList
	EventId(eventId string) event.Interface
	AddEvent(newEvent event.Interface)
}
