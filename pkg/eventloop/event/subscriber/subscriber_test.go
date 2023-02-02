package subscriber

import (
	"reflect"
	"sync"
	"testing"

	"github.com/google/uuid"
)

func TestNewSubscriberEvent(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "Default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSubscriberEvent()
			tt.want = &component{channels: got.Channels(),
				exit:   got.Exit(),
				esType: Listener}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSubscriberEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewTriggerEvent(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "Default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTriggerEvent()
			tt.want = &component{channels: got.Channels(),
				trigger: got.ChanTrigger(),
				exit:    got.Exit(),
				esType:  Trigger}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTriggerEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSubscriber_AddChannel(t *testing.T) {
	var b = true
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	type args struct {
		eventID string
		infoCh  chan SubChInfo
		b       *bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Default",
			fields: fields{channels: make(channelsByUUIDString)},
			args:   args{eventID: uuid.NewString(), infoCh: make(chan SubChInfo), b: &b},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			ev.AddChannel(tt.args.eventID, tt.args.infoCh, tt.args.b)

			if got := ev.Channels()[tt.args.eventID]; got.GetInfoCh() != tt.args.infoCh || got.
				IsClosed() != *tt.args.b {
				t.Errorf("Added channel inconsistent")
			}
		})
	}
}

func Test_eventSubscriber_ChanTrigger(t *testing.T) {
	var trig = make(chan struct{})
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	tests := []struct {
		name   string
		fields fields
		want   chan struct{}
	}{
		{
			name:   "Default",
			fields: fields{trigger: trig},
			want:   trig,
		},
		{
			name:   "No chan",
			fields: fields{},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			if got := ev.ChanTrigger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChanTrigger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSubscriber_Channels(t *testing.T) {
	var (
		id       = uuid.NewString()
		b        = true
		channels = channelsByUUIDString{id: &SubChannel{infoCh: make(chan SubChInfo),
			isClosed: &b}}
	)
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	tests := []struct {
		name   string
		fields fields
		want   channelsByUUIDString
	}{
		{
			name:   "Default",
			fields: fields{channels: channels},
			want:   channels,
		},
		{
			name:   "No channels",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			if got := ev.Channels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSubscriber_Exit(t *testing.T) {
	var ch = make(chan struct{})
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	tests := []struct {
		name   string
		fields fields
		want   chan struct{}
	}{
		{
			name:   "Default",
			fields: fields{exit: ch},
			want:   ch,
		},
		{
			name:   "No exit",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			if got := ev.Exit(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Exit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSubscriber_GetType(t *testing.T) {
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	tests := []struct {
		name   string
		fields fields
		want   Type
	}{
		{
			name:   "Trigger",
			fields: fields{esType: Trigger},
			want:   Trigger,
		},
		{
			name:   "Listener",
			fields: fields{esType: Listener},
			want:   Listener,
		},
		{
			name:   "No type",
			fields: fields{},
			want:   "",
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			if got := ev.GetType(); got != tt.want {
				t.Errorf("GetType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSubscriber_LockMutex(t *testing.T) {
	type fields struct {
		trigger  chan struct{}
		channels channelsByUUIDString
		exit     chan struct{}
		mx       sync.Mutex
		esType   Type
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name:   "Default",
			fields: fields{},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			ev := &component{
				trigger:  tt.fields.trigger,
				channels: tt.fields.channels,
				exit:     tt.fields.exit,
				mx:       tt.fields.mx, //nolint:govet
				esType:   tt.fields.esType,
			}
			go ev.UnlockMutex()
			ev.LockMutex()
		})
	}
}
