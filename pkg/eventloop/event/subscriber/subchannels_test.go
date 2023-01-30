package subscriber

import (
	"reflect"
	"sync"
	"testing"
)

func TestSubChannel_GetInfoCh(t *testing.T) {
	var ch = make(chan SubChInfo)
	type fields struct {
		infoCh   chan SubChInfo
		isClosed *bool
		mx       sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   chan SubChInfo
	}{
		{
			name:   "Default",
			fields: fields{infoCh: ch},
			want:   ch,
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			sc := &SubChannel{
				infoCh:   tt.fields.infoCh,
				isClosed: tt.fields.isClosed,
				mx:       tt.fields.mx, //nolint:govet
			}
			if got := sc.GetInfoCh(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInfoCh() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubChannel_IsCLosed(t *testing.T) {
	var b = true
	type fields struct {
		infoCh   chan SubChInfo
		isClosed *bool
		mx       sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Default",
			fields: fields{isClosed: &b},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &SubChannel{
				infoCh:   tt.fields.infoCh,
				isClosed: tt.fields.isClosed,
				mx:       tt.fields.mx,
			}
			if got := sc.IsClosed(); got != tt.want {
				t.Errorf("IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubChannel_SetIsClosed(t *testing.T) {
	var b = true
	type fields struct {
		infoCh   chan SubChInfo
		isClosed *bool
		mx       sync.RWMutex
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "Default",
			fields: fields{isClosed: &b},
			want:   true,
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			sc := &SubChannel{
				infoCh:   tt.fields.infoCh,
				isClosed: tt.fields.isClosed,
				mx:       tt.fields.mx, //nolint:govet
			}
			sc.SetIsClosed()
			if got := sc.IsClosed(); got != tt.want {
				t.Errorf("IsClosed() = %v, want %v", got, tt.want)
			}
		})
	}
}
