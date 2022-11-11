package eventsList

type priorityList map[int]EventIdsList

func (pl priorityList) Priority(priority int) EventId {
	if pl[priority] == nil {
		pl[priority] = make(EventIdsList)
	}
	return pl[priority]
}

func (pl priorityList) Len() int {
	return len(pl)
}
