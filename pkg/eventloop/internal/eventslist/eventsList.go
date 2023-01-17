package eventslist

import (
	"fmt"
	"sync"

	"golang.org/x/exp/slices"
)

type eventsList struct {
	priorities map[string]priorityList
	mx         sync.Mutex
}

func New() Interface {
	result := eventsList{priorities: make(map[string]priorityList)}
	return &result
}

func (el *eventsList) EventName(triggerName string) Priority {
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
		return nil, fmt.Errorf("no such event name: %v", triggerName)
	}
	for _, priors := range el.priorities[triggerName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetUUID())
		}
	}
	return result, nil
}

func (eil *EventsByUUIDString) iterateDeletionEvents(uuids []string) []string {
	for id, ev := range *eil {
		if index := slices.Index(uuids, ev.GetUUID()); index != -1 {
			delete(*eil, id)
			uuids[index] = uuids[len(uuids)-1]
			uuids = uuids[:len(uuids)-1]

			if intervalComp, errInterval := ev.Interval(); errInterval == nil && intervalComp.IsRunning() {
				intervalComp.GetQuitChannel() <- true
			}

			if afterComp, afterErr := ev.After(); afterErr == nil {
				afterComp.GetBreakChannel() <- true
			}

			if sub, subErr := ev.Subscriber(); subErr == nil {
				sub.Exit() <- struct{}{}
			}

			if len(uuids) == 0 {
				return uuids
			}
		}
	}
	return uuids
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
func (el *eventsList) RemoveEventByUUIDs(ids ...string) []string {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.priorities {
		if modIds := eventNameValue.iterateDeletionPriorities(ids); len(modIds) != len(ids) {
			if len(eventNameValue) == 0 {
				delete(el.priorities, eventNameKey)
			}
			ids = modIds
			if len(ids) == 0 {
				return ids
			}
		}
	}
	return ids
}
