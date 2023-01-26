package triggerslist

import (
	"eventloop/pkg/eventloop/event"
)

type EventsByUUIDString map[string]event.Interface

func (eil *EventsByUUIDString) List() EventsByUUIDString {
	return *eil
}

func (eil *EventsByUUIDString) AddEvent(newEvent event.Interface) {
	(*eil)[newEvent.GetUUID()] = newEvent
}

func (eil *EventsByUUIDString) Event(eventUUIDString string) event.Interface {
	return (*eil)[eventUUIDString]
}
