package internal

import (
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
