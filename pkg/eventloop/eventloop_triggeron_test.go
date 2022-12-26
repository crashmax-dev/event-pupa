package eventloop

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"eventloop/pkg/channelEx"
	"eventloop/pkg/eventloop/event"
	"golang.org/x/sync/errgroup"
)

func TestOnAndTrigger(t *testing.T) {
	tests := []test{{name: "Simple", f: TriggerOn_Simple, want: 1},
		{name: "Multiple", f: TriggerOn_Multiple, want: 3},
		{name: "Once", f: TriggerOn_Once, want: 1},
		{name: "MultipleDefaultAndOnce", f: TriggerOn_MultipleDefaultAndOnce, want: 7},
		{name: "NoEventsTrigger", f: TriggerOn_NoEventsTrigger, want: 0},
		{name: "NoEventsTriggerWithChannel", f: TriggerOn_NoEventsTriggerWithChannel, want: 0}}

	var (
		workFunc = func(ctx context.Context) func(ctx context.Context) string {
			var number int
			return func(ctx context.Context) string {
				number++
				fmt.Printf("Current number: %d \n", number)
				return strconv.Itoa(number)
			}
		}
		ctx, _ = context.WithTimeout(context.Background(), time.Second*10)
	)

	for _, tst := range tests {
		curTest := tst
		t.Run(curTest.name, func(t *testing.T) {
			t.Parallel()
			result, _ := strconv.Atoi(curTest.f(ctx, t, curTest.name, workFunc(ctx)))

			if result != curTest.want {
				t.Errorf("test %s Number = %d; WANT %d", curTest.name, result, curTest.want)
			}
		})
	}
}

func TriggerOn_NoEventsTriggerWithChannel(ctx context.Context, t *testing.T, name string, _ func(ctx context.Context) string) string {
	ch := channelEx.NewChannel(1)
	go func() {
		err := evLoop.Trigger(ctx, name, ch)
		if err != nil {
			t.Error(err)
		}
	}()
	return "0"
}

func TriggerOn_NoEventsTrigger(ctx context.Context, t *testing.T, name string, _ func(ctx context.Context) string) string {
	err := evLoop.Trigger(ctx, name, nil)

	if err != nil {
		t.Errorf("Empty trigger failed: %v", err)
		return "1"
	}
	return "0"
}

func TriggerOn_Simple(ctx context.Context, t *testing.T, eventName string, farg func(ctx context.Context) string) string {
	var (
		eventDefault = event.NewEvent(farg)
		errG, errCtx = errgroup.WithContext(ctx)
	)

	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventDefault, nil)
	})
	time.Sleep(time.Millisecond * 20)

	ch := channelEx.NewChannel(1)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, ch)
	})

	result := <-ch.Channel()

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}

func TriggerOn_Multiple(ctx context.Context, t *testing.T, eventName string, farg func(ctx context.Context) string) string {

	var (
		eventDefault  = event.NewEvent(farg)
		eventDefault2 = event.NewEvent(farg)
		errG, errCtx  = errgroup.WithContext(ctx)
	)

	ch := channelEx.NewChannel(1)
	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventDefault, nil)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, nil)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventDefault2, nil)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, ch)
	})
	time.Sleep(time.Millisecond * 10)
	var (
		result string
		chnl   = ch.Channel()
	)
	for elem := range chnl {
		result = elem
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}

func TriggerOn_Once(ctx context.Context, _ *testing.T, eventName string, farg func(ctx context.Context) string) string {

	errG, errCtx := errgroup.WithContext(ctx)

	eventSingle := event.NewOnceEvent(farg)
	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventSingle, nil)
	})
	time.Sleep(time.Millisecond * 20)
	ch := channelEx.NewChannel(1)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, ch)
	})
	time.Sleep(time.Millisecond * 10)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, ch)
	})
	time.Sleep(time.Millisecond * 20)

	var (
		result string
		chnl   = ch.Channel()
	)
	for elem := range chnl {
		result = elem
	}
	return result
}

func TriggerOn_MultipleDefaultAndOnce(ctx context.Context, t *testing.T, eventName string, farg func(ctx context.Context) string) string {

	var (
		eventFirst   = event.NewEvent(farg)
		eventSecond  = event.NewEvent(farg)
		eventOnce    = event.NewOnceEvent(farg)
		ch           = channelEx.NewChannel(1)
		errG, errCtx = errgroup.WithContext(ctx)
	)

	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventFirst, nil)
	})
	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventSecond, nil)
	})
	errG.Go(func() error {
		return evLoop.On(errCtx, eventName, eventOnce, nil)
	})

	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, nil)
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, nil)
	})
	time.Sleep(time.Millisecond * 20)
	errG.Go(func() error {
		return evLoop.Trigger(errCtx, eventName, ch)
	})
	time.Sleep(time.Millisecond * 20)

	var (
		result string
		chnl   = ch.Channel()
	)
	for elem := range chnl {
		result = elem
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}
