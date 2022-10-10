package eventloop

import (
	"context"
	"eventloop/event"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

var (
	evLoop Interface
)

type Test struct {
	name string
	f    func(ctx context.Context, farg func(ctx context.Context) string) string
	want int
}

func TestOnAndTrigger(t *testing.T) {

	t.Parallel()

	tests := []Test{{name: "Simple", f: TriggerOn_Simple, want: 1},
		{name: "Multiple", f: TriggerOn_Multiple, want: 2},
		{name: "Once", f: TriggerOn_Once, want: 1},
		{name: "ToggleOn", f: TriggerOn_ToggleOn, want: 2},
		{name: "ToggleTrigger", f: TriggerOn_ToggleTrigger, want: 1},
		{name: "MultipleDefaultAndOnce", f: TriggerOn_MultipleDefaultAndOnce, want: 7}}

	const WANT = 1
	var (
		workFunc = func(ctx context.Context) func(ctx context.Context) string {
			var number int
			return func(ctx context.Context) string {
				number++
				return strconv.Itoa(number)
			}
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	)

	t.Run("OnTriggerGroup", func(t *testing.T) {
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				t.Parallel()
				curTest := test
				result, _ := strconv.Atoi(test.f(ctx, workFunc(ctx)))

				if result != curTest.want {
					t.Errorf("Test %s Number = %d; WANT %d", curTest.name, result, curTest.want)
				}
			})
		}
	})

	cancel()
}

func TriggerOn_Simple(ctx context.Context, farg func(ctx context.Context) string) string {
	var (
		eventDefault = event.NewEvent(farg)
	)

	go evLoop.On(ctx, "test7", eventDefault, nil)
	time.Sleep(time.Millisecond * 20)

	ch := make(chan string, 1)
	go evLoop.Trigger(ctx, "test7", ch)
	time.Sleep(time.Millisecond * 50)

	result := <-ch
	return result
}

func TriggerOn_Multiple(ctx context.Context, farg func(ctx context.Context) string) string {

	var (
		eventDefault  = event.NewEvent(farg)
		eventDefault2 = event.NewEvent(farg)
	)

	ch := make(chan string, 1)
	go evLoop.On(ctx, "test6", eventDefault, nil)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, "test6", ch)
	time.Sleep(time.Millisecond * 10)

	go evLoop.On(ctx, "test6", eventDefault2, nil)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, "test6", ch)
	time.Sleep(time.Millisecond * 10)

	return <-ch
}

func TriggerOn_Once(ctx context.Context, farg func(ctx context.Context) string) string {

	eventSingle := event.NewOnceEvent(farg)
	go evLoop.On(ctx, "test1", eventSingle, nil)
	ch := make(chan string, 1)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test1", ch)
	time.Sleep(time.Millisecond * 10)
	go evLoop.Trigger(ctx, "test1", ch)
	time.Sleep(time.Millisecond * 20)

	return <-ch
}

func TriggerOn_ToggleOn(ctx context.Context, farg func(ctx context.Context) string) string {

	const WANT = 2

	var (
		eventDefault = event.NewEvent(farg)
		ch           = make(chan string, 1)
	)

	go evLoop.On(ctx, "test4", eventDefault, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 20)
	go evLoop.On(ctx, "test4", eventDefault, nil)

	time.Sleep(time.Millisecond * 20)
	go evLoop.Toggle(ON)
	time.Sleep(time.Millisecond * 20)
	go evLoop.On(ctx, "test4", eventDefault, nil)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Trigger(ctx, "test4", ch)
	time.Sleep(time.Millisecond * 20)

	return <-ch
}

func TriggerOn_ToggleTrigger(ctx context.Context, farg func(ctx context.Context) string) string {

	var (
		eventDefault = event.NewEvent(farg)
		ch           = make(chan string, 1)
	)

	go evLoop.On(ctx, "test", eventDefault, nil)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Trigger(ctx, "test", ch)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Toggle(TRIGGER)
	time.Sleep(time.Millisecond * 20)

	go evLoop.Trigger(ctx, "test", ch)
	time.Sleep(time.Millisecond * 20)

	return <-ch
}

func TriggerOn_MultipleDefaultAndOnce(ctx context.Context, farg func(ctx context.Context) string) string {

	var (
		eventFirst  = event.NewEvent(farg)
		eventSecond = event.NewEvent(farg)
		eventOnce   = event.NewOnceEvent(farg)
		ch          = make(chan string, 1)
	)

	go evLoop.On(ctx, "test2", eventFirst, nil)
	go evLoop.On(ctx, "test2", eventSecond, nil)
	go evLoop.On(ctx, "test2", eventOnce, nil)

	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test2", ch)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test2", ch)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test2", ch)
	time.Sleep(time.Millisecond * 20)

	return <-ch
}

func TestStartScheduler(t *testing.T) {
	t.Parallel()

	const WANT = 4
	var (
		number int
		numInc = func(ctx context.Context) string {
			number++
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)

	evSched := event.NewIntervalEvent(numInc, time.Millisecond*20)
	go evLoop.ScheduleEvent(ctx, evSched, nil)
	time.Sleep(time.Millisecond * 10)
	go evLoop.StartScheduler(ctx)
	time.Sleep(time.Millisecond * 100)
	cancel()

	fmt.Println(WANT, number)
	if number != WANT && number != WANT+1 {
		t.Errorf("Number = %d; WANT %d or %d", number, WANT, WANT+1)
	}
}

func TestSubevent(t *testing.T) {
	t.Parallel()

	const WANT = 10
	var (
		number      int
		numIncMutex = func(ctx context.Context) string {
			evLoop.LockMutex()
			number++
			evLoop.UnlockMutex()
			return ""
		}
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)

	var (
		evListener    = event.NewEvent(numIncMutex)
		evListener2   = event.NewEvent(numIncMutex)
		eventDefault  = event.NewEvent(numIncMutex)
		eventDefault2 = event.NewEvent(numIncMutex)
		eventDefault3 = event.NewEvent(numIncMutex)
	)

	go evLoop.On(ctx, "test3", eventDefault, nil)
	go evLoop.On(ctx, "test3", eventDefault2, nil)
	go evLoop.On(ctx, "test3", eventDefault3, nil)
	time.Sleep(time.Millisecond * 20)
	go evLoop.Subscribe(ctx, []event.Interface{eventDefault, eventDefault2, eventDefault3},
		[]event.Interface{evListener, evListener2})
	time.Sleep(time.Millisecond * 20)
	go evLoop.Trigger(ctx, "test3", nil)
	go evLoop.Trigger(ctx, "test3", nil)
	time.Sleep(time.Millisecond * 20)

	if number != WANT {
		t.Errorf("Number = %d; WANT %d", number, WANT)
	}

	cancel()
}

func TestMain(m *testing.M) {
	evLoop = NewEventLoop()
	os.Exit(m.Run())
}
