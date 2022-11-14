package eventsList

import "eventloop/pkg/eventloop/event"

type EventIdsList map[string]event.Interface

func (eil *EventIdsList) List() EventIdsList {
	return *eil
}

func (eil *EventIdsList) AddEvent(newEvent event.Interface) {
	(*eil)[newEvent.GetId().String()] = newEvent
}

func (eil *EventIdsList) EventId(eventId string) event.Interface {
	return (*eil)[eventId]
}
