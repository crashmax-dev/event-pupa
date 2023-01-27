package triggerslist

import (
	"sync"
)

type Triggers map[string]*priorityList

type eventsList struct {
	prioritiesByTrigger Triggers
	mx                  sync.Mutex
}

func New() Interface {
	result := eventsList{prioritiesByTrigger: make(map[string]*priorityList)}
	return &result
}

// EventName
func (el *eventsList) TriggerName(triggerName string) Priority {
	if el.prioritiesByTrigger[triggerName] == nil {
		el.prioritiesByTrigger[triggerName] = &priorityList{data: make(map[int]EventsByUUIDString)}
	}
	return el.prioritiesByTrigger[triggerName]
}

func (el *eventsList) RemoveTriggers(triggers ...string) []string {
	panic("Not implemented")
	return []string{}
}

// RemoveEventByUUIDs удаляет события. Возвращает пустой срез, если всё удалено, или срез айдишек, которые не были
// найдены и не удалены.
func (el *eventsList) RemoveEventByUUIDs(uuids ...string) []string {
	el.mx.Lock()
	defer el.mx.Unlock()
	for eventNameKey, eventNameValue := range el.prioritiesByTrigger {
		if modIds := eventNameValue.iterateDeletionPriorities(uuids); len(modIds) != len(uuids) {
			if len(eventNameValue.data) == 0 {
				delete(el.prioritiesByTrigger, eventNameKey)
			}
			uuids = modIds
			if len(uuids) == 0 {
				return uuids
			}
		}
	}
	return uuids
}
