package eventslist

import (
	"fmt"
	"sync"
)

type eventsList struct {
	priorities map[string]priorityList
	mx         sync.Mutex
}

func New() Interface {
	result := eventsList{priorities: make(map[string]priorityList)}
	return &result
}

// EventName
func (el *eventsList) TriggerName(triggerName string) Priority {
	if el.priorities[triggerName] == nil {
		el.priorities[triggerName] = make(map[int]EventsByUUIDString)
	}
	result := el.priorities[triggerName]
	return &result
}

// GetEventIdsByName возвращает из eventloop список айдишек событий, повешенных на событие eventName
func (el *eventsList) GetEventIdsByTriggerName(triggerName string) (result []string, err error) {
	pl := el.priorities[triggerName]
	if pl.Len() == 0 {
		return nil, fmt.Errorf("no such trigger name: %v", triggerName)
	}
	for _, priors := range el.priorities[triggerName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetUUID())
		}
	}
	return result, nil
}

func (eil *EventsByUUIDString) iterateDeletionEvents(uuids []string) (remainings []string) {
	remainings = uuids[:0]
	for _, id := range uuids {
		ev := eil.EventID(id)
		if ev == nil {
			remainings = append(remainings, id)
			continue
		}
		delete(*eil, id)

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
	return remainings
}

func (pl *priorityList) iterateDeletionPriorities(uuids []string) []string {
	for priorKey, priorValue := range *pl {
		if modIds := priorValue.iterateDeletionEvents(uuids); len(modIds) != len(uuids) {
			if len(priorValue) == 0 {
				delete(*pl, priorKey)
			}
			uuids = modIds
			if len(uuids) == 0 {
				return uuids
			}
		}
	}
	return uuids
}

// RemoveEventByUUIDs удаляет события. Возвращает пустой срез, если всё удалено, или срез айдишек, которые не были
// найдены и не удалены.
func (el *eventsList) RemoveEventByUUIDs(uuids ...string) []string {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.priorities {
		if modIds := eventNameValue.iterateDeletionPriorities(uuids); len(modIds) != len(uuids) {
			if len(eventNameValue) == 0 {
				delete(el.priorities, eventNameKey)
			}
			uuids = modIds
			if len(uuids) == 0 {
				return uuids
			}
		}
	}
	return uuids
}
