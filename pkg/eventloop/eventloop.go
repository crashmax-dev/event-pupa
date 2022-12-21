package eventloop

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"eventloop/internal/logger"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal"
	"eventloop/pkg/eventloop/internal/eventslist"
	loggerEventLoop "eventloop/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
)

// eventLoop представляет собой менеджер событий. Позволяет использовать как классические события с названиями для
// каждого, так и одноразовые, выполняющиеся с определённым интервалом. Также можно задавать приоритет обычным событиям.
// Для использования нужно создавать event.
type eventLoop struct {
	events eventslist.Interface
	mx     *sync.RWMutex

	disabled []EventFunction

	logger loggerEventLoop.Interface
}

// NewEventLoop - конструктор для менеджера событий. Инициализирует новый Event Loop.
// Для level рекомендуются DebugLevel для Dev, и ErrorLevel для Prod. Можно указать любой уровень, он нормализуется в
// Debug и Error, в зависимости от велчины уровня.
func NewEventLoop(level string) Interface {
	elLogger, err := logger.NewLogger(level, "logs", "")
	if err != nil {
		fmt.Printf("logger init error: %v", err)
	}

	return &eventLoop{
		mx:        &sync.RWMutex{},
		events:    eventslist.New(),
		scheduler: scheduler.NewScheduler(),
		logger:    elLogger,
	}
}

func (e *eventLoop) RegisterEvent(ctx context.Context, newEvent event.Interface) error {
	if ctxErr := e.checkContext(ctx, "can't register event, context is done",
		"event", newEvent.GetID().String(),
		"trigger", newEvent.GetTriggerName()); ctxErr != nil {
		return ctxErr
	}

	// Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, REGISTER) {
		errStr := "register disabled, can't register event"
		e.logger.Warnw(errStr,
			"event", newEvent.GetID())
		return errors.New(errStr)
	}

	// ON
	if triggerName := newEvent.GetTriggerName(); triggerName != "" {
		if slices.Contains(eventLoopEvents, eventLoopSystemEvent(triggerName)) {
			errStr := fmt.Sprintf("Trigger name %v is reserved", triggerName)
			e.logger.Warnf("Trigger name %v is reserved", triggerName)
			return errors.New(errStr)
		}

		e.mx.Lock()
		defer e.mx.Unlock()

		e.addEvent(newEvent.GetTriggerName(), newEvent)

		e.logger.Debugw("Event added", "triggerName", newEvent.GetTriggerName())

	} else if _, intervalErr := newEvent.Interval(); intervalErr != nil { // INTERVAL
		e.addEvent(string(INTERVALED), newEvent)
		e.logger.Debugw("Event added", "interval", intervalComp.GetDuration())
	} else if afterComp, afterErr := newEvent.After(); afterErr == nil {
		e.addEvent(string(AFTER), newEvent)
		e.logger.Debugw("Event added", "start_time", afterComp.GetDuration())
	} else {
		return errors.New("event must be at least ON, INTERVAL or AFTER")
	}

	return nil
}

// Subscribe подписывает список событий listeners на список событий triggers. Само событие триггерится с помощью Trigger/
// В случае передачи контекста с дедлайном или таймаутом, если контекст ещё живой, подписанные события всё равно
// выполнятся один раз в случае триггера.
func (e *eventLoop) Subscribe(ctx context.Context, triggers []event.Interface, listeners []event.Interface) error {
	subCtx := context.WithValue(ctx, "logger", e.logger)

	if isContextDone(subCtx) {
		errStr := "can't subscribe, context is done"
		e.logger.Warnw(errStr,
			"triggers", triggers,
			"listeners", listeners)
		return errors.New(errStr)
	}
	for _, listener := range listeners {
		listenerSubComponent := listener.Subscriber()
		for _, t := range triggers {
			ch := make(chan int, 1)
			listenerSubComponent.AddChannel(ch)
			e.logger.Infow("Event subscribed", "listenerSubComponent", t.GetID(), "listener", listener.GetID())
			tSub := t.Subscriber()
			tSub.SetIsTrigger(true)
			tSub.AddChannel(ch)
		}

		// Запскаем ждуна, когда триггеры сработают, и срабатываем сами
		go func(ctx context.Context, v event.Interface) {
			subComponent := v.Subscriber()
			for {
				select {
				case <-ctx.Done():
					e.logger.Infow("Stop listening because of context", "eventId", v.GetID())
					return
				default:
					subComponent.LockMutex()
					channels := subComponent.GetChannels()
					for i, ch := range channels {
						logTxt := fmt.Sprintf("Reading channel %v of %v", i+1, len(channels))
						e.logger.Debugw(logTxt, "event", v.GetID())
						<-ch
					}
					e.logger.Infow("Subscriber event fired", "event", v.GetID())
					v.RunFunction(ctx)
					subComponent.UnlockMutex()
				}
			}
		}(subCtx, listener)
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
			"event", newEvent.GetID().String(),
			"eventname", eventName)
		return errors.New(errStr)
	}

	// Если выключено добавление - не добавляем
	if slices.Contains(e.disabled, ON) {
		if out != nil {
			out <- newEvent.GetID()
		}
		errStr := "can't attach listener, On disabled"
		e.logger.Warnw(errStr,
			"event", newEvent.GetID(),
			"eventname", eventName)
		return errors.New(errStr)
	}

	//TODO: переписать возврат ошипки
	if slices.Contains(eventLoopEvents, eventLoopSystemEvent(eventName)) {
		errStr := fmt.Sprintf("Event name %v is reserved", INTERVALED)
		e.logger.Warnf("Event name %v is reserved", INTERVALED)
		return errors.New(errStr)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.addEvent(eventName, newEvent)

	e.logger.Debugw("Event added", "eventname", eventName)

	if out != nil {
		out <- newEvent.GetID()
	}
	return nil
}

// TODO: godoc
func (e *eventLoop) OnBefore(ctx context.Context, eventName string, elFunction event.EventFunc) {
	// Без имени ивента = триггерится будет на все эвенты
	newEv := event.NewPriorityEvent(elFunction, BEFORE_PRIORITY)

	if eventName == "" {
		e.addEvent(string(BEFORE_TRIGGER), newEv)
	} else {
		e.addEvent(eventName, newEv)
	}
}

// TODO: godoc
func (e *eventLoop) OnAfter(ctx context.Context, eventName string, elFunction event.EventFunc) {

	// Без имени ивента = триггерится будет на все эвенты
	newEv := event.NewPriorityEvent(elFunction, AFTER_PRIORITY)

	if eventName == "" {
		e.addEvent(string(AFTER_TRIGGER), newEv)
	} else {
		e.addEvent(eventName, newEv)
	}
}

// Trigger вызывает событие с определённым eventName. Функция ждёт выполнения всех добавленных на событие функций,
// поэтому синхронный вызов заблокирует родительский цикл выполнения программы.
// В Ch пишется резульат выполнения каждого триггера, после использования канал закрывается. Поэтому для каждого вызова
// нужно создавать новый channelEx
func (e *eventLoop) Trigger(ctx context.Context, triggerName string) error {
	triggerCtx := context.WithValue(ctx, "logger", e.logger)

	var deferErr error

	// Обработка канала, закрываем по завершении
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

	// Выключен ли Триггер
	if slices.Contains(e.disabled, TRIGGER) {
		str := "can't trigger event, trigger is disabled"
		e.logger.Warnw(str,
			"eventname", eventName)
		return errors.New(str)
	}

	if slices.Contains(eventLoopEvents, eventLoopSystemEvent(eventName)) {
		str := fmt.Sprintf("Event name %v is reserved", eventLoopSystemEvent(eventName))
		e.logger.Warnf(str)
		return errors.New(str)
	}

	e.mx.Lock()
	defer e.mx.Unlock()

	e.logger.Infow("Events triggered", "eventname", eventName)

	var (
		wg sync.WaitGroup
	)

	// Run before global events
	e.triggerEventFuncList(ctx, e.events.EventName(string(BEFORE_TRIGGER)).Priority(BEFORE_PRIORITY).List())

	// Run before events
	e.triggerEventFuncList(ctx, e.events.EventName(eventName).Priority(BEFORE_PRIORITY).List())

	//TODO: приоритеты не обязательно идут по порядку, переписать чтобы учитывало все приоритеты
	for priorIndex := e.events.EventName(eventName).Len() - 1; priorIndex >= 0; priorIndex-- {
		for _, loopevent := range e.events.EventName(eventName).Priority(priorIndex).List() {
			wg.Add(1)

			go e.triggerEventFunc(triggerCtx, loopevent)

			if once, err := loopevent.Once(); err == nil {
				once.Do(func() {
					e.RemoveEventByUUIDs([]uuid.UUID{loopevent.GetID()})
				})
			}
		}
	}

	wg.Wait()

	// Run after global events
	e.triggerEventFuncList(ctx, e.events.EventName(string(AFTER_TRIGGER)).Priority(AFTER_PRIORITY).List())

	// Run after events
	e.triggerEventFuncList(ctx, e.events.EventName(eventName).Priority(AFTER_PRIORITY).List())

	return deferErr
}

func (e *eventLoop) triggerEventFuncList(ctx context.Context, list eventslist.EventIdsList) {
	for _, event := range list {
		event.RunFunction(ctx)
	}
}

func (e *eventLoop) triggerEventFunc(ctx context.Context, ev event.Interface, wg *sync.WaitGroup, ch channelEx.Interface[string]) {
	defer wg.Done()

	if after, afterErr := ev.After(); afterErr == nil {
		after.Wait()
	}

	if interval, err := ev.Interval(); err == nil {
		if interval.IsRunning() {
			interval.GetQuitChannel() <- true
		} else {
			e.runScheduledEvent(ctx, ev)
		}
	} else {
		ev.RunFunction(ctx)
	}
}

// Toggle выключает функции менеджера событий, ON и TRIGGER. При попытке использования этих функций выводится ошибка.
// Функции можно включить обратно простым прокидыванием тех же параметров, в зависимости от того что надо включить.
func (e *eventLoop) Toggle(eventFuncs ...EventFunction) (result string) {
	for _, v := range eventFuncs {
		if result != "" {
			result += " | "
		}
		// Включение
		if x := slices.Index(e.disabled, v); x != -1 {
			result += fmt.Sprintf("Enabling function %v", v)
			e.logger.Info(result)
			e.disabled = internal.RemoveSliceItemByIndex(e.disabled, x)
		} else { // Выключение
			result += fmt.Sprintf("Disabling function %v", v)
			e.logger.Infof(result)
			e.disabled = append(e.disabled, v)
		}
	}
	return
}

// isScheduledEventDone нужен для прекращения работы ивентов-интервалов.
// Чекает разные каналы, и если с любого пришёл сигнал - всё, гг (либо канал самого ивента, канал ивентлупа и context.Done()
func isScheduledEventDone(ctx context.Context, eventCh, eventLoopCh <-chan bool, logger loggerInterface.Interface) <-chan struct{} {
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
	evntSchedule, _ := event.Interval()
	evntInterval := evntSchedule.GetDuration()
	e.logger.Infow("Scheduled event starting with interval",
		"event", event.GetID(),
		"interval", evntInterval)
	ticker := time.NewTicker(evntInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			go func() {
				e.scheduler.SetSchedulerResults(event.GetID(), event.RunFunction(ctx))
			}()
		case <-isScheduledEventDone(ctx, evntSchedule.GetQuitChannel(), e.scheduler.Stopper, e.logger):
			return
		}
	}
}

// ScheduleEvent добавляет событие в список интервальных событий. Если шедулер уже запущен до добавления - запускает
// это событие сразу.
// out возвращает UUID события в хранилище событий
func (e *eventLoop) ScheduleEvent(ctx context.Context, newEvent event.Interface, out chan<- uuid.UUID) error {
	if _, err := newEvent.Interval(); err != nil {
		e.logger.Errorw(err.Error(), "event", newEvent)
		return err
	}

	if isContextDone(ctx) {
		errStr := "can't schedule, context done"
		e.logger.Warnw(errStr)
		return errors.New(errStr)
	}

	e.addEvent(string(INTERVALED), newEvent)

	if e.scheduler.IsSchedulerRunning() {
		go e.runScheduledEvent(ctx, newEvent)
	}
	if out != nil {
		out <- newEvent.GetID()
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

	if e.scheduler.IsSchedulerRunning() {
		errStr := "scheduler is already running"
		e.logger.Warnw(errStr)
		return errors.New(errStr)
	}

	e.scheduler.ResetResults()

	for _, evts := range e.events.EventName(string(INTERVALED)).Priority(0).List() {
		curEvts := evts
		go e.runScheduledEvent(ctx, curEvts)
	}

	e.scheduler.RunSchedule(true)
	e.logger.Infow("Scheduler started")
	return nil
}

func (e *eventLoop) Scheduler() scheduler.Interface {
	return &e.scheduler
}

// StopScheduler останавливает выполнение ранее добавленных интервальных событий.
func (e *eventLoop) StopScheduler() {
	e.logger.Infow("Scheduler stopping...")
	e.mx.Lock()
	defer e.mx.Unlock()
	if len(e.events.EventName(string(INTERVALED)).Priority(0).List()) > 0 && e.scheduler.IsSchedulerRunning() {
		e.logger.Infow("Send signal to stop")
		e.scheduler.Stopper <- true
	}
	e.scheduler.RunSchedule(false)
	e.logger.Infow("Send signal to stop")
}

func (e *eventLoop) RemoveEventByUUIDs(ids []uuid.UUID) []uuid.UUID {
	return e.events.RemoveEventByUUIDs(ids)
}

// GetAttachedEvents возвращает все события, прикреплённые к eventName
func (e *eventLoop) GetAttachedEvents(eventName string) (result []uuid.UUID, err error) {
	return e.events.GetEventIdsByName(eventName)
}
