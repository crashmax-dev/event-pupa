package channelEx

type Interface[T any] interface {
	Channel() chan T
	IsClosed() bool
	Close() error
}
