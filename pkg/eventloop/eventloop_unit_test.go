package eventloop

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"eventloop/internal/loggerImplementation"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/subscriber"
	"eventloop/pkg/eventloop/internal/triggerslist"
	"eventloop/pkg/logger"
)

func Test_eventLoop_GetAttachedEvents(t *testing.T) {
	const (
		cTestPrefix = "GetAttachedEvents"
	)
	var (
		defaultEventArgsFunc = func(triggerName string) event.Args {
			return event.Args{TriggerName: triggerName, Fun: func(ctx context.Context) string {
				return ""
			}}
		}
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
	}
	type args struct {
		triggerName string
	}
	type testInput struct {
		name       string
		fields     fields
		args       args
		initFunc   func(triggerName string, e Interface) []string
		wantResult []string
		wantErr    bool
	}
	var defaultFields = fields{
		triggerslist.New(),
		new(sync.RWMutex),
		[]EventFunction{},
	}
	tests := []testInput{
		{
			name:   "One event",
			fields: defaultFields,
			args: args{
				cTestPrefix + "OneEvent",
			},
			initFunc: func(triggerName string, e Interface) []string {
				var evnt, _ = event.NewEvent(defaultEventArgsFunc(triggerName))
				e.RegisterEvent(context.Background(), evnt)
				return []string{evnt.GetUUID()}
			},
			wantErr: false,
		},
		{
			name:       "No events",
			fields:     defaultFields,
			args:       args{cTestPrefix + "NoEvents"},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:   "Two events one trigger",
			fields: defaultFields,
			args:   args{cTestPrefix + "2e1t"},
			initFunc: func(triggerName string, e Interface) []string {
				var evnt1, _ = event.NewEvent(event.Args{TriggerName: triggerName,
					Fun: func(ctx context.Context) string {
						return ""
					}})
				var evnt2, _ = event.NewEvent(event.Args{TriggerName: triggerName,
					Fun: func(ctx context.Context) string {
						return ""
					}})
				e.RegisterEvent(context.Background(), evnt1)
				e.RegisterEvent(context.Background(), evnt2)
				return []string{evnt1.GetUUID(), evnt2.GetUUID()}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lgger, _ := loggerImplementation.NewLogger("DEBUG", "test", "")
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   lgger,
			}
			if tt.initFunc != nil {
				tt.wantResult = tt.initFunc(tt.args.triggerName, e)
			}
			gotResult, err := e.GetAttachedEvents(tt.args.triggerName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAttachedEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetAttachedEvents() gotResult = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_eventLoop_RegisterEvent(t *testing.T) {
	var (
		defaultFunc = func(ctx context.Context) string {
			return ""
		}
		lgger, _ = loggerImplementation.NewLogger("Debug",
			"test", "test")
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
		ctx             = context.Background()
		ev1, _          = event.NewEvent(event.Args{TriggerName: "TRIGGER", Fun: defaultFunc})
		ev2, _          = event.NewEvent(event.Args{TriggerName: "TRIGGER2", Fun: defaultFunc})
		evReserved, _   = event.NewEvent(event.Args{TriggerName: "@AFTER", Fun: defaultFunc})
		evMock, _       = event.NewEvent(event.Args{Fun: defaultFunc, Subscriber: subscriber.Listener})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx       context.Context
		newEvent1 event.Interface
		newEvent2 event.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Cancelled context",
			fields:  fields{events: triggerslist.New(), logger: lgger, mx: &sync.RWMutex{}},
			args:    args{ctx: ctxCancelled, newEvent1: ev1, newEvent2: ev2},
			wantErr: true,
		},
		{
			name: "Disabled register",
			fields: fields{events: triggerslist.New(), logger: lgger,
				disabled: []EventFunction{REGISTER}, mx: &sync.RWMutex{}},
			args:    args{ctx: ctx, newEvent1: ev1, newEvent2: ev2},
			wantErr: true,
		},
		{
			name:    "Reserved trigger name",
			fields:  fields{events: triggerslist.New(), logger: lgger, mx: &sync.RWMutex{}},
			args:    args{ctx: ctx, newEvent1: evReserved, newEvent2: ev1},
			wantErr: true,
		},
		{
			name:    "No event type",
			fields:  fields{events: triggerslist.New(), logger: lgger, mx: &sync.RWMutex{}},
			args:    args{ctx: ctx, newEvent1: evMock, newEvent2: ev2},
			wantErr: true,
		},
		{
			name:   "Default",
			fields: fields{events: triggerslist.New(), logger: lgger, mx: &sync.RWMutex{}},
			args:   args{ctx: ctx, newEvent1: ev1, newEvent2: ev2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if err := e.RegisterEvent(tt.args.ctx, tt.args.newEvent1, tt.args.newEvent2); (err != nil) != tt.
				wantErr {
				t.Errorf("RegisterEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_RemoveEventByUUIDs(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
		ev, _    = event.NewEvent(event.Args{TriggerName: "T", Fun: func(ctx context.Context) string {
			return ""
		}})
	)
	type fields struct {
		events triggerslist.Interface
		mx     *sync.RWMutex
		logger logger.Interface
	}
	type args struct {
		ids []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		init   func(loop Interface, p event.Interface) args
		want   []string
	}{
		{
			name:   "Default",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{},
			init: func(loop Interface, p event.Interface) args {
				loop.RegisterEvent(context.Background(), p, p)
				return args{ids: []string{p.GetUUID()}}
			},
			want: []string{},
		},
		{
			name:   "Empty",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ids: []string{}},
			want:   []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events: tt.fields.events,
				mx:     tt.fields.mx,
				logger: tt.fields.logger,
			}
			if tt.init != nil {
				tt.args = tt.init(e, ev)
			}
			if got := e.RemoveEventByUUIDs(tt.args.ids...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveEventByUUIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventLoop_Subscribe(t *testing.T) {
	var (
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
		lgger, _        = loggerImplementation.NewLogger("Debug", "test", "test")
		evSub, _        = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, Subscriber: subscriber.Listener})
		evTrig, _ = event.NewEvent(event.Args{TriggerName: "T", Fun: func(ctx context.Context) string {
			return ""
		}, Subscriber: subscriber.Trigger})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx       context.Context
		triggers  []event.Interface
		listeners []event.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Context cancelled",
			fields:  fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:    args{ctxCancelled, []event.Interface{evTrig}, []event.Interface{evSub}},
			wantErr: true,
		},
		{
			name:   "Default",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args: args{ctx: context.Background(), triggers: []event.Interface{evTrig},
				listeners: []event.Interface{evSub}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if err := e.Subscribe(tt.args.ctx, tt.args.triggers, tt.args.listeners); (err != nil) != tt.wantErr {
				t.Errorf("Subscribe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_Sync(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Default",
			fields:  fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if err := e.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_Trigger(t *testing.T) {
	var (
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
		lgger           = newTestLogger()
		ev, _           = event.NewEvent(event.Args{TriggerName: "Trig", Fun: func(ctx context.Context) string {
			return "OK"
		}, IsOnce: true})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx         context.Context
		triggerName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		init    func(el Interface, p event.Interface)
		wantErr bool
	}{
		{
			name:    "Context cancelled",
			fields:  fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:    args{ctx: ctxCancelled, triggerName: "Trig"},
			wantErr: true,
		},
		{
			name: "Trigger func disabled",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger,
				disabled: []EventFunction{TRIGGER}},
			args:    args{ctx: context.Background(), triggerName: "Trig"},
			wantErr: true,
		},
		{
			name:   "Trigger name disabled",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: context.Background(), triggerName: "Trig"},
			init: func(el Interface, p event.Interface) {
				el.ToggleTriggers("Trig")
			},
			wantErr: true,
		},
		{
			name:   "Once",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: context.Background(), triggerName: "Trig"},
			init: func(el Interface, p event.Interface) {
				el.RegisterEvent(context.Background(), p)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if tt.init != nil {
				tt.init(e, ev)
			}
			if err := e.Trigger(tt.args.ctx, tt.args.triggerName); (err != nil) != tt.wantErr {
				t.Errorf("Trigger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_addEvent(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
		ev, _    = event.NewEvent(event.Args{Priority: 322, TriggerName: "Trig", Fun: func(ctx context.Context) string {
			return ""
		}})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		triggerName string
		newEvent    event.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Default",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{triggerName: "Trigger", newEvent: ev},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.addEvent(tt.args.triggerName, tt.args.newEvent)
		})
	}
}

func Test_eventLoop_checkContext(t *testing.T) {
	var (
		lgger, _        = loggerImplementation.NewLogger("Debug", "test", "test")
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx        context.Context
		message    string
		loggerArgs []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Context Cancelled",
			fields:  fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:    args{ctx: ctxCancelled, message: "Error"},
			wantErr: true,
		},
		{
			name:    "Context OK",
			fields:  fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if err := e.checkContext(tt.args.ctx, tt.args.message, tt.args.loggerArgs...); (err != nil) != tt.wantErr {
				t.Errorf("checkContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_runScheduledEvent(t *testing.T) {
	var (
		lgger, _        = loggerImplementation.NewLogger("Debug", "test", "test")
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
		ctx             = logger.WithLogger(context.Background(), lgger)
		ev, _           = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return "OK"
		}, IntervalTime: time.Microsecond, IsOnce: true})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx context.Context
		ev  event.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Tick",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: ctx, ev: ev},
		},
		{
			name:   "Stop",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: ctxCancelled, ev: ev},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.runScheduledEvent(tt.args.ctx, tt.args.ev)
		})
	}
}

func Test_eventLoop_runnerListener(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
		initFunc = func(ctx context.Context, ev event.Interface) context.Context {
			b := false
			sub, _ := ev.Subscriber()
			sub.AddChannel("ID", make(chan subscriber.SubChInfo), &b)
			return logger.WithLogger(ctx, lgger)
		}
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
		ev1, _          = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			ch := ctx.Value("chan").(chan struct{})
			go func() {
				ch <- struct{}{}
			}()
			return ""
		}, Subscriber: subscriber.Listener})
		ev2, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, Subscriber: subscriber.Listener})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx context.Context
		v   event.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		init   func(ctx context.Context, ev event.Interface) context.Context
	}{
		{
			name:   "Default",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: context.Background(), v: ev1},
			init: func(ctx context.Context, ev event.Interface) context.Context {
				ctx = initFunc(ctx, ev)
				sub, _ := ev.Subscriber()
				go func() {
					for _, subCh := range sub.Channels() {
						subCh.GetInfoCh() <- 1
					}
				}()
				return context.WithValue(ctx, "chan", sub.Exit())
			},
		},
		{
			name:   "Exit",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: ctxCancelled, v: ev2},
			init:   initFunc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if tt.init != nil {
				tt.args.ctx = tt.init(tt.args.ctx, tt.args.v)
			}
			e.runnerListener(tt.args.ctx, tt.args.v)
		})
	}
}

func Test_eventLoop_runnerTrigger(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
		initFunc = func(ctx context.Context, ev event.Interface) context.Context {
			b := false
			sub, _ := ev.Subscriber()
			sub.AddChannel("ID", make(chan subscriber.SubChInfo), &b)
			return logger.WithLogger(ctx, lgger)
		}
		ev1, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, Subscriber: subscriber.Trigger})
		ev2, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, Subscriber: subscriber.Trigger})
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx context.Context
		v   event.Interface
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		init   func(ctx context.Context, ev event.Interface) context.Context
	}{
		{
			name: "Default",
			fields: fields{
				events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger,
			},
			args: args{ctx: context.Background(), v: ev1},
			init: func(ctx context.Context, ev event.Interface) context.Context {
				ctx = initFunc(ctx, ev)
				sub, _ := ev.Subscriber()
				go func() {
					sub.ChanTrigger() <- struct{}{}
					for _, subCh := range sub.Channels() {
						<-subCh.GetInfoCh()
					}
					sub.Exit() <- struct{}{}
				}()
				return context.WithValue(ctx, "chan", sub.Exit())
			},
		},
		{
			name:   "Exit",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{}, logger: lgger},
			args:   args{ctx: ctxCancelled, v: ev2},
			init:   initFunc,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if tt.init != nil {
				tt.args.ctx = tt.init(tt.args.ctx, tt.args.v)
			}
			e.runnerTrigger(tt.args.ctx, tt.args.v)
		})
	}
}

func Test_eventLoop_triggerEventFunc(t *testing.T) {
	var (
		lgger, _   = loggerImplementation.NewLogger("Debug", "test", "test")
		evAfter, _ = event.NewEvent(event.Args{
			Fun: func(ctx context.Context) string {
				return ""
			},
			TriggerName: "TRIG",
			DateAfter:   after.Args{Date: time.Now().Add(time.Millisecond)}})
		evInterval, _ = event.NewEvent(event.Args{
			Fun: func(ctx context.Context) string {
				return ""
			},
			TriggerName:  "TRIG",
			IntervalTime: time.Millisecond,
		})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx        context.Context
		ev         event.Interface
		isRunTwice bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "After",
			fields: fields{
				events: triggerslist.New(),
				mx:     &sync.RWMutex{},
				logger: lgger,
			},
			args: args{
				ctx:        context.Background(),
				ev:         evAfter,
				isRunTwice: true,
			},
		},
		{
			name: "Interval",
			fields: fields{
				events: triggerslist.New(),
				mx:     &sync.RWMutex{},
				logger: lgger,
			},
			args: args{
				ctx:        context.Background(),
				ev:         evInterval,
				isRunTwice: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			tt.args.ctx = logger.WithLogger(tt.args.ctx, tt.fields.logger)
			if tt.args.isRunTwice {
				go func() {
					time.Sleep(time.Millisecond)
					e.triggerEventFunc(tt.args.ctx, tt.args.ev)
				}()
			}
			e.triggerEventFunc(tt.args.ctx, tt.args.ev)
		})
	}
}

func Test_eventLoop_triggerEventFuncList(t *testing.T) {
	var (
		lgger, _ = loggerImplementation.NewLogger("Debug", "test", "test")
		ev, _    = event.NewEvent(event.Args{TriggerName: "TRIG", Fun: func(ctx context.Context) string {
			return ""
		}})
	)
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx  context.Context
		list triggerslist.EventsByUUIDString
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "Default",
			fields: fields{
				events: triggerslist.New(),
				mx:     &sync.RWMutex{},
				logger: lgger,
			},
			args: args{ctx: context.Background(), list: triggerslist.EventsByUUIDString{"1": ev,
				"2": ev}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			tt.args.ctx = logger.WithLogger(tt.args.ctx, tt.fields.logger)
			e.triggerEventFuncList(tt.args.ctx, tt.args.list)
		})
	}
}

func Test_isContextDone(t *testing.T) {
	var (
		ctxDone, _ = context.WithDeadline(context.Background(), time.Time{})
		ctx        = context.Background()
	)
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Default",
			args: args{
				ctx,
			},
			want: false,
		},
		{
			name: "Done",
			args: args{ctxDone},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isContextDone(tt.args.ctx); got != tt.want {
				t.Errorf("isContextDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isEventDone(t *testing.T) {
	var (
		lgger, _        = loggerImplementation.NewLogger("Debug", "test", "test")
		ctxCancelled, _ = context.WithDeadline(context.Background(), time.Time{})
	)
	type S struct{}
	type args[T any] struct {
		ctx     context.Context
		eventCh chan T
		logger  logger.Interface
	}
	type testCase[T any] struct {
		name      string
		args      args[T]
		writeFunc func(ch chan T)
		want      struct{}
	}
	tests := []testCase[S]{
		{
			name: "Default",
			args: args[S]{
				ctx:     context.Background(),
				eventCh: make(chan S),
				logger:  lgger,
			},
			writeFunc: func(ch chan S) {
				ch <- S{}
			},
			want: struct{}{},
		},
		{
			name: "Exit",
			args: args[S]{
				ctx:     ctxCancelled,
				eventCh: make(chan S),
				logger:  lgger,
			},
			want: struct{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := isEventDone(tt.args.ctx, tt.args.eventCh, tt.args.logger)
			if tt.writeFunc != nil {
				go tt.writeFunc(tt.args.eventCh)
			}
			if got := <-ch; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isEventDone() = %v, want %v", got, tt.want)
			}
		})
	}
}
