package subscriber

import (
	"sync"
)

type SubChInfo int

type Type string

type channelsByUUIDString map[string]InterfaceSubChannels

const (
	TriggerListener SubChInfo = 1
	DeleteChannel   SubChInfo = -1
)

const (
	Listener Type = "LISTENER"
	Trigger  Type = "TRIGGER"
)

// eventSubscriber - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type eventSubscriber struct {
	trigger chan struct{}

	channels channelsByUUIDString
	exit     chan struct{}
	mx       sync.Mutex
	esType   Type
}

func NewSubscriberEvent() Interface {
	return &eventSubscriber{channels: make(channelsByUUIDString),
		exit:   make(chan struct{}),
		esType: Listener}
}

func NewTriggerEvent() Interface {
	return &eventSubscriber{channels: make(channelsByUUIDString),
		trigger: make(chan struct{}),
		exit:    make(chan struct{}),
		esType:  Trigger}
}

func (ev *eventSubscriber) LockMutex() {
	ev.mx.Lock()
}

func (ev *eventSubscriber) UnlockMutex() {
	ev.mx.Unlock()
}

func (ev *eventSubscriber) Channels() channelsByUUIDString {
	return ev.channels
}

// AddChannel добавляет каналы, по которым тригеррится событие, и по этим же каналам событие-триггер триггерит
// события слушатели.
func (ev *eventSubscriber) AddChannel(eventUUID string, infoCh chan SubChInfo, b *bool) {
	ev.channels[eventUUID] = &SubChannel{infoCh: infoCh, isClosed: b}
}

func (ev *eventSubscriber) ChanTrigger() chan struct{} {
	return ev.trigger
}

func (ev *eventSubscriber) Exit() chan struct{} {
	return ev.exit
}

func (ev *eventSubscriber) GetType() Type {
	return ev.esType
}
