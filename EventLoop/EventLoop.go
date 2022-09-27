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
	On(eventName string, eventFunc func())
	Trigger(eventName string)
	Toggle(eventFunc ...EventFunc)
}

//type Event interface {
//	GetName() string
//}

type eventLoop struct {
	//events []event
	events   map[string][]event
	mx       *sync.RWMutex
	disabled []EventFunc
}

type event struct {
	//name string
	priority int
	fun      func()
	isOnce   bool
}

//func (e *event) GetName() string {
//	return e.name
//}

func (e *eventLoop) On(eventName string, newEvent event) {
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

	fmt.Print(e.events[eventName])
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

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal)

	evLoop := eventLoop{
		//events: make([]event, 0),
		events:   make(map[string][]event, 0),
		mx:       &sync.RWMutex{},
		disabled: []EventFunc{},
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
