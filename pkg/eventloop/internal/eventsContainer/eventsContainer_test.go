package eventsContainer

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"gitlab.com/YSX/eventloop/pkg/eventloop/event"
)

var (
	testDefaultEvent, _ = event.NewEvent(
		event.Args{
			Fun: func(ctx context.Context) string {
				return ""
			}, TriggerName: "TRIG1",
		},
	)
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		want Interface
	}{
		{
			name: "Default",
			want: &eventsList{
				events: make(eventsMap),
				eventsByCriteria: eventsByCriteriaName{
					"TRIGGER":  make(eventsByCriteria),
					"TYPE":     make(eventsByCriteria),
					"PRIORITY": make(eventsByCriteria),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := New(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("New() = %v, wantResult %v", got, tt.want)
				}
			},
		)
	}
}

func Test_eventsList_EventsByTrigger(t *testing.T) {
	type fields struct {
		events           eventsMap
		eventsByCriteria eventsByCriteriaName
		mx               sync.Mutex
	}
	type args struct {
		triggerName string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []event.Interface
	}{
		{
			name: "With triggername",
			fields: fields{
				events:           eventsMap{"1": testDefaultEvent},
				eventsByCriteria: eventsByCriteriaName{TRIGGER: eventsByCriteria{"TRIG1": criteriaInfo{data: eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent}}}},
			},
			args: args{triggerName: "TRIG1"},
			want: []event.Interface{testDefaultEvent},
		},
		{
			name: "No triggername",
			fields: fields{
				events:           eventsMap{"1": testDefaultEvent},
				eventsByCriteria: eventsByCriteriaName{TRIGGER: eventsByCriteria{"TRIG1": criteriaInfo{data: eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent}}}},
			},
			args: args{triggerName: "BIBA"},
			want: []event.Interface{},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(
			tt.name, func(t *testing.T) {
				el := &eventsList{
					events:           tt.fields.events,
					eventsByCriteria: tt.fields.eventsByCriteria,
					mx:               tt.fields.mx, //nolint:govet
				}
				if got := el.EventsByTrigger(tt.args.triggerName); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("EventByTrigger() = %v, wantResult %v", got, tt.want)
				}
			},
		)
	}
}

func Test_priorityList_GetAllEvents(t *testing.T) {
	var (
		ev, _ = event.NewEvent(
			event.Args{
				Fun: func(ctx context.Context) string {
					return ""
				}, TriggerName: "1",
			},
		)
	)
	type fields struct {
		events eventsMap
		mx     sync.Mutex
	}
	tests := []struct {
		name   string
		fields fields
		want   []event.Interface
	}{
		{
			name:   "Default",
			fields: fields{events: eventsMap{ev.GetUUID(): ev, testDefaultEvent.GetUUID(): testDefaultEvent}},
			want:   []event.Interface{ev, testDefaultEvent},
		},
		{
			name:   "Empty",
			fields: fields{events: eventsMap{}},
			want:   []event.Interface{},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(
			tt.name, func(t *testing.T) {
				el := &eventsList{
					events: tt.fields.events,
					mx:     tt.fields.mx, //nolint:govet
				}
				if gotResult := el.GetAll(); !reflect.DeepEqual(gotResult, tt.want) {
					t.Errorf("GetAll() got = %v, want %v", gotResult, tt.want)
				}
			},
		)
	}
}

func Test_eventsList_RemoveEventByUUIDs(t *testing.T) {
	var ev, _ = event.NewEvent(
		event.Args{
			Fun: func(ctx context.Context) string {
				return ""
			}, TriggerName: "TRIG_RANDOM", IntervalTime: time.Second,
		},
	)
	type fields struct {
		events           eventsMap
		eventsByCriteria eventsByCriteriaName
		mx               sync.Mutex
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
			name: "Partially find",
			fields: fields{
				events: eventsMap{ev.GetUUID(): ev, testDefaultEvent.GetUUID(): testDefaultEvent},
				eventsByCriteria: eventsByCriteriaName{
					PRIORITY: eventsByCriteria{
						"0": criteriaInfo{
							data: eventsMap{
								ev.GetUUID():               ev,
								testDefaultEvent.GetUUID(): testDefaultEvent,
							},
						},
					},
					TRIGGER: eventsByCriteria{
						"TRIG1":       criteriaInfo{data: eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent}},
						"TRIG_RANDOM": criteriaInfo{data: eventsMap{ev.GetUUID(): ev}},
					},
					TYPE: eventsByCriteria{
						"TRIGGER": criteriaInfo{
							data: eventsMap{
								ev.GetUUID():               ev,
								testDefaultEvent.GetUUID(): testDefaultEvent,
							},
						},
						"INTERVAL": criteriaInfo{data: eventsMap{ev.GetUUID(): ev}},
					},
				},
				mx: sync.Mutex{},
			},
			args: args{[]string{ev.GetUUID(), "456"}},
			want: []string{"456"},
		},
		{
			name: "Empty container",
			fields: fields{
				events: eventsMap{},
				eventsByCriteria: eventsByCriteriaName{
					PRIORITY: eventsByCriteria{},
					TYPE:     eventsByCriteria{},
					TRIGGER:  eventsByCriteria{},
				},
				mx: sync.Mutex{},
			},
			args: args{[]string{"1", "2"}},
			want: []string{"1", "2"},
		},
	}
	for _, tt := range tests { //nolint:govet
		t.Run(
			tt.name, func(t *testing.T) {
				el := &eventsList{
					events:           tt.fields.events,
					eventsByCriteria: tt.fields.eventsByCriteria,
					mx:               tt.fields.mx, //nolint:govet
				}
				if got := el.RemoveEventByUUIDs(tt.args.ids...); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RemoveEventByUUIDs() = %v, wantResult %v", got, tt.want)
				}
			},
		)
	}
}

func Test_eventsList_RemoveTriggers(t *testing.T) {
	var (
		ev, _ = event.NewEvent(
			event.Args{
				Fun: func(ctx context.Context) string {
					return ""
				}, TriggerName: "TRIG1",
			},
		)
	)
	type fields struct {
		events           eventsMap
		eventsByCriteria eventsByCriteriaName
		mx               sync.Mutex
	}
	type args struct {
		triggers []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "Default",
			fields: fields{
				events: eventsMap{ev.GetUUID(): ev},
				eventsByCriteria: eventsByCriteriaName{
					PRIORITY: eventsByCriteria{"0": criteriaInfo{data: eventsMap{ev.GetUUID(): ev}}},
					TYPE:     eventsByCriteria{"TRIGGER": criteriaInfo{data: eventsMap{ev.GetUUID(): ev}}},
					TRIGGER:  eventsByCriteria{"TRIG1": criteriaInfo{data: eventsMap{ev.GetUUID(): ev}}},
				},
			},
			args: args{triggers: []string{"TRIG1", "TRIG3"}},
			want: []string{"TRIG3"},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				el := &eventsList{
					events:           tt.fields.events,
					eventsByCriteria: tt.fields.eventsByCriteria,
					mx:               tt.fields.mx,
				}
				if got := el.RemoveTriggers(tt.args.triggers...); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("RemoveTriggers() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func Test_eventsList_AddEvent(t *testing.T) {
	type fields struct {
		events           eventsMap
		eventsByCriteria eventsByCriteriaName
		mx               sync.Mutex
	}
	type args struct {
		newEvent event.Interface
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantEvents   eventsMap
		wantCriteria eventsByCriteriaName
	}{
		{
			name: "Default",
			fields: fields{
				events: eventsMap{},
				eventsByCriteria: eventsByCriteriaName{
					"TRIGGER":  eventsByCriteria{},
					"TYPE":     eventsByCriteria{},
					"PRIORITY": eventsByCriteria{},
				},
				mx: sync.Mutex{},
			},
			args:       args{newEvent: testDefaultEvent},
			wantEvents: eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent},
			wantCriteria: eventsByCriteriaName{
				"TRIGGER": eventsByCriteria{
					"TRIG1": criteriaInfo{
						isEnabled: true,
						data:      eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent},
					},
				},
				"TYPE": eventsByCriteria{
					"TRIGGER": criteriaInfo{
						isEnabled: true,
						data:      eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent},
					},
				},
				"PRIORITY": eventsByCriteria{
					"0": criteriaInfo{
						isEnabled: true,
						data:      eventsMap{testDefaultEvent.GetUUID(): testDefaultEvent},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				el := &eventsList{
					events:           tt.fields.events,
					eventsByCriteria: tt.fields.eventsByCriteria,
					mx:               tt.fields.mx,
				}
				el.AddEvent(tt.args.newEvent)
				if !reflect.DeepEqual(tt.wantEvents, el.events) || !reflect.DeepEqual(
					tt.wantCriteria,
					el.eventsByCriteria,
				) {
					t.Errorf(
						"Events want = %v\ngot = %v\n\nCriteria want = %v\ngot = %v",
						tt.wantEvents, el.events, tt.wantCriteria, el.eventsByCriteria,
					)
				}
			},
		)
	}
}
