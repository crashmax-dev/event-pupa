package eventloop

type eventLoopEvent string

const (
	INTERVALED     eventLoopEvent = "@INTERVALED"
	BEFORE_TRIGGER eventLoopEvent = "@BEFORE"
	AFTER_TRIGGER  eventLoopEvent = "@AFTER"
)

var eventLoopEvents = []eventLoopEvent{INTERVALED, BEFORE_TRIGGER, AFTER_TRIGGER}
