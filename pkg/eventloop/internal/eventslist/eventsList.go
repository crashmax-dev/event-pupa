package eventslist

import (
	"fmt"
	"github.com/google/uuid"
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

func (el *eventsList) EventName(eventName string) Priority {
	if el.priorities[eventName] == nil {
		el.priorities[eventName] = make(map[int]EventIdsList)
	}
	result := el.priorities[eventName]
	return &result
}

// GetEventIdsByName возвращает из eventloop список айдишек событий, повешенных на событие eventName
func (el *eventsList) GetEventIdsByName(eventName string) (result []uuid.UUID, err error) {
	pl := el.priorities[eventName]
	if pl.Len() == 0 {
		return nil, fmt.Errorf("no such event name: %v", eventName)
	}
	for _, priors := range el.priorities[eventName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetID())
		}
	}
	return result, nil
}

func (eil *EventIdsList) iterateEvents(id uuid.UUID) bool {
	for eventIDKey, eventIdValue := range *eil {
		if eventIdValue.GetID() == id {
			delete(*eil, eventIDKey)
			return true
		}
	}
	return false
}

func (pl *priorityList) iteratePriorities(id uuid.UUID) bool {
	for priorKey, priorValue := range *pl {
		if priorValue.iterateEvents(id) {
			if len(priorValue) == 0 {
				delete(*pl, priorKey)
			}
			return true
		}
	}
	return false
}

// RemoveEvent удаляет событие. Возвращает true если событие было в хранилище, false если не было
func (el *eventsList) RemoveEvent(id uuid.UUID) bool {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.priorities {
		if eventNameValue.iteratePriorities(id) {
			if len(eventNameValue) == 0 {
				delete(el.priorities, eventNameKey)
			}
			return true
		}
	}
	return false
}
