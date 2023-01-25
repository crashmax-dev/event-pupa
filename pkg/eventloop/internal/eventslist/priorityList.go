package eventslist

import (
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type priorityList map[int]EventsByUUIDString

func (pl *priorityList) Priority(priority int) *EventsByUUIDString {
	if (*pl)[priority] == nil {
		(*pl)[priority] = make(EventsByUUIDString)
	}
	result := (*pl)[priority]
	return &result
}

func (pl *priorityList) Len() int {
	return len(*pl)
}

// GetKeys выводит все приоритеты по возрастанию
func (pl *priorityList) GetSortedPriorityNums() (keys []int) {
	keys = maps.Keys((*pl).data)
	slices.Sort(keys)
	return keys
}
