package interval

import (
	"reflect"
	"testing"
	"time"
)

func TestNewIntervalEvent(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
		want     Interface
	}{
		{
			name:     "Default",
			interval: time.Second,
			want:     &eventInterval{interval: time.Second},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewIntervalEvent(tt.interval); got.IsRunning() != tt.want.IsRunning() &&
				got.GetDuration() == tt.want.GetDuration() {
				t.Errorf("NewIntervalEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSchedule_GetDuration(t *testing.T) {
	type fields struct {
		interval  time.Duration
		isRunning bool
		quit      chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{
			name:   "Default",
			fields: fields{interval: time.Second},
			want:   time.Second,
		},
		{
			name:   "No interval",
			fields: fields{},
			want:   time.Duration(0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventInterval{
				interval:  tt.fields.interval,
				isRunning: tt.fields.isRunning,
				quit:      tt.fields.quit,
			}
			if got := e.GetDuration(); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSchedule_GetQuitChannel(t *testing.T) {
	var ch = make(chan bool)
	type fields struct {
		interval  time.Duration
		isRunning bool
		quit      chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   chan bool
	}{
		{
			name:   "Default",
			fields: fields{quit: ch},
			want:   ch,
		},
		{
			name:   "No channel",
			fields: fields{},
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventInterval{
				interval:  tt.fields.interval,
				isRunning: tt.fields.isRunning,
				quit:      tt.fields.quit,
			}
			if got := e.GetQuitChannel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetQuitChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSchedule_IsRunning(t *testing.T) {
	type fields struct {
		interval  time.Duration
		isRunning bool
		quit      chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Default",
			fields: fields{},
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventInterval{
				interval:  tt.fields.interval,
				isRunning: tt.fields.isRunning,
				quit:      tt.fields.quit,
			}
			if got := e.IsRunning(); got != tt.want {
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventSchedule_SetRunning(t *testing.T) {
	type fields struct {
		interval  time.Duration
		isRunning bool
		quit      chan bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Default",
			fields: fields{},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &eventInterval{
				interval:  tt.fields.interval,
				isRunning: tt.fields.isRunning,
				quit:      tt.fields.quit,
			}
			e.SetRunning(tt.want)
			if got := e.IsRunning(); got != tt.want {
				t.Errorf("IsRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}
