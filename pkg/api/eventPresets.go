package api

import (
	"context"
	"eventloop/pkg/eventloop/event"
	"strconv"
)

type EventFunc func() event.Interface

var events = [...]EventFunc{event1, event2}

func event1() event.Interface {
	var number int
	return event.NewEvent(func(ctx context.Context) string {
		number++
		return strconv.Itoa(number)
	})

}

func event2() event.Interface {
	var number int
	return event.NewEvent(func(ctx context.Context) string {
		number--
		return strconv.Itoa(number)
	})

}

func init() {
	//events = []ApiEvents{{eventFunc: event1}}
}
