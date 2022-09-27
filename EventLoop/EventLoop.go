package main

import (
	"Biba/helpers"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type EventLoop interface {
	On(eventName string, eventFunc func())
	Trigger(eventName string)
}

//type Event interface {
//	GetName() string
//}

type eventLoop struct {
	//events []event
	events map[string][]event
	mx     *sync.RWMutex
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

func main() {
	// easiest way to loop main forever, in case of async code test
	quit := make(chan os.Signal)

	evLoop := eventLoop{
		//events: make([]event, 0),
		events: make(map[string][]event, 0),
		mx:     &sync.RWMutex{},
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

	go evLoop.Trigger("keke")
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
