package after

import (
	"reflect"
	"testing"
	"time"
)

func Test_eventAfter_GetDurationSec(t *testing.T) {
	type fields struct {
		date Args
	}
	want1, _ := time.ParseDuration("3s")
	want2, _ := time.ParseDuration("1m")

	tests := []struct {
		name   string
		fields fields
		want   time.Duration
	}{
		{name: "Absolute",
			fields: fields{
				Args{Date: time.Now().Add(time.Second * 3)},
			},
			want: want1,
		},
		{
			name: "Relative",
			fields: fields{Args{Date: time.Time{}.Add(time.Minute),
				IsRelative: true}},
			want: want2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := component{
				date: tt.fields.date,
			}
			if got := e.GetDuration(); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			} else {
				t.Log(got)
			}
		})
	}
}

func TestNew(t *testing.T) {
	var currentDate = time.Now()
	tests := []struct {
		name string
		args Args
		want Interface
	}{
		{
			name: "Default",
			args: Args{Date: currentDate.UTC(), IsRelative: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args)
			tt.want = &component{date: Args{Date: currentDate, IsRelative: true},
				breakCh: got.GetBreakChannel()}
			if got.IsDone() != tt.want.IsDone() || got.GetDuration() != tt.want.GetDuration() || got.GetBreakChannel() != got.GetBreakChannel() {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventAfter_GetBreakChannel(t *testing.T) {
	var ch = make(chan bool)
	type fields struct {
		date    Args
		breakCh chan bool
		isDone  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   chan bool
	}{
		{
			name:   "Default",
			fields: fields{date: Args{Date: time.Now()}, breakCh: ch},
			want:   ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &component{
				date:    tt.fields.date,
				breakCh: tt.fields.breakCh,
				isDone:  tt.fields.isDone,
			}
			if got := e.GetBreakChannel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBreakChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventAfter_IsDone(t *testing.T) {
	type fields struct {
		date    Args
		breakCh chan bool
		isDone  bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{name: "Default",
			fields: fields{date: Args{Date: time.Now()}, breakCh: make(chan bool), isDone: true},
			want:   true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &component{
				date:    tt.fields.date,
				breakCh: tt.fields.breakCh,
				isDone:  tt.fields.isDone,
			}
			if got := e.IsDone(); got != tt.want {
				t.Errorf("IsDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventAfter_Wait(t *testing.T) {
	type fields struct {
		date    Args
		breakCh chan bool
		isDone  bool
	}
	tests := []struct {
		name      string
		fields    fields
		breakFunc func(ch chan bool)
	}{
		{
			name: "Absolute",
			fields: fields{date: Args{Date: time.Now().Add(time.Millisecond * 20)},
				breakCh: make(chan bool)},
		},
		{
			name: "Relative",
			fields: fields{date: Args{Date: time.Time{}.Add(time.Millisecond * 20), IsRelative: true},
				breakCh: make(chan bool)},
		},
		{
			name: "Break by channel",
			fields: fields{date: Args{Date: time.Now().Add(time.Second * 3)},
				breakCh: make(chan bool)},
			breakFunc: func(ch chan bool) {
				ch <- true
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &component{
				date:    tt.fields.date,
				breakCh: tt.fields.breakCh,
				isDone:  tt.fields.isDone,
			}
			if tt.breakFunc != nil {
				go tt.breakFunc(e.GetBreakChannel())
			}
			e.Wait()
			if got := e.IsDone(); !got {
				t.Errorf("IsDone() = %v, want %v", got, true)
			}
		})
	}
}
