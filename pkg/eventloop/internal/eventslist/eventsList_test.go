package eventslist

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"eventloop/pkg/eventloop/event"
)

func TestEventIdsList_AddEvent(t *testing.T) {
	var ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
		return ""
	}, TriggerName: "TRIGGER"})
	type args struct {
		newEvent event.Interface
	}
	tests := []struct {
		name string
		eil  EventsByUUIDString
		args args
	}{
		{
			name: "Default",
			eil:  make(EventsByUUIDString),
			args: args{ev},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.eil.AddEvent(tt.args.newEvent)
			if newEvt := tt.eil.EventID(tt.args.newEvent.GetUUID()); newEvt.GetUUID() != tt.
				args.newEvent.GetUUID() {
				t.Errorf("AddEvent() = %v, wantResult %v", newEvt.GetUUID(), tt.
					args.newEvent.GetUUID())
			}
		})
	}
}

func TestEventIdsList_EventID(t *testing.T) {
	var ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
		return ""
	}, TriggerName: "TRIGGER"})
	type args struct {
		eventID string
	}
	tests := []struct {
		name string
		eil  EventsByUUIDString
		args args
		want event.Interface
	}{
		{
			name: "Default",
			eil:  EventsByUUIDString{ev.GetUUID(): ev},
			args: args{ev.GetUUID()},
			want: ev,
		},
		{
			name: "No event",
			eil:  EventsByUUIDString{},
			args: args{ev.GetUUID()},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eil.EventID(tt.args.eventID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventID() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func TestEventIdsList_List(t *testing.T) {
	var (
		ev1, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, TriggerName: "TRIGGER"})
		ev2, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, TriggerName: "TRIGGER"})
	)
	tests := []struct {
		name string
		eil  EventsByUUIDString
		want EventsByUUIDString
	}{
		{
			name: "Default",
			eil:  EventsByUUIDString{ev1.GetUUID(): ev1, ev2.GetUUID(): ev2},
			want: EventsByUUIDString{ev1.GetUUID(): ev1, ev2.GetUUID(): ev2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eil.List(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("List() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func TestEventIdsList_iterateDeletionEvents(t *testing.T) {
	var (
		ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, TriggerName: "TRIGGER"})
	)
	type args struct {
		ids []string
	}
	tests := []struct {
		name       string
		eil        EventsByUUIDString
		args       args
		wantResult []string
		want       EventsByUUIDString
	}{
		{
			name:       "Empty with index",
			eil:        make(EventsByUUIDString),
			args:       args{ids: []string{ev.GetUUID()}},
			wantResult: []string{ev.GetUUID()},
			want:       make(EventsByUUIDString),
		},
		{
			name:       "Full with wrong index",
			eil:        EventsByUUIDString{"1": ev, "2": ev},
			args:       args{ids: []string{ev.GetUUID()}},
			wantResult: []string{ev.GetUUID()},
			want:       EventsByUUIDString{"1": ev, "2": ev},
		},
		{
			name:       "Single",
			eil:        EventsByUUIDString{"1": ev},
			args:       args{ids: []string{"1"}},
			wantResult: []string{},
			want:       EventsByUUIDString{},
		},
		{
			name:       "First of two",
			eil:        EventsByUUIDString{"1": ev, "2": ev},
			args:       args{ids: []string{"1"}},
			wantResult: []string{},
			want:       EventsByUUIDString{"2": ev},
		},
		{
			name:       "Last of two",
			eil:        EventsByUUIDString{"1": ev, "2": ev},
			args:       args{ids: []string{"2"}},
			wantResult: []string{},
			want:       EventsByUUIDString{"1": ev},
		},
		{
			name:       "Middle",
			eil:        EventsByUUIDString{"1": ev, "2": ev, "3": ev, "4": ev, "5": ev},
			args:       args{ids: []string{"4"}},
			wantResult: []string{},
			want:       EventsByUUIDString{"1": ev, "2": ev, "3": ev, "5": ev},
		},
		{
			name:       "Multiple",
			eil:        EventsByUUIDString{"1": ev, "2": ev, "3": ev, "4": ev},
			args:       args{ids: []string{"2", "1", "4"}},
			wantResult: []string{},
			want:       EventsByUUIDString{"3": ev},
		},
		{
			name:       "Multiple with wrong index",
			eil:        EventsByUUIDString{"1": ev, "2": ev, "3": ev, "4": ev},
			args:       args{ids: []string{"2", "1", "5"}},
			wantResult: []string{"5"},
			want:       EventsByUUIDString{"3": ev, "4": ev},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eil.iterateDeletionEvents(tt.args.ids); !reflect.DeepEqual(got,
				tt.wantResult) || !reflect.DeepEqual(tt.want, tt.eil) {
				t.Errorf("iterateDeletionEvents() = %v, wantResult %v;\nmodified container %v, "+
					"container %v",
					got,
					tt.wantResult, tt.want, tt.eil)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "Default",
			want: &eventsList{priorities: make(map[string]priorityList)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func Test_eventsList_EventName(t *testing.T) {
	var searchList = priorityList{1: make(EventsByUUIDString), 2: make(EventsByUUIDString)}
	type fields struct {
		priorities map[string]priorityList
		mx         sync.Mutex
	}
	type args struct {
		eventName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Priority
	}{
		{
			name: "With priority",
			fields: fields{priorities: map[string]priorityList{"1": make(priorityList),
				"2": searchList}},
			args: args{eventName: "2"},
			want: &searchList,
		},
		{
			name:   "No priority",
			fields: fields{priorities: make(map[string]priorityList)},
			args:   args{eventName: "BIBA"},
			want:   &priorityList{},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			el := &eventsList{
				priorities: tt.fields.priorities,
				mx:         tt.fields.mx, //nolint:govet
			}
			if got := el.EventName(tt.args.eventName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EventName() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func Test_eventsList_GetEventIdsByName(t *testing.T) {
	var (
		ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
			return ""
		}, TriggerName: "1"})
		container = map[string]priorityList{"TRIG1": {1: EventsByUUIDString{"1": ev, "2": ev}},
			"TRIG2": {0: EventsByUUIDString{"4": ev}, 2: EventsByUUIDString{"5": ev}},
			"TRIG3": {10: {}}}
	)
	type fields struct {
		priorities map[string]priorityList
		mx         sync.Mutex
	}
	type args struct {
		eventName string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name:       "Default",
			fields:     fields{priorities: container},
			args:       args{"TRIG1"},
			wantResult: []string{ev.GetUUID(), ev.GetUUID()},
			wantErr:    false,
		},
		{
			name:       "No trigger name",
			fields:     fields{priorities: container},
			args:       args{"NO TRIGGER"},
			wantResult: nil,
			wantErr:    true,
		},
		{
			name:       "Empty priority",
			fields:     fields{priorities: container},
			args:       args{"TRIG3"},
			wantResult: nil,
			wantErr:    false,
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			el := &eventsList{
				priorities: tt.fields.priorities,
				mx:         tt.fields.mx, //nolint:govet
			}
			gotResult, err := el.GetEventIdsByTriggerName(tt.args.eventName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEventIdsByTriggerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetEventIdsByTriggerName() gotResult = %v, wantResult %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_eventsList_RemoveEventByUUIDs(t *testing.T) {
	var ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
		return ""
	}, TriggerName: "TRIG"})
	type fields struct {
		priorities map[string]priorityList
		mx         sync.Mutex
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
		{
			name: "Default",
			fields: fields{priorities: map[string]priorityList{"TRIG1": {1: EventsByUUIDString{"1": ev, "2": ev}},
				"TRIG2": {0: EventsByUUIDString{"4": ev}, 2: EventsByUUIDString{"5": ev}},
				"TRIG3": {10: {}}}},
			args: args{[]string{"4", "2"}},
			want: []string{},
		},
		{
			name:   "Empty",
			fields: fields{priorities: make(map[string]priorityList)},
			args:   args{[]string{"1", "2"}},
			want:   []string{"1", "2"},
		},
		{
			name: "Partially find",
			fields: fields{priorities: map[string]priorityList{"TRIG1": {1: EventsByUUIDString{"1": ev, "2": ev}},
				"TRIG2": {0: EventsByUUIDString{"4": ev}, 2: EventsByUUIDString{"5": ev}},
				"TRIG3": {10: {}}}},
			args: args{[]string{"10", "4"}},
			want: []string{"10"},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(tt.name, func(t *testing.T) {
			el := &eventsList{
				priorities: tt.fields.priorities,
				mx:         tt.fields.mx, //nolint:govet
			}
			if got := el.RemoveEventByUUIDs(tt.args.ids...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveEventByUUIDs() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func Test_priorityList_GetKeys(t *testing.T) {
	tests := []struct {
		name     string
		pl       priorityList
		wantKeys []int
	}{
		{
			name:     "Default",
			pl:       priorityList{1: {}, 6: {}, 3: {}, -5: {}},
			wantKeys: []int{-5, 1, 3, 6},
		},
		{
			name:     "Empty",
			pl:       priorityList{},
			wantKeys: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotKeys := tt.pl.GetKeys(); !reflect.DeepEqual(gotKeys, tt.wantKeys) {
				t.Errorf("GetKeys() = %v, wantResult %v", gotKeys, tt.wantKeys)
			}
		})
	}
}

func Test_priorityList_Len(t *testing.T) {
	tests := []struct {
		name string
		pl   priorityList
		want int
	}{
		{
			name: "Default",
			pl:   priorityList{1: {}, 2: {}},
			want: 2,
		},
		{
			name: "Empty",
			pl:   priorityList{},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pl.Len(); got != tt.want {
				t.Errorf("Len() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func Test_priorityList_Priority(t *testing.T) {
	var ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
		return ""
	}, TriggerName: "TRIG1"})
	type args struct {
		priority int
	}
	tests := []struct {
		name string
		pl   priorityList
		args args
		want *EventsByUUIDString
	}{
		{
			name: "Default",
			pl:   priorityList{1: {}, 2: {"1": ev, "2": ev}},
			args: args{2},
			want: &EventsByUUIDString{"1": ev, "2": ev},
		},
		{
			name: "No priority",
			pl:   priorityList{1: {}, 2: {"1": ev, "2": ev}},
			args: args{10},
			want: &EventsByUUIDString{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pl.Priority(tt.args.priority); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Priority() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}

func Test_priorityList_iterateDeletionPriorities(t *testing.T) {
	var ev, _ = event.NewEvent(event.Args{Fun: func(ctx context.Context) string {
		return ""
	}, TriggerName: "TRIG1"})
	type args struct {
		ids []string
	}
	tests := []struct {
		name string
		pl   priorityList
		args args
		want []string
	}{
		{
			name: "Default",
			pl:   priorityList{1: {"6": ev}, 2: {"1": ev, "2": ev}},
			args: args{[]string{"6", "2"}},
			want: []string{},
		},
		{
			name: "Partially deleted",
			pl:   priorityList{1: {"6": ev}, 2: {"1": ev, "2": ev}},
			args: args{[]string{"10", "2"}},
			want: []string{"10"},
		},
		{
			name: "Fully deleted",
			pl:   priorityList{1: {"6": ev}, 2: {"1": ev, "2": ev}},
			args: args{[]string{"6", "2", "1"}},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pl.iterateDeletionPriorities(tt.args.ids); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("iterateDeletionPriorities() = %v, wantResult %v", got, tt.want)
			}
		})
	}
}
