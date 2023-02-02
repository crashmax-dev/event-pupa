package subscriber

import (
	"sync"
)

type SubChInfo int

type Type string

type channelsByUUIDString map[string]InterfaceSubChannels

const (
	TriggerListener SubChInfo = 1
)

const (
	Listener Type = "LISTENER"
	Trigger  Type = "TRIGGER"
)

// component - событие, которое триггерит другие события и/или триггерится само по другим событиям.
type component struct {
	trigger   chan struct{}
	isRunning bool

	channels channelsByUUIDString
	exit     chan struct{}
	mx       sync.Mutex
	esType   Type
}

func NewSubscriberEvent() Interface {
	return &component{channels: make(channelsByUUIDString),
		exit:   make(chan struct{}),
		esType: Listener}
}

func NewTriggerEvent() Interface {
	return &component{channels: make(channelsByUUIDString),
		trigger: make(chan struct{}),
		exit:    make(chan struct{}),
		esType:  Trigger}
}

func (ev *component) LockMutex() {
	ev.mx.Lock()
}

func (ev *component) UnlockMutex() {
	ev.mx.Unlock()
}

func (ev *component) Channels() channelsByUUIDString {
	return ev.channels
}

// AddChannel добавляет каналы, по которым тригеррится событие, и по этим же каналам событие-триггер триггерит
// события слушатели.
func (ev *component) AddChannel(eventUUID string, infoCh chan SubChInfo, b *bool) {
	ev.channels[eventUUID] = &SubChannel{infoCh: infoCh, isClosed: b}
}

func (ev *component) ChanTrigger() chan struct{} {
	return ev.trigger
}

func (ev *component) Exit() chan struct{} {
	return ev.exit
}

func (ev *component) GetType() Type {
	return ev.esType
}

func (ev *component) IsRunning() bool {
	return ev.isRunning
}

func (ev *component) SetIsRunning(b bool) {
	ev.isRunning = b
}
