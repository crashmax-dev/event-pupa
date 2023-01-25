package eventslist

import (
	"fmt"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type Priorities map[int]EventsByUUIDString

type priorityList struct {
	data     Priorities
	disabled bool
}

func (pl *priorityList) Priority(priority int) *EventsByUUIDString {
	if pl.data[priority] == nil {
		pl.data[priority] = make(EventsByUUIDString)
	}
	result := pl.data[priority]
	return &result
}

func (pl *priorityList) Len() int {
	return len(pl.data)
}

// GetKeys выводит все приоритеты по возрастанию
func (pl *priorityList) GetSortedPriorityNums() (keys []int) {
	keys = maps.Keys(pl.data)
	slices.Sort(keys)
	return keys
}

func (pl *priorityList) IsDisabled() bool {
	return pl.disabled
}

func (pl *priorityList) SetIsDisabled(b bool) {
	pl.disabled = b
}

func (el *priorityList) GetAllEvents() (result []string, err error) {
	for _, priors := range el.data {
		for _, evnt := range priors {
			result = append(result, evnt.GetUUID())
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no events")
	}
	return result, nil
}
