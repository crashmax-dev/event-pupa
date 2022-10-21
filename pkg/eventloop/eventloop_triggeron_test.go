package eventloop

import (
	"context"
	"eventloop/pkg/eventloop/event"
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestOnAndTrigger(t *testing.T) {

	t.Parallel()

	tests := []Test{{name: "Simple", f: TriggerOn_Simple, want: 1},
		{name: "Multiple", f: TriggerOn_Multiple, want: 3},
		{name: "Once", f: TriggerOn_Once, want: 1},
		{name: "MultipleDefaultAndOnce", f: TriggerOn_MultipleDefaultAndOnce, want: 7},
		{name: "NoEventsTrigger", f: TriggerOn_NoEventsTrigger, want: 0},
		{name: "NoEventsTriggerWithChannel", f: TriggerOn_NoEventsTriggerWithChannel, want: 0}}

	var (
		workFunc = func(ctx context.Context) func(ctx context.Context) string {
			var number int
			return func(ctx context.Context) string {
				fmt.Printf("Current number: %d \n", number)
				number++
				return strconv.Itoa(number)
			}
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*1)
	)

	t.Run("OnTriggerGroup", func(t *testing.T) {
		for _, test := range tests {
			curTest := test
			t.Run(curTest.name, func(t *testing.T) {
				t.Parallel()
				result, _ := strconv.Atoi(curTest.f(ctx, curTest.name, workFunc(ctx)))

				if result != curTest.want {
					t.Errorf("Test %s Number = %d; WANT %d", curTest.name, result, curTest.want)
				}
			})
		}
	})

	cancel()
}

func TriggerOn_NoEventsTriggerWithChannel(ctx context.Context, name string, _ func(ctx context.Context) string) string {
	ch := Channel[string]{Ch: make(chan string, 1)}
	go evLoop.Trigger(ctx, name, &ch)
	result := <-ch.Ch
	return result
}

func TriggerOn_NoEventsTrigger(ctx context.Context, name string, _ func(ctx context.Context) string) string {
	go evLoop.Trigger(ctx, name, nil)
	return "0"
}

func TriggerOn_Simple(ctx context.Context, eventName string, farg func(ctx context.Context) string) string {
	var (
		eventDefault = event.NewEvent(farg)
	)

	go evLoop.On(ctx, eventName, eventDefault, nil)
	time.Sleep(time.Millisecond * 20)

	ch := Channel[string]{Ch: make(chan string, 1)}
	go evLoop.Trigger(ctx, eventName, &ch)
	time.Sleep(time.Millisecond * 20)

	result := <-ch.Ch
	return result
}

func TriggerOn_Multiple(ctx context.Context, eventName string, farg func(ctx context.Context) string) string {

	var (
		eventDefault  = event.NewEvent(farg)
		eventDefault2 = event.NewEvent(farg)
	)

	ch := Channel[string]{Ch: make(chan string, 1)}
	go evLoop.On(ctx, eventName, eventDefault, nil)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, eventName, nil)
	time.Sleep(time.Millisecond * 10)

	go evLoop.On(ctx, eventName, eventDefault2, nil)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, eventName, &ch)
	time.Sleep(time.Millisecond * 10)
	var result string
	for elem := range ch.Ch {
		result = elem
	}
	return result
}

func TriggerOn_Once(ctx context.Context, eventName string, farg func(ctx context.Context) string) string {

	eventSingle := event.NewOnceEvent(farg)
	go evLoop.On(ctx, eventName, eventSingle, nil)
	ch := Channel[string]{Ch: make(chan string, 1)}
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, eventName, &ch)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, eventName, &ch)
	time.Sleep(time.Millisecond * 20)

	var result string
	for elem := range ch.Ch {
		result = elem
	}
	return result
}

func TriggerOn_MultipleDefaultAndOnce(ctx context.Context, eventName string, farg func(ctx context.Context) string) string {

	var (
		eventFirst  = event.NewEvent(farg)
		eventSecond = event.NewEvent(farg)
		eventOnce   = event.NewOnceEvent(farg)
		ch          = Channel[string]{Ch: make(chan string, 1)}
	)

	go evLoop.On(ctx, eventName, eventFirst, nil)
	go evLoop.On(ctx, eventName, eventSecond, nil)
	go evLoop.On(ctx, eventName, eventOnce, nil)

	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, eventName, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, eventName, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, eventName, &ch)
	time.Sleep(time.Millisecond * 20)

	var result string
	for elem := range ch.Ch {
		result = elem
	}
	return result
}
