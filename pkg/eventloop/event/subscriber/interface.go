package subscriber

type Interface interface {
	LockMutex()
	UnlockMutex()
	GetChannels() []chan int
	AddChannel(chan int)
	IsTrigger() bool
	SetIsTrigger(b bool)
}
