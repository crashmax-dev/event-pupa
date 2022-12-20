package after

import (
	"testing"
	"time"
)

func Test_eventAfter_GetDurationSec(t *testing.T) {
	type fields struct {
		date DateAfter
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
				DateAfter{date: time.Now().Add(time.Second * 3)},
			},
			want: want1,
		},
		{
			name: "Relative",
			fields: fields{DateAfter{date: time.Date(1, 1, 1, 0, 1, 0, 0,
				time.UTC),
				isRelative: true}},
			want: want2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := eventAfter{
				date: tt.fields.date,
			}
			if got := e.GetDuration(); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}
