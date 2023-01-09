package subscriber

import (
	"github.com/google/uuid"
)

type Interface interface {
	LockMutex()
	UnlockMutex()
	AddChannel(eventID uuid.UUID, infoCh chan SubChInfo, b *bool)
	Channels() channelCollection
	ChanTrigger() chan struct{}
	Exit() chan struct{}
	IsTrigger() bool
}

type InterfaceSubChannels interface {
	IsCLosed() bool
	SetIsClosed()
	GetInfoCh() chan SubChInfo
}
