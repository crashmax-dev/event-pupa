package eventloop

import (
	"sync"
	"testing"

	"eventloop/internal/loggerImplementation"
	"eventloop/pkg/eventloop/internal/triggerslist"
	"eventloop/pkg/logger"
)

func newTestLogger() logger.Interface {
	lg, err := loggerImplementation.NewLogger("Debug", "test", "test")
	if err != nil {
		panic(err)
	}
	return lg
}

func Test_eventLoop_ToggleTriggers(t *testing.T) {
	type fields struct {
		events   triggerslist.Interface
		mx       *sync.RWMutex
		disabled []EventFunction
		logger   logger.Interface
	}
	type args struct {
		triggerNames []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		init   func(p triggerslist.Interface, tNames ...string)
		want   string
	}{
		{
			name: "No events",
			fields: fields{events: triggerslist.New(),
				mx:     &sync.RWMutex{},
				logger: newTestLogger()},
			args: args{triggerNames: []string{"Trigger1", "Trigger2"}},
			want: "Disabling Trigger1 | Disabling Trigger2",
		},
		{
			name: "Disabled trigger",
			fields: fields{events: triggerslist.New(),
				mx:     &sync.RWMutex{},
				logger: newTestLogger()},
			args: args{triggerNames: []string{"Trigger1", "Trigger2"}},
			init: func(p triggerslist.Interface, tNames ...string) {
				for _, v := range tNames {
					p.TriggerName(v).SetIsDisabled(true)
				}
			},
			want: "Enabling Trigger1 | Disabling Trigger2",
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
				tt.init(e.events, tt.args.triggerNames[0])
			}
			if got := e.ToggleTriggers(tt.args.triggerNames...); got != tt.want {
				t.Errorf("ToggleTriggers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventLoop_Toggle(t *testing.T) {
	type fields struct {
		events   triggerslist.Interface
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
		{
			name: "Disable and Enable",
			fields: fields{events: triggerslist.New(), mx: &sync.RWMutex{},
				disabled: []EventFunction{REGISTER}, logger: newTestLogger()},
			args:       args{eventFuncs: []EventFunction{REGISTER, TRIGGER}},
			wantResult: "Enabling REGISTER | Disabling TRIGGER",
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
			if gotResult := e.ToggleEventLoopFuncs(tt.args.eventFuncs...); gotResult != tt.wantResult {
				t.Errorf("ToggleEventLoopFunc() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
