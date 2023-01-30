package triggerslist

import (
	"sync"
)

type Triggers map[string]*priorityList

type eventsList struct {
	prioritiesByTriggerName Triggers
	mx                      sync.Mutex
}

func New() Interface {
	result := eventsList{prioritiesByTriggerName: make(map[string]*priorityList)}
	return &result
}

// EventName
func (el *eventsList) TriggerName(triggerName string) Priority {
	if el.prioritiesByTriggerName[triggerName] == nil {
		el.prioritiesByTriggerName[triggerName] = &priorityList{data: make(map[int]EventsByUUIDString)}
	}
	return el.prioritiesByTriggerName[triggerName]
}

func (el *eventsList) GetAll() (result []string) {
	result = make([]string, 0, len(el.prioritiesByTriggerName))
	for k := range el.prioritiesByTriggerName {
		result = append(result, k)
	}
	return
}

func (el *eventsList) RemoveTriggers(triggers ...string) (result []string) {
	el.mx.Lock()
	defer el.mx.Unlock()

	for _, v := range triggers {
		if _, ok := el.prioritiesByTriggerName[v]; !ok {
			result = append(result, v)
			continue
		}
		el.prioritiesByTriggerName[v].clearPriorities()
		delete(el.prioritiesByTriggerName, v)
	}
	return result
}

// RemoveEventByUUIDs удаляет события. Возвращает пустой срез, если всё удалено, или срез айдишек, которые не были
// найдены и не удалены.
func (el *eventsList) RemoveEventByUUIDs(uuids ...string) []string {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.prioritiesByTriggerName {
		if modIds := eventNameValue.iterateDeletionPriorities(uuids); len(modIds) != len(uuids) {
			if len(eventNameValue.data) == 0 {
				delete(el.prioritiesByTriggerName, eventNameKey)
			}
			uuids = modIds
			if len(uuids) == 0 {
				return uuids
			}
		}
	}
	return uuids
}
