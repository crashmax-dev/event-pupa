package subscriber

import (
	"sync"
)

type SubChannel struct {
	infoCh   chan SubChInfo
	isClosed *bool
	mx       sync.RWMutex
}

func (sc *SubChannel) GetInfoCh() chan SubChInfo {
	return sc.infoCh
}

func (sc *SubChannel) IsCLosed() bool {
	sc.mx.RLock()
	defer sc.mx.RUnlock()
	return *sc.isClosed
}

func (sc *SubChannel) SetIsClosed() {
	sc.mx.Lock()
	defer sc.mx.Unlock()
	*sc.isClosed = true
}
