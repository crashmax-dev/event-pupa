package eventloop

import "errors"

type EventFunction uint8

const (
	TRIGGER EventFunction = iota + 1
	REGISTER
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
	"TRIGGER":  TRIGGER,
	"REGISTER": REGISTER,
}
