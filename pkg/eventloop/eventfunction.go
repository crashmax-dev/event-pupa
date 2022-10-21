package eventloop

type EventFunction uint8

const (
	ON EventFunction = iota + 1
	TRIGGER
)

func (ef EventFunction) String() string {
	switch ef {
	case ON:
		return "ON"
	case TRIGGER:
		return "TRIGGER"
	default:
		panic("EventFunction is not added!")
	}
}

var EventFunctionMapping = map[string]EventFunction{
	"ON":      ON,
	"TRIGGER": TRIGGER,
}
