package eventloop

type eventLoopSystemEvent string

const (
	INTERVALED     eventLoopSystemEvent = "@INTERVALED"
	BEFORE_TRIGGER eventLoopSystemEvent = "@BEFORE"
	AFTER_TRIGGER  eventLoopSystemEvent = "@AFTER"
)

const (
	BEFORE_PRIORITY = -2
	AFTER_PRIORITY  = -1
)

var eventLoopEvents = []eventLoopSystemEvent{INTERVALED, BEFORE_TRIGGER, AFTER_TRIGGER}
