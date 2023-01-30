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

func (eil *EventsByUUIDString) RemoveEvent(ev event.Interface) {
	delete(*eil, ev.GetUUID())

	if intervalComp, errInterval := ev.Interval(); errInterval == nil && intervalComp.IsRunning() {
		intervalComp.GetQuitChannel() <- true
	}

	if afterComp, afterErr := ev.After(); afterErr == nil {
		afterComp.GetBreakChannel() <- true
	}

	if sub, subErr := ev.Subscriber(); subErr == nil {
		sub.Exit() <- struct{}{}
	}
}

func (eil *EventsByUUIDString) Event(eventUUIDString string) event.Interface {
	return (*eil)[eventUUIDString]
}

func (eil *EventsByUUIDString) iterateDeletionEvents(uuids []string) (remainings []string) {
	remainings = uuids[:0]
	for _, id := range uuids {
		ev := eil.Event(id)
		if ev == nil {
			remainings = append(remainings, id)
			continue
		}
		eil.RemoveEvent(ev)
	}
	return remainings
}

func (eil *EventsByUUIDString) clearEvents() {
	for _, e := range *eil {
		eil.RemoveEvent(e)
	}
}
