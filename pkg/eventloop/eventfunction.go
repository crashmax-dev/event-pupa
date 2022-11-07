package eventloop

import "errors"

type EventFunction uint8

const (
	ON EventFunction = iota + 1
	TRIGGER
)

func (ef EventFunction) String() (string, error) {
	for k, v := range EventFunctionMapping {
		if v == ef {
			return k, nil
		}
	}
	return "", errors.New("no such Event Function")
}

var EventFunctionMapping = map[string]EventFunction{
	"ON":      ON,
	"TRIGGER": TRIGGER,
}
