package once

import (
	"reflect"
	"sync"
	"testing"
)

func TestNewOnce(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "Default",
			want: NewOnce(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOnce(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOnce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventOnce_Do(t *testing.T) {
	type fields struct {
		Once sync.Once
	}
	type args struct {
		f func()
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Default",
			fields: fields{},
			args: args{f: func() {
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := &eventOnce{
				Once: tt.fields.Once,
			}
			ev.Do(tt.args.f)
		})
	}
}
