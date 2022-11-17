package eventloop

import (
	"context"
	"errors"
	"eventloop/internal/logger"
	"eventloop/pkg/channelEx"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal"
	"eventloop/pkg/eventloop/internal/eventsList"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"sync"
	"time"
)

// INTERVALED - зарезервированное название события для интервальных событий
const INTERVALED = "@INTERVALED"

// eventLoop представляет собой менеджер событий. Позволяет использовать как классические события с названиями для
// каждого, так и одноразовые, выполняющиеся с определённым интервалом. Также можно задавать приоритет обычным событиям.
// Для использования нужно создавать event.
type eventLoop struct {
	events eventsList.Interface
	mx     *sync.RWMutex

	disabled           []EventFunction
	isSchedulerRunning bool
	stopScheduler      chan bool

	logger logger.Interface
}

// NewEventLoop - конструктор для менеджера событий.
// Для level рекомендуются DebugLevel для Dev, и ErrorLevel для Prod. Можно указать любой уровень, он нормализуется в
// Debug и Error, в зависимости от велчины уровня.
func NewEventLoop(level string) Interface {
	elLogger, err := logger.NewLogger(level, "logs", "")
	if err != nil {
		fmt.Printf("logger init error: %v", err)
	}

	return &eventLoop{
		mx:            &sync.RWMutex{},
		events:        eventsList.New(),
		stopScheduler: make(chan bool),
		logger:        elLogger,
	}
}

// Subscribe подписывает список событий listeners на список событий triggers. Само событие триггерится с помощью Trigger/
// В случае передачи контекста с дедлайном или таймаутом, если контекст ещё живой, подписанные события всё равно
// выполнятся один раз в случае триггера.
func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error {
	if isContextDone(ctx) {
		errStr := "can't subscribe, context is done"
		e.logger.Warnw(errStr,
			"triggers", triggers,
			"listeners", listeners)
		return errors.New(errStr)
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
	return nil
}

func isContextDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (e *eventLoop) addEvent(eventName string, newEvent event.Interface) {
	e.events.EventName(eventName).Priority(newEvent.GetPriority()).AddEvent(newEvent)
}

// On вешает newEvent на событие eventName. Событие затем срабатывает по вызову Trigger с тем же eventName,
// с которым было добавлено.
// out возвращает UUID события в хранилище событий
func (e *eventLoop) On(ctx context.Context, eventName string, newEvent event.Interface, out chan<- uuid.UUID) error {
	if isContextDone(ctx) {
		errStr := "can't add listener to event, context is done"
		e.logger.Warnw(errStr,
			"event", newEvent.GetId(),
			"eventname", eventName)
		return errors.New(errStr)
	}

	//Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- newEvent.GetId()
		}
		errStr := "can't attach listener, On disabled"
		e.logger.Warnw(errStr,
			"event", newEvent.GetId(),
			"eventname", eventName)
		return errors.New(errStr)
	}

	if eventName == INTERVALED {
		errStr := fmt.Sprintf("Event name %v is reserved", INTERVALED)
		e.logger.Warnf("Event name %v is reserved", INTERVALED)
		return errors.New(errStr)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.addEvent(eventName, newEvent)

	e.logger.Debugw("Event added", "eventname", eventName)

	if out != nil {
		out <- newEvent.GetId()
	}
	return nil
}

// Trigger вызывает событие с определённым eventName. Функция ждёт выполнения всех добавленных на событие функций,
// поэтому синхронный вызов заблокирует родительский цикл выполнения программы.
// В Ch пишется резульат выполнения каждого триггера, после использования канал закрывается. Поэтому для каждого вызова
// нужно создавать новый channelEx
func (e *eventLoop) Trigger(ctx context.Context, eventName string, ch channelEx.Interface[string]) error {

	var deferErr error

	if ch != nil && !ch.IsClosed() {
		if ch.IsClosed() {
			deferErr = errors.New("trigger accept only new channels")
		} else {
			defer func(ch channelEx.Interface[string]) {
				err := ch.Close()
				if err != nil {
					e.logger.Warn(err)
				}
			}(ch)
		}

	}

	if isContextDone(ctx) {
		str := "can't trigger event, context is done"
		e.logger.Warnw(str,
			"eventname", eventName)
		return errors.New(str)
	}
	if slices.Contains(e.disabled, TRIGGER) {
		str := "can't trigger event, trigger is disabled"
		e.logger.Warnw(str,
			"eventname", eventName)
		return errors.New(str)
	}

	if eventName == INTERVALED {
		str := fmt.Sprintf("Event name %v is reserved", INTERVALED)
		e.logger.Warnf(str)
		return errors.New(str)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("Events triggered", "eventname", eventName)

	var (
		wg sync.WaitGroup
	)

	for priorIndex := e.events.EventName(eventName).Len() - 1; priorIndex >= 0; priorIndex-- {
		for _, loopevent := range e.events.EventName(eventName).Priority(priorIndex).List() {
			wg.Add(1)

			go func(ev event.Interface) {
				defer wg.Done()

				result := ev.RunFunction(ctx)
				if ch != nil && !ch.IsClosed() {
					ch.Channel() <- result
				}

				listener := ev.GetSubscriber()
				if listener == nil {
					return
				}
				if listenerChannels := listener.GetChannels(); len(listenerChannels) > 0 {
					evTrigger := ev.GetSubscriber()
					evTrigger.LockMutex()
					e.logger.Debugw("Sending messages...",
						"listener", listenerChannels,
						"trigger", ev.GetId())
					for _, chnl := range listenerChannels {
						e.logger.Debugw("Writing to channel", "channel", ch, "trigger", ev.GetId())
						chnl <- 1
					}
					e.logger.Infow("All messages send", "trigger", ev.GetId())
					evTrigger.UnlockMutex()
				}
			}(loopevent)

			if loopevent.IsOnce() {
				loopevent.GetOnce().Do(func() {
					e.RemoveEvent(loopevent.GetId())
				})
			}
		}
	}
	wg.Wait()
	return deferErr
}

// Toggle выключает функции менеджера событий, ON и TRIGGER. При попытке использования этих функций выводится ошибка.
// Функции можно включить обратно простым прокидыванием тех же параметров, в зависимости от того что надо включить.
func (e *eventLoop) Toggle(eventFuncs ...EventFunction) {
	for _, v := range eventFuncs {
		//Включение
		if x := slices.Index(e.disabled, v); x != -1 {
			e.logger.Infof("Enabling function %v", v)
			e.disabled = internal.RemoveSliceItemByIndex(e.disabled, x)
		} else { //Выключение
			e.logger.Infof("Disabling function %v", v)
			e.disabled = append(e.disabled, v)
		}
	}
}

// isScheduledEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - всё, гг (либо канал самого ивента, канал ивентлупа и context.Done()
func isScheduledEventDone(eventCh, eventLoopCh <-chan bool, ctx context.Context, logger logger.Interface) <-chan struct{} {
	result := make(chan struct{}, 1)
	result <- struct{}{}
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

// ScheduleEvent добавляет событие в список интервальных событий. Если шедулер уже запущен до добавления - запускает
// это событие сразу.
// out возвращает UUID события в хранилище событий
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID) error {

	if _, err := newEvent.GetSchedule(); err != nil {
		e.logger.Errorw(err.Error(), "event", newEvent)
		return err
	}

	if isContextDone(ctx) {
		errStr := "can't schedule, context done"
		e.logger.Warnw(errStr)
		return errors.New(errStr)
	}

	e.addEvent(INTERVALED, newEvent)

	if e.isSchedulerRunning {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- newEvent.GetId()
	}
	return nil
}

// StartScheduler запускает выполнение ранее добавленных интервальных событий. Если вызвать ScheduleEvent после
// StartScheduler (т.е. планировщик уже запущен), свежедобавленное событие начнёт выполняться сразу.
func (e *eventLoop) StartScheduler(ctx context.Context) error {
	if isContextDone(ctx) {
		errStr := "scheduler can't start, context is done"
		e.logger.Warnw(errStr)
		return errors.New(errStr)
	}

	if e.isSchedulerRunning {
		errStr := "scheduler is already running"
		e.logger.Warnw(errStr)
		return errors.New(errStr)
	}

	for _, evts := range e.events.EventName(INTERVALED).Priority(0).List() {
		curEvts := evts
		go e.runScheduledEvent(ctx, curEvts)
	}

	e.isSchedulerRunning = true
	e.logger.Infow("Scheduler started")
	return nil
}

// StopScheduler останавливает выполнение ранее добавленных интервальных событий.
func (e *eventLoop) StopScheduler() {
	e.logger.Infow("Scheduler stopping...")
	e.mx.Lock()
	defer e.mx.Unlock()
	if len(e.events.EventName(INTERVALED).Priority(0).List()) > 0 && e.isSchedulerRunning {
		e.logger.Infow("Send signal to stop")
		e.stopScheduler <- true
	}
	e.isSchedulerRunning = false
	e.logger.Infow("Send signal to stop")
}

// RemoveEvent удаляет событие. Возвращает true если событие было в хранилище, false если не было
func (e *eventLoop) RemoveEvent(id uuid.UUID) bool {
	return e.events.RemoveEvent(id)
}

// GetAttachedEvents возвращает все события, прикреплённые к eventName
func (e *eventLoop) GetAttachedEvents(eventName string) (result []uuid.UUID, err error) {
	return e.events.GetEventIdsByName(eventName)
}
