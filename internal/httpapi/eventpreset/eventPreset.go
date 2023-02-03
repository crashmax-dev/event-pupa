package eventpreset

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/YSX/eventloop/pkg/eventloop/event"
)

type (
	EventFunc func() func(ctx context.Context) string
	EventType uint8
)

const (
	REGULAR EventType = iota + 1
	INTERVALED
)

var Events = [...]EventFunc{event1, event2}

// CreateEvent создаёт событие из пресета id с типом eventType (типы REGULAR, INTERVALED).
// Возвращает ошибку, если такого пресета нет
// Интервал интервального ивента 500 ms
func CreateEvent(id int, eventType EventType, triggerName string) (event.Interface, error) {
	switch eventType {
	case REGULAR:
		return event.NewEvent(event.Args{Fun: Events[id-1](), TriggerName: triggerName})
	case INTERVALED:
		return event.NewEvent(event.Args{Fun: Events[id-1](), IntervalTime: 500 * time.Millisecond})
	default:
		return nil, fmt.Errorf("No such type: %v", eventType)
	}
}

func event1() func(ctx context.Context) string {
	var number int
	fn := func(ctx context.Context) string {
		number++
		fmt.Println(number)
		return strconv.Itoa(number)
	}
	return fn

}

func event2() func(ctx context.Context) string {
	var number int
	fn := func(ctx context.Context) string {
		number--
		return strconv.Itoa(number)
	}
	return fn
}
