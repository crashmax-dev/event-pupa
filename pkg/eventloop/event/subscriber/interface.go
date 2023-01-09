package subscriber

import (
	"github.com/google/uuid"
)

type Interface interface {
	LockMutex()
	UnlockMutex()
	AddChannel(eventID uuid.UUID, infoCh chan SubChInfo, b *bool)
	Channels() channelCollection
	Trigger() chan struct{}
	Exit() chan struct{}
	IsTrigger() bool
	SetIsTrigger(b bool)
}

type InterfaceSubChannels interface {
	IsCLosed() bool
	SetIsClosed()
	GetInfoCh() chan SubChInfo
}
