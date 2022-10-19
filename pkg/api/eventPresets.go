package api

import (
	"context"
	"eventloop/pkg/eventloop/event"
	"strconv"
)

type EventFunction func() func() event.Interface

var events = [...]EventFunction{event1}

func event1() func() event.Interface {
	var number int
	return func() event.Interface {
		return event.NewEvent(func(ctx context.Context) string {
			number++
			return strconv.Itoa(number)
		})
	}
}

func init() {
	//events = []ApiEvents{{eventFunc: event1}}
}
