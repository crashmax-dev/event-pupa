package triggerslist

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

func (pl *priorityList) clearPriorities() {
	for priorityKey, p := range pl.data {
		p.clearEvents()
		delete(pl.data, priorityKey)
	}
}
