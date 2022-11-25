package eventslist

import "eventloop/pkg/eventloop/event"

type EventIdsList map[string]event.Interface

func (eil *EventIdsList) List() EventIdsList {
	return *eil
}

func (eil *EventIdsList) AddEvent(newEvent event.Interface) {
	(*eil)[newEvent.GetID().String()] = newEvent
}

func (eil *EventIdsList) EventID(eventID string) event.Interface {
	return (*eil)[eventID]
}
