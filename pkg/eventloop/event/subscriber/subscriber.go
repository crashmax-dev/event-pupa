package subscriber

import (
	"sync"

	"github.com/google/uuid"
)

type SubChInfo int

type channelCollection map[uuid.UUID]InterfaceSubChannels

const (
	TriggerListener SubChInfo = 1
	DeleteChannel   SubChInfo = -1
)

// eventSubscriber - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type eventSubscriber struct {
	isTrigger bool
	trigger   chan struct{}

	channels channelCollection
	exit     chan struct{}
	mx       sync.Mutex
}

func NewSubscriberEvent() Interface {
	return &eventSubscriber{channels: make(channelCollection), exit: make(chan struct{})}
}

func NewTriggerEvent() Interface {
	return &eventSubscriber{channels: make(channelCollection), isTrigger: true}
}

func (ev *eventSubscriber) LockMutex() {
	ev.mx.Lock()
}

func (ev *eventSubscriber) UnlockMutex() {
	ev.mx.Unlock()
}

func (ev *eventSubscriber) Channels() channelCollection {
	return ev.channels
}

// AddChannel добавляет каналы, по которым тригеррится событие, и по этим же каналам событие-триггер триггерит
// события слушатели.
func (ev *eventSubscriber) AddChannel(eventID uuid.UUID, infoCh chan SubChInfo, b *bool) {
	ev.channels[eventID] = &SubChannel{infoCh: infoCh, isClosed: b}
}

func (ev *eventSubscriber) IsTrigger() bool {
	return ev.isTrigger
}

func (ev *eventSubscriber) SetIsTrigger(b bool) {
	ev.isTrigger = b
	ev.trigger = make(chan struct{})
}

func (ev *eventSubscriber) Trigger() chan struct{} {
	return ev.trigger
}

func (ev *eventSubscriber) Exit() chan struct{} {
	return ev.exit
}
