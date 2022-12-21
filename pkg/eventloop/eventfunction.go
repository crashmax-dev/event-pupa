package eventloop

type EventFunction uint8

const (
	TRIGGER EventFunction = iota + 1
	REGISTER
)

func (ef EventFunction) String() string {
	for k, v := range EventFunctionMapping {
		if v == ef {
			return k
		}
	}
	return ""
}

var EventFunctionMapping = map[string]EventFunction{
	"TRIGGER":  TRIGGER,
	"REGISTER": REGISTER,
}
