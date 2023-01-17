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

func (pl *priorityList) GetKeys() (keys []int) {
	keys = maps.Keys(*pl)
	slices.Sort(keys)
	return keys
}
