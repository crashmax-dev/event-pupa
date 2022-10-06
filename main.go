package main

import (
	"EventManager/event"
	eventloop "EventManager/eventloop"
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())

	evLoop := eventloop.NewEventLoop()
	var eventDefault = event.NewEvent(func(ctx context.Context) {
		fmt.Printf("%s\n", "lol")
	})
	go evLoop.On(ctx, "keke", eventDefault, nil)

	eventPriority := event.NewEvent(func(ctx context.Context) {
		fmt.Printf("%s\n", "lol2")
	})
	eventPriority.SetPriority(1)
	go evLoop.On(ctx, "keke", eventPriority, nil)

	time.Sleep(50)
	go evLoop.Trigger(ctx, "keke")
	//
	time.Sleep(100)
	go evLoop.Trigger(ctx, "keke")
	//
	time.Sleep(150)
	go evLoop.Toggle(eventloop.TRIGGER)
	go evLoop.Trigger(ctx, "keke")
	time.Sleep(200)
	go evLoop.Trigger(ctx, "keke")
	//
	intervalEvent := eventSchedule{quit: make(chan bool), interval: 500 * time.Millisecond, base: event{fun: func(ctx context.Context) func(ctx context.Context) {
		i := 1
		return func(ctx context.Context) {
			fmt.Printf("Hi, scheduler fire count: %v\t", i)
			i++
		}
	}(ctx)}}
	//
	//intervalEvtCh := make(chan int)
	//go evLoop.ScheduleEvent(ctx, &intervalEvent, intervalEvtCh)
	////evId := <-intervalEvtCh
	//
	//go evLoop.StartScheduler(ctx)
	//time.Sleep(2 * time.Second)
	//
	////go evLoop.RemoveEvent(evId)
	//
	//time.Sleep(time.Second)
	////go evLoop.StopScheduler()
	//
	//go evLoop.Toggle(TRIGGER)
	//time.Sleep(time.Second)
	//fmt.Println("Triggering sub event")
	//
	//evListener := eventListener{base: event{fun: func(ctx context.Context) {
	//	fmt.Println("Interface in event srabotalo N O R M A L N O")
	//}}}
	//eventDefault.subscriber = eventTrigger{}
	//go evLoop.Subscribe(ctx, []*eventTrigger{&eventDefault.subscriber}, []*eventListener{&evListener})
	//time.Sleep(time.Second)
	//go evLoop.Trigger(ctx, "keke")
	////time.Sleep(time.Second)
	//go evLoop.Trigger(ctx, "keke")

	// easiest way to loop main forever
	for {
		select {
		case sType := <-quit:
			signal.Stop(quit)
			fmt.Printf("Program finished: %v", sType)
			evLoop.StopScheduler()
			cancel()
			//evLoop.StopScheduler()
			os.Exit(0)
			return
		default:
			time.Sleep(time.Second)
		}
	}
}
