package eventloop

import (
	"context"
	"eventloop/event"
	"eventloop/helpers"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/exp/slices"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

//type LoggerLevel string
//
//const (
//	DEBUG LoggerLevel = iota
//	PROD
//)

type eventLoop struct {
	//events []event
	events             map[string]map[string]event.Interface
	intervalEvents     []event.Interface
	mx                 *sync.RWMutex
	disabled           []EventFunction
	isSchedulerRunning bool
	stopScheduler      chan bool

	logger *zap.SugaredLogger
}

func initLogger(level zapcore.Level) *zap.SugaredLogger {

	const LOGPATH = "logs"

	var (
		levelSelected zapcore.Level
	)

	if level == zap.DebugLevel {
		levelSelected = zapcore.DebugLevel
	} else {
		levelSelected = zapcore.ErrorLevel
	}
	atom := zap.NewAtomicLevelAt(levelSelected)

	err := os.MkdirAll(LOGPATH, os.ModePerm)
	if err != nil {
		log.Println(err)
	}

	filename := helpers.GetOSFilePath(filepath.Join("logs",
		fmt.Sprintf("log%s.log",
			//time.Now().Format("02-01-2006T150405-0700"))))
			time.Now().Format("02012006"))))

	config := zap.Config{
		Level:    atom,
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:     "time",
			MessageKey:  "message",
			LevelKey:    "level",
			NameKey:     "namekey",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime:  zapcore.ISO8601TimeEncoder},
		OutputPaths:      []string{filename},
		ErrorOutputPaths: []string{filename},
	}
	newWinFileSink := func(u *url.URL) (zap.Sink, error) {
		// Remove leading slash left by url.Parse()
		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	}

	//rawJSON := []byte(`{
	//  "level": "debug",
	//  "encoding": "json",
	//  "encoderConfig": {
	//    "messageKey": "message",
	//    "levelKey": "level",
	//    "levelEncoder": "lowercase"
	//  }
	//}`)

	zap.RegisterSink("winfile", newWinFileSink)
	logger := zap.Must(config.Build())

	defer logger.Sync()

	logger.Info("logger construction succeeded")

	return logger.Sugar()
}

func NewEventLoop(level zapcore.Level) Interface {
	//var evLoop eventloop = &eventLoop{
	//	events:         make(map[string][]*event, 0),
	//	mx:             &sync.RWMutex{},
	//	disabled:       []EventFunc{},
	//	intervalEvents: make([]*eventSchedule, 0),
	//	stopScheduler:  make(chan bool),
	//}
	return &eventLoop{
		mx:            &sync.RWMutex{},
		events:        make(map[string]map[string]event.Interface),
		stopScheduler: make(chan bool),
		logger:        initLogger(level),
	}
}

func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) {
	if isContextDone(ctx) {
		e.logger.Warnw("Can't subscribe, context is done", "triggers", triggers, "listeners", listeners)
		return
	}
	for _, v := range listeners {
		trigger := v.GetSubscriber()
		for _, t := range triggers {
			ch := make(chan int, 1)
			trigger.AddChannel(ch)
			e.logger.Infow("Event subscribed", "trigger", t.GetId(), "listener", v.GetId())
			t.GetSubscriber().AddChannel(ch)
		}

		//Запскаем ждуна, когда триггеры сработают, и срабатываем сами
		go func(ctx context.Context, v event.Interface) {
			for {
				select {
				case <-ctx.Done():
					e.logger.Infow("Stop listening because of context", "eventId", v.GetId())
					return
				default:
					trigger.LockMutex()
					//v.mx.Lock()
					e.logger.Debugw("Waiting for triggers...", "event", v.GetId())
					channels := trigger.GetChannels()
					e.logger.Debugw("Reading channels", "channels", channels, "event", v.GetId())
					for _, ch := range channels {
						e.logger.Debugw("Waiting for channel", "ch", ch, "event", v.GetId())
						<-ch
					}
					e.logger.Infow("Subscriber event fired", "event", v.GetId())
					v.RunFunction(ctx)
					trigger.UnlockMutex()
				}
			}
		}(ctx, v)
	}
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *eventLoop) On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- uuid.UUID) {
	if isContextDone(ctx) {
		e.logger.Warnw("Can't add listener to event, context is done",
			"event", newEvent.GetId(),
			"eventname", eventName)
		return
	}

	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- newEvent.GetId()
		}
		e.logger.Warnw("Can't attach listener, On disabled",
			"event", newEvent.GetId(),
			"eventname", eventName)
		return
	}
	e.mx.Lock()
	defer e.mx.Unlock()

	e.events[eventName][newEvent.GetId().String()] = newEvent
	//append(e.events[eventName], newEvent)
	if newEvent.GetPriority() > 0 {
		sort.Slice(e.events[eventName], func(i, j int) bool {
			return e.events[eventName][i].GetPriority() < e.events[eventName][j].GetPriority()
		})
	}

	e.logger.Debugw("Event added", "eventname", eventName, "eventlist", e.events[eventName])

	if out != nil {
		out <- newEvent.GetId()
	}
}

func (e *eventLoop) Trigger(ctx context.Context, eventName string, out chan<- string) {

	if isContextDone(ctx) {
		e.logger.Warnw("Can't trigger event, context is done",
			"eventname", eventName)
		return
	}
	if slices.Contains(e.disabled, TRIGGER) {
		e.logger.Warnw("Can't trigger event, trigger is disabled",
			"eventname", eventName)
		return
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("Events triggered", "eventname", eventName, "eventscount", len(e.events[eventName]))

	for i := len(e.events[eventName]) - 1; i >= 0; i-- {
		curEvent := e.events[eventName][i]

		go func(ev event.Interface) {
			result := ev.RunFunction(ctx)
			if out != nil {
				out <- result
			}

			listener := ev.GetSubscriber()
			if listener == nil {
				return
			}
			if listenerChannels := listener.GetChannels(); len(listenerChannels) > 0 {
				evTrigger := ev.GetSubscriber()
				evTrigger.LockMutex()
				//evTrigger.mx.Lock()
				e.logger.Debugw("Sending messages...", "listener", listenerChannels, "trigger", ev.GetId())
				for _, ch := range listenerChannels {
					e.logger.Debugw("Writing to channel", "channel", ch, "trigger", ev.GetId())
					ch <- 1
				}
				e.logger.Infow("All messages send", "trigger", ev.GetId())
				evTrigger.UnlockMutex()
				//ev.subscriber.mx.Unlock()
			}
		}(curEvent)

		if curEvent.IsOnce() {
			e.events[eventName] = helpers.RemoveIndex(e.events[eventName], i)
		}
	}
}

func (e *eventLoop) Toggle(eventFuncs ...EventFunction) {
	for _, v := range eventFuncs {
		if x := slices.Index(e.disabled, v); x != -1 {
			e.logger.Infow("Enabling functions", "eventfuncs", eventFuncs)
			e.disabled = helpers.RemoveIndex(e.disabled, x)
		} else {
			e.logger.Infow("Disabling functions", "eventfuncs", eventFuncs)
			e.disabled = append(e.disabled, v)
		}
	}
}

// isScheduledEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - гг (канал самого ивента, канал ивентлупа и context.Done()
func isScheduledEventDone(eventCh, eventLoopCh <-chan bool, ctx context.Context, logger *zap.SugaredLogger) <-chan struct{} {
	result := make(chan struct{}, 1)
	result <- struct{}{}
	//fmt.Println("Bobs")
	select {
	case <-ctx.Done():
		if logger != nil {
			logger.Warnw("Scheduler stopped because of context")
		}
		return result
	case <-eventCh:
		if logger != nil {
			logger.Infow("Scheduler stopped because of event want to stop")
		}
		return result
	case <-eventLoopCh:
		if logger != nil {
			logger.Infow("Scheduler stopped because of event manager commands")
		}
		return result
	default:
		return make(chan struct{})
	}
}

func (e *eventLoop) runScheduledEvent(ctx context.Context, event event.Interface) {
	evntSchedule, _ := event.GetSchedule()
	evntInterval := evntSchedule.GetInterval()
	e.logger.Infow("Scheduled event starting with interval",
		"event", event.GetId(),
		"interval", evntInterval)
	ticker := time.NewTicker(evntInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go event.RunFunction(ctx)
		case <-isScheduledEventDone(evntSchedule.GetQuitChannel(), e.stopScheduler, ctx, e.logger):
			return
		}
	}
}

// ScheduleEvent добавляет ивент в список ивентов-таймеров. Если шедулер запущен - запускает этот ивент.
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID) {

	if _, err := newEvent.GetSchedule(); err != nil {
		e.logger.Errorw(err.Error(), "event", newEvent)
		return
	}

	if isContextDone(ctx) {
		e.logger.Warnw("Can't schedule, context done")
		return
	}

	e.intervalEvents = append(e.intervalEvents, newEvent)

	if e.isSchedulerRunning {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- newEvent.GetId()
	}
}

func (e *eventLoop) StartScheduler(ctx context.Context) {
	if isContextDone(ctx) {
		e.logger.Warnw("Scheduler can't start, context is done")
		return
	}

	if e.isSchedulerRunning {
		e.logger.Warnw("Scheduler is already running")
		return
	}

	for _, evts := range e.intervalEvents {
		curEvts := evts
		go e.runScheduledEvent(ctx, curEvts)
	}

	e.isSchedulerRunning = true
	e.logger.Infow("Scheduler started")
}

func (e *eventLoop) StopScheduler() {
	e.logger.Infow("Scheduler stopping...")
	e.mx.Lock()
	defer e.mx.Unlock()
	if len(e.intervalEvents) > 0 && e.isSchedulerRunning {
		e.logger.Infow("Send signal to stop")
		e.stopScheduler <- true
	}
	e.isSchedulerRunning = false
	e.logger.Infow("Send signal to stop")
}

func (e *eventLoop) RemoveEvent(id uuid.UUID) bool {
	e.mx.Lock()
	defer e.mx.Unlock()

	for key, events := range e.events {
		for i := len(events) - 1; i >= 0; i-- {
			if events[i].GetId() == id {
				e.events[key] = helpers.RemoveIndex(e.events[key], i)
				e.logger.Infow("Event removed from regular events", "event", events[i].GetId())
				return true
			}
		}
	}

	for i := len(e.intervalEvents) - 1; i >= 0; i-- {
		if e.intervalEvents[i].GetId() == id {
			schedEvent, _ := e.intervalEvents[i].GetSchedule()
			if e.isSchedulerRunning {
				e.logger.Infow("Event is running, stopping...", "event", e.intervalEvents[i].GetId())
				schedEvent.GetQuitChannel() <- true
			}
			e.intervalEvents = helpers.RemoveIndex(e.intervalEvents, i)
			e.logger.Infow("Event removed from regular events", "event", e.intervalEvents[i].GetId())
			return true
		}
	}
	return false
}

func (e *eventLoop) LockMutex() {
	e.mx.Lock()
}

func (e *eventLoop) UnlockMutex() {
	e.mx.Unlock()
}
