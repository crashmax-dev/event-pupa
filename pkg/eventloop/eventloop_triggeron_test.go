package eventloop

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal"
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
		ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	)
	defer cancel()

	t.Run("Outer test", func(t *testing.T) {
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
	})
}

func TriggerOn_NoEventsTriggerWithChannel(ctx context.Context,
	t *testing.T,
	name string,
	_ func(ctx context.Context) string) string {
	err := evLoop.Trigger(ctx, name)
	if err != nil {
		t.Error(err)
	}
	return "0"
}

func TriggerOn_NoEventsTrigger(ctx context.Context,
	t *testing.T,
	name string,
	_ func(ctx context.Context) string) string {
	err := evLoop.Trigger(ctx, name)

	if err != nil {
		t.Errorf("Empty trigger failed: %v", err)
		return "1"
	}
	return "0"
}

func TriggerOn_Simple(ctx context.Context,
	t *testing.T,
	triggerName string,
	farg func(ctx context.Context) string) string {
	var (
		eventDefault, _ = event.NewEvent(event.Args{
			Fun:         farg,
			TriggerName: triggerName,
		})

		execCh       = make(chan string)
		errG, errCtx = errgroup.WithContext(ctx)
		testCtx, _   = ctxWithValueAndTimeout(errCtx, internal.EXEC_CH_CTX_KEY, execCh, time.Second)
	)

	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventDefault)
	})
	<-execCh

	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})

	result := <-execCh

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}

func TriggerOn_Multiple(ctx context.Context,
	t *testing.T,
	triggerName string,
	farg func(ctx context.Context) string) (result string) {

	var (
		eventArgs = event.Args{
			Fun:         farg,
			TriggerName: triggerName,
		}
		eventDefault, _  = event.NewEvent(eventArgs)
		eventDefault2, _ = event.NewEvent(eventArgs)
		execCh           = make(chan string)
		errG, errCtx     = errgroup.WithContext(ctx)
		testCtx, _       = ctxWithValueAndTimeout(errCtx, internal.EXEC_CH_CTX_KEY, execCh, time.Second)
	)

	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventDefault)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventDefault2)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})

	for i := 0; i < 2; i++ {
		result = <-execCh
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}

func TriggerOn_Once(ctx context.Context,
	_ *testing.T,
	triggerName string,
	farg func(ctx context.Context) string) (result string) {
	var (
		errG, errCtx = errgroup.WithContext(ctx)
		execCh       = make(chan string)
		testCtx, _   = ctxWithValueAndTimeout(errCtx, internal.EXEC_CH_CTX_KEY, execCh, time.Second)
	)

	eventSingle, _ := event.NewEvent(event.Args{
		Fun:         farg,
		TriggerName: triggerName,
		IsOnce:      true,
	})
	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventSingle)
	})
	<-execCh
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	result = <-execCh
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	<-execCh
	return result
}

func TriggerOn_MultipleDefaultAndOnce(ctx context.Context,
	t *testing.T,
	triggerName string,
	farg func(ctx context.Context) string) (result string) {
	var (
		evArgs = event.Args{Fun: farg,
			TriggerName: triggerName,
		}
		eventFirst, _  = event.NewEvent(evArgs)
		eventSecond, _ = event.NewEvent(evArgs)
		eventOnce, _   = event.NewEvent(event.Args{Fun: farg,
			TriggerName: triggerName,
			IsOnce:      true})
		execCh       = make(chan string)
		errG, errCtx = errgroup.WithContext(ctx)
		testCtx, _   = ctxWithValueAndTimeout(errCtx, internal.EXEC_CH_CTX_KEY, execCh,
			time.Second)
	)

	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventFirst)
	})
	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventSecond)
	})
	errG.Go(func() error {
		return evLoop.RegisterEvent(testCtx, eventOnce)
	})
	for i := 0; i < 3; i++ {
		<-execCh
	}

	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	for i := 0; i < 3; i++ {
		<-execCh
	}
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	for i := 0; i < 2; i++ {
		<-execCh
	}
	errG.Go(func() error {
		return evLoop.Trigger(testCtx, triggerName)
	})
	for i := 0; i < 2; i++ {
		result = <-execCh
	}

	if err := errG.Wait(); err != nil {
		t.Log(err)
	}

	return result
}
