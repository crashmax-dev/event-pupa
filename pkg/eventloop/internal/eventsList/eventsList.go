package eventsList

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type eventsList struct {
	data map[string]priorityList
	mx   sync.Mutex
}

func New() Interface {
	result := eventsList{data: make(map[string]priorityList)}
	return &result
}

func (el *eventsList) EventName(eventName string) Priority {
	if el.data[eventName] == nil {
		el.data[eventName] = make(priorityList)
	}
	result := el.data[eventName]
	return &result
}

// GetEventIdsByName возвращает из eventloop список айдишек событий, повешенных на событие eventName
func (el *eventsList) GetEventIdsByName(eventName string) (result []uuid.UUID, err error) {
	pl := el.data[eventName]
	if pl.Len() == 0 {
		return nil, errors.New(fmt.Sprintf("no such event name: %v", eventName))
	}
	for _, priors := range el.data[eventName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetId())
		}
	}
	return result, nil
}

func (eil *EventIdsList) iterateEvents(id uuid.UUID) bool {
	for eventIdKey, eventIdValue := range *eil {
		if eventIdValue.GetId() == id {
			delete(*eil, eventIdKey)
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
	for eventNameKey, eventNameValue := range el.data {
		if eventNameValue.iteratePriorities(id) {
			if len(eventNameValue) == 0 {
				delete(el.data, eventNameKey)
			}
			return true
		}
	}
	return false
}
