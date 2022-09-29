package main

import (
	"Biba/helpers"
	"fmt"
	"golang.org/x/exp/slices"
	"os"
	"sort"
	"sync"
	"time"
)

type EventFunc string

const (
	ON      EventFunc = "ON"
	TRIGGER EventFunc = "TRIGGER"
)

type EventLoop interface {
	On(eventName string, newEvent event)
	Trigger(eventName string)
	Toggle(eventFunc ...EventFunc)
	ScheduleEvent(event eventSchedule)
	StartScheduler()
	StopScheduler()
}

//type Event interface {
//	GetName() string
//}

type eventLoop struct {
	//events []event
	events             map[string][]event
	intervalEvents     []eventSchedule
	mx                 *sync.RWMutex
	disabled           []EventFunc
	quitScheduler      chan bool
	isSchedulerRunning bool
}

type event struct {
	//name string
	priority int
	fun      func()
	isOnce   bool
}

type eventSchedule struct {
	base     event
	interval int
}

//func (e *event) GetName() string {
//	return e.name
//}

func (e *eventLoop) On(eventName string, newEvent event) {
	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		return
	}
	e.mx.Lock()
	defer e.mx.Unlock()

	e.events[eventName] = append(e.events[eventName], newEvent)
	if newEvent.priority > 0 {
		sort.Slice(e.events[eventName], func(i, j int) bool {
			return e.events[eventName][i].priority > e.events[eventName][j].priority
		})
	}
}

func (e *eventLoop) Trigger(eventName string) {
	if slices.Contains(e.disabled, TRIGGER) {
		return
	}

	e.mx.RLock()
	defer e.mx.RUnlock()
	//for _, ev := range e.events {
	//	if ev.GetName() == eventName {
	//		go ev.fun()
	//	}
	//}

	//fmt.Print(e.events[eventName])
	for i := len(e.events[eventName]) - 1; i >= 0; i-- {
		curEvent := e.events[eventName][i]
		go func(ev event) {
			ev.fun()
		}(curEvent)

		if curEvent.isOnce {
			e.events[eventName] = helpers.RemoveIndex(e.events[eventName], i)
		}
	}
}

func (e *eventLoop) Toggle(eventFuncs ...EventFunc) {
	for _, v := range eventFuncs {
		if x := slices.Index(e.disabled, v); x != -1 {
			helpers.RemoveIndex(e.disabled, x)
		} else {
			e.disabled = append(e.disabled, v)
		}
	}
}

func (e *eventLoop) runScheduledEvent(event eventSchedule) {
	fmt.Println("Scheduled event started")
	ticker := time.NewTicker(time.Duration(event.interval) * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			event.base.fun()
		case <-e.quitScheduler:
			ticker.Stop()
			fmt.Println("Scheduled event stopped")
			return
		}
	}
}

func (e *eventLoop) ScheduleEvent(newEvent eventSchedule) {
	e.intervalEvents = append(e.intervalEvents, newEvent)
	if e.isSchedulerRunning {
		e.runScheduledEvent(newEvent)
	}
}

func (e *eventLoop) StartScheduler() {
	for _, evts := range e.intervalEvents {
		go e.runScheduledEvent(evts)
	}
	e.isSchedulerRunning = true
	fmt.Println("Scheduler started")
}

func (e *eventLoop) StopScheduler() {
	fmt.Println("Scheduler stopping...")
	e.quitScheduler <- true
	fmt.Println("Scheduler stopped...")
	e.isSchedulerRunning = false

}

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal)

	var evLoop EventLoop = &eventLoop{
		//events: make([]event, 0),
		events:         make(map[string][]event, 0),
		mx:             &sync.RWMutex{},
		disabled:       []EventFunc{},
		intervalEvents: make([]eventSchedule, 0),
		quitScheduler:  make(chan bool),
	}

	eventDefault := event{fun: func() {
		fmt.Printf("%s\n", "lol")
	}}
	go evLoop.On("keke", eventDefault)

	eventPriority := event{fun: func() {
		fmt.Printf("%s\n", "lol2")
	}, priority: 1}
	go evLoop.On("keke", eventPriority)

	eventSingle := event{fun: func() {
		fmt.Printf("%s\n", "Lol single")
	}, isOnce: true}
	go evLoop.On("keke", eventSingle)

	time.Sleep(50)
	go evLoop.Trigger("keke")

	time.Sleep(100)
	go evLoop.Trigger("keke")

	time.Sleep(150)
	go evLoop.Toggle(TRIGGER)
	go evLoop.Trigger("keke")
	time.Sleep(200)
	go evLoop.Trigger("keke")

	intervalEvent := eventSchedule{interval: 500, base: event{fun: func() {
		i := 1
		fmt.Printf("Hi, scheduler fire count: %v\t", i)
		i++
	}}}

	go evLoop.ScheduleEvent(intervalEvent)
	go evLoop.StartScheduler()
	time.Sleep(2 * time.Second)
	go evLoop.StopScheduler()
	time.Sleep(1 * time.Second)
	go evLoop.StartScheduler()
	// create goroutine which will emulate work
	// preventing deadlock
	go func() {
		for {
			time.Sleep(100)
		}
	}()

	// easiest way to loop main forever
	<-quit
}
