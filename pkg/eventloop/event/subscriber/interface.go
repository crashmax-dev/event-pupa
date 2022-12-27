package subscriber

import (
	"github.com/google/uuid"
)

type Interface interface {
	LockMutex()
	UnlockMutex()
	GetChannels() channelCollection
	AddChannel(eventID uuid.UUID, ch chan int)
	IsTrigger() bool
	SetIsTrigger(b bool)
}
