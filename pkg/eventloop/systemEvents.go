package eventloop

type eventLoopSystemEvent string

const (
	INTERVALED     eventLoopSystemEvent = "@INTERVALED"
	AFTER          eventLoopSystemEvent = "@AFTER"
	BEFORE_TRIGGER eventLoopSystemEvent = "@BEFORE_TRIGGER"
	AFTER_TRIGGER  eventLoopSystemEvent = "@AFTER_TRIGGER"
	BEFORE_CREATE  eventLoopSystemEvent = "@BEFORE_CREATE"
	AFTER_CREATE   eventLoopSystemEvent = "@AFTER_CREATE"
)

const (
	BEFORE_PRIORITY = -2
	AFTER_PRIORITY  = -1
)

var eventLoopEvents = []eventLoopSystemEvent{INTERVALED,
	BEFORE_TRIGGER,
	AFTER_TRIGGER,
	BEFORE_CREATE,
	AFTER_CREATE,
	AFTER}

var restrictedEvents = []eventLoopSystemEvent{
	INTERVALED,
	AFTER,
}
