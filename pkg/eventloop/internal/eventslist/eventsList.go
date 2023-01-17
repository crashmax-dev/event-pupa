package eventslist

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
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
		el.priorities[triggerName] = make(map[int]EventsByUuidString)
	}
	result := el.priorities[triggerName]
	return &result
}

// GetEventIdsByName возвращает из eventloop список айдишек событий, повешенных на событие eventName
func (el *eventsList) GetEventIdsByTriggerName(triggerName string) (result []uuid.UUID, err error) {
	pl := el.priorities[triggerName]
	if pl.Len() == 0 {
		return nil, fmt.Errorf("no such event name: %v", triggerName)
	}
	for _, priors := range el.priorities[triggerName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetID())
		}
	}
	return result, nil
}

func (eil *EventsByUuidString) iterateDeletionEvents(ids []uuid.UUID) []uuid.UUID {
	for id, ev := range *eil {
		if index := slices.Index(ids, ev.GetID()); index != -1 {
			delete(*eil, id)
			ids[index] = ids[len(ids)-1]
			ids = ids[:len(ids)-1]

			if intervalComp, errInterval := ev.Interval(); errInterval == nil && intervalComp.IsRunning() {
				intervalComp.GetQuitChannel() <- true
			}

			if afterComp, afterErr := ev.After(); afterErr == nil {
				afterComp.GetBreakChannel() <- true
			}

			if len(ids) == 0 {
				return ids
			}
		}
	}
	return ids
}

func (pl *priorityList) iterateDeletionPriorities(ids []uuid.UUID) []uuid.UUID {
	for priorKey, priorValue := range *pl {
		if modIds := priorValue.iterateDeletionEvents(ids); len(modIds) != len(ids) {
			if len(priorValue) == 0 {
				delete(*pl, priorKey)
			}
			ids = modIds
			if len(ids) == 0 {
				return ids
			}
		}
	}
	return ids
}

// RemoveEventByUUIDs удаляет события. Возвращает пустой срез, если всё удалено, или срез айдишек, которые не были
// найдены и не удалены.
func (el *eventsList) RemoveEventByUUIDs(ids ...uuid.UUID) []uuid.UUID {
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
