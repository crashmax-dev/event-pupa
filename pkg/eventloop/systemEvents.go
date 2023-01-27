package eventloop

type eventLoopSystemTrigger string

type AllTriggers struct {
	userTriggers   []string
	systemTriggers []eventLoopSystemTrigger
}

var (
	ReturnIriggers = AllTriggers{
		systemTriggers: allSystemTriggers,
	}
)

const (
	INTERVALED     eventLoopSystemTrigger = "@INTERVALED"
	AFTER          eventLoopSystemTrigger = "@AFTER"
	BEFORE_TRIGGER eventLoopSystemTrigger = "@BEFORE_TRIGGER"
	AFTER_TRIGGER  eventLoopSystemTrigger = "@AFTER_TRIGGER"
	BEFORE_CREATE  eventLoopSystemTrigger = "@BEFORE_CREATE"
	AFTER_CREATE   eventLoopSystemTrigger = "@AFTER_CREATE"
)

const (
	BEFORE_PRIORITY = -2
	AFTER_PRIORITY  = -1
)

var allSystemTriggers = []eventLoopSystemTrigger{INTERVALED,
	BEFORE_TRIGGER,
	AFTER_TRIGGER,
	BEFORE_CREATE,
	AFTER_CREATE,
	AFTER}

var restrictedTriggers = []eventLoopSystemTrigger{
	INTERVALED,
	AFTER,
}
