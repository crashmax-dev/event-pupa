package channelEx

import (
	"errors"
	"sync"
)

type channel[T any] struct {
	ch       chan T
	isClosed bool
	mx       sync.Mutex
}

func (c *channel[T]) Channel() chan T {
	return c.ch
}

func (c *channel[T]) IsClosed() bool {
	return c.isClosed
}

func (c *channel[T]) Close() error {
	c.mx.Lock()
	defer c.mx.Unlock()
	if !c.isClosed {
		close(c.ch)
		c.isClosed = true
		return nil
	}
	return errors.New("channel already closed")
}

func NewChannel(bufferSize int) Interface[string] {
	return &channel[string]{ch: make(chan string, bufferSize)}
}
