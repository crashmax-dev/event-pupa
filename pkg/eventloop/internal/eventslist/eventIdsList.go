package eventslist

import (
	"eventloop/pkg/eventloop/event"
)

type EventsByUuidString map[string]event.Interface

func (eil *EventsByUuidString) List() EventsByUuidString {
	return *eil
}

func (eil *EventsByUuidString) AddEvent(newEvent event.Interface) {
	(*eil)[newEvent.GetID().String()] = newEvent
}

func (eil *EventsByUuidString) EventID(eventUuidString string) event.Interface {
	return (*eil)[eventUuidString]
}
