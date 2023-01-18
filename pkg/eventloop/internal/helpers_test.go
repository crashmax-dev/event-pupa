package internal

import (
	"context"
	"reflect"
	"testing"
)

func TestRemoveSliceItemByIndex(t *testing.T) {
	type args[T any] struct {
		s     []T
		index int
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want []T
	}
	tests := []testCase[int]{
		{
			name: "First",
			args: args[int]{s: []int{1, 2, 3, 4, 5}, index: 0},
			want: []int{2, 3, 4, 5},
		},
		{
			name: "Middle",
			args: args[int]{s: []int{1, 2, 3, 4, 5}, index: 2},
			want: []int{1, 2, 4, 5},
		},
		{
			name: "Last",
			args: args[int]{s: []int{1, 2, 3, 4, 5}, index: 4},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "First of Two",
			args: args[int]{s: []int{1, 2}, index: 0},
			want: []int{2},
		},
		{
			name: "Last of Two",
			args: args[int]{s: []int{1, 2}, index: 1},
			want: []int{1},
		},
		{
			name: "Single",
			args: args[int]{s: []int{1}, index: 0},
			want: []int{},
		},
		{
			name: "None",
			args: args[int]{s: []int{}, index: 4},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveSliceItemByIndex(tt.args.s, tt.args.index); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RemoveSliceItemByIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteToExecCh(t *testing.T) {
	var ch = make(chan string)
	type args struct {
		ctx    context.Context
		result string
	}
	tests := []struct {
		name string
		args args
		init func() string
		want string
	}{
		{
			name: "Default",
			args: args{ctx: context.WithValue(context.Background(), EXEC_CH_CTX_KEY, ch),
				result: "Hello RPTRP"},
			want: "Hello RPTRP",
			init: func() string {
				return <-ch
			},
		},
		{
			name: "No channel",
			args: args{ctx: context.Background(), result: "TEST"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			go WriteToExecCh(tt.args.ctx, tt.args.result)
			if tt.init != nil {
				result = tt.init()
			}
			if result != tt.want {
				t.Errorf("RemoveSliceItemByIndex() = %v, want %v", result, tt.want)
			}
		})
	}
}
