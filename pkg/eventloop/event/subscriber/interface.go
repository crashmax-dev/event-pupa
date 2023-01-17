package subscriber

type Interface interface {
	LockMutex()
	UnlockMutex()
	AddChannel(eventUUID string, infoCh chan SubChInfo, b *bool)
	Channels() channelsByUUIDString
	ChanTrigger() chan struct{}
	Exit() chan struct{}
	GetType() Type
}

type InterfaceSubChannels interface {
	IsClosed() bool
	SetIsClosed()
	GetInfoCh() chan SubChInfo
}
