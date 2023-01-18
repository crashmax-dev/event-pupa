package eventloop

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	loggerImplementation "eventloop/internal/logger"
	"eventloop/pkg/eventloop/event"
	"eventloop/pkg/eventloop/internal/eventslist"
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
		events   eventslist.Interface
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
		initFunc   func(triggerName string, e Interface, wantResult []string)
		wantResult []string
		wantErr    bool
	}
	var defaultFields = fields{
		eventslist.New(),
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
			initFunc: func(triggerName string, e Interface, wantResult []string) {
				var evnt, _ = event.NewEvent(defaultEventArgsFunc(triggerName))
				e.RegisterEvent(context.Background(), evnt)
				wantResult = []string{evnt.GetUUID()}
			},
			wantErr: false,
		},
		{
			name:       "No events",
			fields:     defaultFields,
			args:       args{cTestPrefix + "NoEvents"},
			wantResult: []string{},
			wantErr:    true,
		},
		{
			name:   "Two events one trigger",
			fields: defaultFields,
			args:   args{cTestPrefix + "2e1t"},
			initFunc: func(triggerName string, e Interface, t *testInput) {
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
				t.wantResult = []string{evnt1.GetUUID(), evnt2.GetUUID()}
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
				tt.initFunc(tt.args.triggerName, e, &tt)
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
		lgger, _ = loggerImplementation.NewLogger("Debug",
			"test", "test")
		ctxCamcelled, _ = context.WithDeadline(context.Background(), time.Time{})
		ev, _           = event.NewEvent(event.Args{TriggerName: "TRIGGER", Fun: func(ctx context.Context) string {
			return ""
		}})
	)
	type fields struct {
		events   eventslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx      context.Context
		newEvent event.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "Cancelled context",
			fields:  fields{events: eventslist.New(), logger: lgger},
			args:    args{ctx: ctxCamcelled, newEvent: ev},
			wantErr: true,
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
			if err := e.RegisterEvent(tt.args.ctx, tt.args.newEvent); (err != nil) != tt.wantErr {
				t.Errorf("RegisterEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_RemoveEventByUUIDs(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ids []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if got := e.RemoveEventByUUIDs(tt.args.ids...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveEventByUUIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventLoop_Subscribe(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
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
		// TODO: Add test cases.
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
	type fields struct {
		events   eventslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
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

func Test_eventLoop_Toggle(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		eventFuncs []EventFunction
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if gotResult := e.Toggle(tt.args.eventFuncs...); gotResult != tt.wantResult {
				t.Errorf("Toggle() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_eventLoop_Trigger(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
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
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			if err := e.Trigger(tt.args.ctx, tt.args.triggerName); (err != nil) != tt.wantErr {
				t.Errorf("Trigger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventLoop_addEvent(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
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
		// TODO: Add test cases.
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
	type fields struct {
		events   eventslist.Interface
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
		// TODO: Add test cases.
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
	type fields struct {
		events   eventslist.Interface
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
		// TODO: Add test cases.
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
	type fields struct {
		events   eventslist.Interface
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.runnerListener(tt.args.ctx, tt.args.v)
		})
	}
}

func Test_eventLoop_runnerTrigger(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
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
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.runnerTrigger(tt.args.ctx, tt.args.v)
		})
	}
}

func Test_eventLoop_triggerEventFunc(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.triggerEventFunc(tt.args.ctx, tt.args.ev)
		})
	}
}

func Test_eventLoop_triggerEventFuncList(t *testing.T) {
	type fields struct {
		events   eventslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		ctx  context.Context
		list eventslist.EventsByUUIDString
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventLoop{
				events:   tt.fields.events,
				mx:       tt.fields.mx,
				disabled: tt.fields.disabled,
				logger:   tt.fields.logger,
			}
			e.triggerEventFuncList(tt.args.ctx, tt.args.list)
		})
	}
}

func Test_isContextDone(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
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
	type args[T any] struct {
		ctx     context.Context
		eventCh <-chan T
		logger  logger.Interface
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want <-chan struct{}
	}
	tests := []testCase[ /* TODO: Insert concrete types here */ any]{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEventDone(tt.args.ctx, tt.args.eventCh, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isEventDone() = %v, want %v", got, tt.want)
			}
		})
	}
}
