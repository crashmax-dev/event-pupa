package eventslist

type priorityList map[int]EventIdsList

func (pl *priorityList) Priority(priority int) *EventIdsList {
	if (*pl)[priority] == nil {
		(*pl)[priority] = make(EventIdsList)
	}
	result := (*pl)[priority]
	return &result
}

func (pl *priorityList) Len() int {
	return len(*pl)
}
