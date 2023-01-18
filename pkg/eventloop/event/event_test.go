package event

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"eventloop/internal/logger"
	"eventloop/pkg/eventloop/event/after"
	"eventloop/pkg/eventloop/event/interval"
	"eventloop/pkg/eventloop/event/once"
	"eventloop/pkg/eventloop/event/subscriber"
	"eventloop/pkg/eventloop/internal"
	"github.com/google/uuid"
)

var testData = struct {
	TRIGGER string
	Daa     after.DateAfterArgs
	F       func(ctx context.Context) string
}{TRIGGER: "TRIGGER",
	Daa: after.DateAfterArgs{Date: time.Now(), IsRelative: true},
	F: func(ctx context.Context) string {
		return ""
	},
}

func nullable[T any](in reflect.Value) T {
	if in.IsNil() {
		var null T
		return null
	}
	return in.Interface().(T)
}

func deepEqual(one Interface, two Interface) bool {
	type returnable struct {
		i   any
		err error
	}

	var isSame = func(f1 any, f2 any, name string) bool {
		tmpResult := reflect.ValueOf(f1).Call([]reflect.Value{})

		result1 := returnable{
			i:   tmpResult[0].Interface(),
			err: nullable[error](tmpResult[1]),
		}
		tmpResult = reflect.ValueOf(f2).Call([]reflect.Value{})
		result2 := returnable{
			i:   tmpResult[0].Interface(),
			err: nullable[error](tmpResult[1]),
		}
		return errors.Is(errors.Unwrap(result1.err), errors.Unwrap(result2.err))
	}

	if one.GetUUID() == two.GetUUID() &&
		one.GetTriggerName() == two.GetTriggerName() &&
		one.GetPriority() == two.GetPriority() &&
		isSame(one.After, two.After, "after") &&
		isSame(one.Subscriber, two.Subscriber, "subscriber") &&
		isSame(one.Interval, two.Interval, "interval") &&
		isSame(one.Once, two.Once, "once") {
		return true
	}
	return false
}

func TestNewEvent(t *testing.T) {

	tests := []struct {
		name    string
		args    Args
		want    func(id string) Interface
		isWant  bool
		wantErr bool
	}{
		{
			name:    "No run function",
			args:    Args{},
			wantErr: true,
		},
		{
			name:    "No type",
			args:    Args{Fun: testData.F},
			wantErr: true,
		},
		{
			name: "Trigger",
			args: Args{Fun: testData.F,
				TriggerName: testData.TRIGGER},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F, triggerName: testData.TRIGGER}
			},
		},
		{
			name: "Listener",
			args: Args{Fun: testData.F, Subscriber: subscriber.Listener},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F,
					subscriber: subscriber.NewSubscriberEvent()}
			},
		},
		{
			name: "Trigger+Once",
			args: Args{Fun: testData.F,
				TriggerName: testData.TRIGGER,
				IsOnce:      true},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F, triggerName: testData.TRIGGER, once: once.NewOnce()}
			},
		},
		{
			name: "Trigger+Once+Interval",
			args: Args{Fun: testData.F,
				TriggerName:  testData.TRIGGER,
				IsOnce:       true,
				IntervalTime: time.Minute},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F, triggerName: testData.TRIGGER, once: once.NewOnce(),
					interval: interval.NewIntervalEvent(time.Minute)}
			},
		},
		{
			name: "Trigger+Once+Interval+After",
			args: Args{Fun: testData.F,
				TriggerName:  testData.TRIGGER,
				IsOnce:       true,
				IntervalTime: time.Minute,
				DateAfter:    testData.Daa},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F, triggerName: testData.TRIGGER, once: once.NewOnce(),
					interval: interval.NewIntervalEvent(time.Minute), after: after.New(testData.Daa)}
			},
		},
		{
			name: "Trigger+Once+Interval+After+Subscriber",
			args: Args{Fun: testData.F,
				TriggerName:  testData.TRIGGER,
				IsOnce:       true,
				IntervalTime: time.Minute,
				DateAfter:    testData.Daa,
				Subscriber:   subscriber.Trigger},
			want: func(id string) Interface {
				return &event{uuid: id, fun: testData.F, triggerName: testData.TRIGGER, once: once.NewOnce(),
					interval: interval.NewIntervalEvent(time.Minute), after: after.New(testData.Daa),
					subscriber: subscriber.NewTriggerEvent()}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEvent(tt.args)
			if tt.wantErr == true && err != nil {
				return
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("NewEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if want := tt.want(got.GetUUID()); deepEqual(got, want) == false {
				t.Errorf("NewEvent() got = %v,\n want %v", got, want)
			}
		})
	}
}

func Test_event_After(t *testing.T) {
	var (
		id  = uuid.NewString()
		aft = after.New(testData.Daa)
	)

	type fields struct {
		id    string
		fun   Func
		after after.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		want    after.Interface
		wantErr bool
	}{
		{
			name:   "Default",
			fields: fields{id: id, fun: testData.F, after: aft},
			want:   aft,
		},
		{
			name:    "No after",
			fields:  fields{id: id, fun: testData.F},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				uuid:  tt.fields.id,
				fun:   tt.fields.fun,
				after: tt.fields.after,
			}
			got, err := ev.After()
			if (err != nil) != tt.wantErr {
				t.Errorf("After() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("After() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_GetID(t *testing.T) {
	var id = uuid.NewString()

	type fields struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"Default",
			fields{id},
			id,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				uuid: tt.fields.id,
			}
			if got := ev.GetUUID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_GetPriority(t *testing.T) {
	type fields struct {
		priority int
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{
			name:   "Default",
			fields: fields{priority: 1},
			want:   1,
		},
		{
			name:   "No priority",
			fields: fields{},
			want:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				priority: tt.fields.priority,
			}
			if got := ev.GetPriority(); got != tt.want {
				t.Errorf("GetPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_GetTriggerName(t *testing.T) {
	type fields struct {
		triggerName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "Default",
			fields: fields{testData.TRIGGER},
			want:   testData.TRIGGER,
		},
		{
			name:   "No trigger name",
			fields: fields{},
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				triggerName: tt.fields.triggerName,
			}
			if got := ev.GetTriggerName(); got != tt.want {
				t.Errorf("GetTriggerName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_GetTypes(t *testing.T) {
	type fields struct {
		triggerName string
		subscriber  subscriber.Interface
		interval    interval.Interface
		once        once.Interface
		after       after.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		wantOut []Type
	}{
		{
			name: "All types",
			fields: fields{
				triggerName: testData.TRIGGER,
				subscriber:  subscriber.NewSubscriberEvent(),
				interval:    interval.NewIntervalEvent(time.Second),
				once:        once.NewOnce(),
				after:       after.New(after.DateAfterArgs{Date: time.Now()}),
			},
			wantOut: []Type{TRIGGER, ONCE, AFTER, INTERVAL, SUBSCRIBER},
		},
		{
			name:    "No types",
			fields:  fields{},
			wantOut: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				triggerName: tt.fields.triggerName,
				subscriber:  tt.fields.subscriber,
				interval:    tt.fields.interval,
				once:        tt.fields.once,
				after:       tt.fields.after,
			}
			if gotOut := ev.GetTypes(); !reflect.DeepEqual(gotOut, tt.wantOut) {
				t.Errorf("GetTypes() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func Test_event_Interval(t *testing.T) {
	var newInterval = interval.NewIntervalEvent(time.Minute)
	type fields struct {
		interval interval.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		want    interval.Interface
		wantErr bool
	}{
		{
			name: "With interval",
			fields: fields{
				newInterval,
			},
			want: newInterval,
		},
		{
			name:    "No interval",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				interval: tt.fields.interval,
			}
			got, err := ev.Interval()
			if (err != nil) != tt.wantErr {
				t.Errorf("Interval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Interval() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_Once(t *testing.T) {
	type fields struct {
		once once.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		want    once.Interface
		wantErr bool
	}{
		{
			name:   "With once",
			fields: fields{once.NewOnce()},
			want:   once.NewOnce(),
		},
		{
			name:    "No once",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				once: tt.fields.once,
			}
			got, err := ev.Once()
			if (err != nil) != tt.wantErr {
				t.Errorf("Once() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Once() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_event_RunFunction(t *testing.T) {
	var (
		lgger, _ = logger.NewLogger("DEBUG", "logs", "test")
		ctx      = context.WithValue(context.Background(), internal.LOGGER_CTX_KEY, lgger)
	)
	type fields struct {
		fun        Func
		subscriber subscriber.Interface
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		needHelper bool
	}{
		{
			name: "With function",
			fields: fields{fun: func(ctx context.Context) string {
				return "OK"
			}},
			args: args{ctx},
		},
		{
			name: "Subscriber",
			fields: fields{
				fun: func(ctx context.Context) string {
					return "OK"
				},
				subscriber: subscriber.NewSubscriberEvent(),
			},
			args:       args{ctx},
			needHelper: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				fun:        tt.fields.fun,
				subscriber: tt.fields.subscriber,
			}
			if tt.needHelper {
				go func() {
					sub, _ := ev.Subscriber()
					<-sub.ChanTrigger()
				}()
			}
			ev.RunFunction(tt.args.ctx)
		})
	}
}

func Test_event_Subscriber(t *testing.T) {
	var sub = subscriber.NewSubscriberEvent()
	type fields struct {
		subscriber subscriber.Interface
	}
	tests := []struct {
		name    string
		fields  fields
		want    subscriber.Interface
		wantErr bool
	}{
		{
			name:   "With subscriber",
			fields: fields{subscriber: sub},
			want:   sub,
		},
		{
			name:    "No subcriber",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &event{
				subscriber: tt.fields.subscriber,
			}
			got, err := ev.Subscriber()
			if (err != nil) != tt.wantErr {
				t.Errorf("Subscriber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Subscriber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSubInterface(t *testing.T) {
	var (
		evnt = &event{triggerName: testData.TRIGGER, fun: func(ctx context.Context) string {
			return ""
		}}
		err = errors.New("ERROR")
	)
	type testCase[T any] struct {
		name    string
		argI    T
		want    T
		wantErr bool
	}
	tests := []testCase[Interface]{
		{
			name: "With interface",
			argI: evnt,
			want: evnt,
		},
		{
			name:    "No interface",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errGot := getSubInterface(tt.argI, err)
			if (errGot != nil) != tt.wantErr {
				t.Errorf("getSubInterface() error = %v, wantErr %v", errGot, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSubInterface() got = %v, want %v", got, tt.want)
			}
		})
	}
}
