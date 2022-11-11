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
	return el.data[eventName]
}

// GetEventIdsByName возвращает из eventloop массив айдишников объектов событий по имени события
// TODO: отрефакторить описание
func (el *eventsList) GetEventIdsByName(eventName string) (result []uuid.UUID, err error) {
	if el.data[eventName].Len() == 0 {
		return nil, errors.New(fmt.Sprintf("no such event name: %v", eventName))
	}
	for _, priors := range el.data[eventName] {
		for _, evnt := range priors {
			result = append(result, evnt.GetId())
		}
	}
	return result, nil
}

// RemoveEvent удаляет событие. Возвращает true если событие было в хранилище, false если не было
// TODO: перефакторить вложенные циклы в функции
func (el *eventsList) RemoveEvent(id uuid.UUID) bool {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.data {
		for priorKey, priorValue := range eventNameValue {
			for eventIdKey, eventIdValue := range priorValue {
				if eventIdValue.GetId() == id {

					delete(priorValue, eventIdKey)
					if len(priorValue) == 0 {
						delete(eventNameValue, priorKey)
						if len(eventNameValue) == 0 {
							delete(el.data, eventNameKey)
						}
					}
					return true
				}
			}
		}
	}
	return false
}
