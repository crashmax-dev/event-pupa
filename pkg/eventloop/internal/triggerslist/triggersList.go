package triggerslist

import (
	"sync"
)

type Triggers map[string]*priorityList

type eventsList struct {
	priorities Triggers
	mx         sync.Mutex
}

func New() Interface {
	result := eventsList{priorities: make(map[string]*priorityList)}
	return &result
}

// EventName
func (el *eventsList) TriggerName(triggerName string) Priority {
	if el.priorities[triggerName] == nil {
		el.priorities[triggerName] = &priorityList{data: make(map[int]EventsByUUIDString)}
	}
	return el.priorities[triggerName]
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
	for priorKey, priorValue := range pl.data {
		if modIds := priorValue.iterateDeletionEvents(uuids); len(modIds) != len(uuids) {
			if len(priorValue) == 0 {
				delete(pl.data, priorKey)
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
			if len(eventNameValue.data) == 0 {
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
