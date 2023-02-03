package eventsContainer

import (
	"sort"
	"sync"

	"gitlab.com/YSX/eventloop/pkg/eventloop/event"
	"golang.org/x/exp/maps"
)

type (
	eventsMap            map[string]event.Interface
	criteriaName         string
	eventsByCriteriaName map[criteriaName]eventsByCriteria
	eventsByCriteria     map[string]criteriaInfo
)

const (
	TRIGGER  criteriaName = "TRIGGER"
	TYPE     criteriaName = "TYPE"
	PRIORITY criteriaName = "PRIORITY"
)

type criteriaInfo struct {
	data      eventsMap
	isEnabled bool
}

type eventsList struct {
	events eventsMap

	eventsByCriteria eventsByCriteriaName

	mx sync.Mutex
}

func New() Interface {
	result := eventsList{
		events: make(eventsMap),
		eventsByCriteria: eventsByCriteriaName{
			"TRIGGER":  make(eventsByCriteria),
			"TYPE":     make(eventsByCriteria),
			"PRIORITY": make(eventsByCriteria),
		},
	}
	return &result
}

func (el *eventsList) AddEvent(newEvent event.Interface) {
	el.events[newEvent.GetUUID()] = newEvent
	for criteria, events := range el.eventsByCriteria {
		switch criteria {
		case "TRIGGER":
			triggerName := newEvent.GetTriggerName()
			events[triggerName] = el.addToMap(events[triggerName], newEvent)
		case "TYPE":
			for _, t := range newEvent.GetTypes() {
				events[string(t)] = el.addToMap(
					events[string(t)],
					newEvent,
				)
			}
		case "PRIORITY":
			priority := newEvent.GetPriorityString()
			events[priority] = el.addToMap(events[priority], newEvent)
		default:
			panic(criteria + " add event not implemented")
		}
	}
}

// Вернули ошибку, а сверху инициализируем красиво
func (el *eventsList) addToMap(criteria criteriaInfo, newEvent event.Interface) criteriaInfo {
	if criteria.data == nil {
		criteria = criteriaInfo{data: make(map[string]event.Interface), isEnabled: true}
	}
	criteria.data[newEvent.GetUUID()] = newEvent
	return criteria
}

// EventName
func (el *eventsList) EventsByTrigger(triggerName string) []event.Interface {
	return maps.Values(el.eventsByCriteria[TRIGGER][triggerName].data)
}

func (el *eventsList) GetTriggers() []string {
	return maps.Keys(el.eventsByCriteria[TRIGGER])
}

func (el *eventsList) GetAll() (result []event.Interface) {
	return maps.Values(el.events)
}

func (el *eventsList) GetEventsByType(eventType string) []event.Interface {
	return maps.Values(el.eventsByCriteria[TYPE][eventType].data)
}

// RemoveEventByUUIDs удаляет события. Возвращает пустой срез, если всё удалено, или срез айдишек, которые не были
// найдены и не удалены.
func (el *eventsList) RemoveEventByUUIDs(uuids ...string) (result []string) {
	el.mx.Lock()
	defer el.mx.Unlock()

	result = make([]string, 0, len(uuids))

	for _, uuid := range uuids {
		if ev, ok := el.events[uuid]; ok {
			el.removeByTrigger(ev)
			el.removeByTypes(ev)
			el.removeByPriority(ev)

			// Удаляем из общего хранилища
			delete(el.events, uuid)
		} else {
			result = append(result, uuid)
		}
	}

	return result
}

func (el *eventsList) RemoveTriggers(triggers ...string) (result []string) {
	el.mx.Lock()
	defer el.mx.Unlock()

	result = make([]string, 0, len(triggers))

	for _, trig := range triggers {
		if _, ok := el.eventsByCriteria[TRIGGER][trig]; ok {
			el.removeTrigger(trig)
		} else {
			result = append(result, trig)
		}
	}
	return result
}

func (el *eventsList) GetPrioritySortedEventsByTrigger(triggerName string) []event.Interface {
	result := maps.Values(el.eventsByCriteria[TRIGGER][triggerName].data)
	sort.Slice(
		result, func(i, j int) bool {
			return result[i].GetPriority() > result[j].GetPriority()
		},
	)
	return result
}

func (el *eventsList) ToggleTrigger(triggerName string, enable bool) {
	triggerInfo, ok := el.eventsByCriteria[TRIGGER][triggerName]
	if !ok {
		triggerInfo = criteriaInfo{data: make(eventsMap)}
	}
	triggerInfo.isEnabled = enable
	el.eventsByCriteria[TRIGGER][triggerName] = triggerInfo
}

func (el *eventsList) IsTriggerEnabled(triggerName string) bool {
	if triggerInfo, ok := el.eventsByCriteria[TRIGGER][triggerName]; ok {
		return triggerInfo.isEnabled
	}
	return true
}
