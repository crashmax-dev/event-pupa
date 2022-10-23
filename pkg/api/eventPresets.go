package api

import (
	"context"
	"errors"
	"eventloop/pkg/eventloop/event"
	"fmt"
	"strconv"
	"time"
)

type (
	EventFunc func() func(ctx context.Context) string
	EventType uint8
)

const (
	REGULAR EventType = iota + 1
	INTERVALED
)

var events = [...]EventFunc{event1, event2}

func CreateEvent(id int, eventType EventType) (event.Interface, error) {
	switch eventType {
	case REGULAR:
		return event.NewEvent(events[id-1]()), nil
	case INTERVALED:
		return event.NewIntervalEvent(events[id-1](), 500*time.Millisecond), nil
	default:
		return nil, errors.New(fmt.Sprintf("No such type: %v", eventType))
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

func init() {
	//events = []ApiEvents{{eventFunc: event1}}
}
