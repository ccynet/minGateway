package fixedqueue

import (
	"fmt"
	"reflect"
	"testing"
)

func TestFixedQueue(t *testing.T) {
	q := NewFixedQueue(8)
	q.Push(0)
	q.Push(1)
	q.Push(2)
	q.Push(3)
	q.Push(4)
	q.Print()
	fmt.Println(q.Len())
	q.Push(5)
	q.Push(6)
	q.Push(7)
	q.Push(8)
	q.Push(9)

	q.Print()
	fmt.Println(q.Len())

	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())

	fmt.Println(q.Len())

	//q.Clear()
	//fmt.Println(q.Len())

	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())
	fmt.Println(q.Pop())

	fmt.Println(q.Len())
	q.Print()
}

func TestNewFixedQueue(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want *FixedQueue
	}{
		// TODO: Add test cases.
		{
			name: "testFixedQueue1",
			args: args{size: 10},
			want: NewFixedQueue(10),
		},
		{
			name: "testFixedQueue2",
			args: args{size: 50},
			want: NewFixedQueue(50),
		},
		{
			name: "testFixedQueue3",
			args: args{size: 1000},
			want: NewFixedQueue(1000),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFixedQueue(tt.args.size)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFixedQueue() = %v, want %v", got, tt.want)
			}
			for j := 0; j < 100000; j++ {
				for i := 0; i < 100; i++ {
					got.Push(i)
				}
				for i := 0; i < 50; i++ {
					got.Pop()
				}
			}
			t.Log(got)
		})
	}
}

func BenchmarkNewFixedQueue(b *testing.B) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want *FixedQueue
	}{
		// TODO: Add test cases.
		{
			name: "testFixedQueue1",
			args: args{size: 10},
			want: NewFixedQueue(10),
		},
		{
			name: "testFixedQueue2",
			args: args{size: 50},
			want: NewFixedQueue(50),
		},
		{
			name: "testFixedQueue3",
			args: args{size: 1000},
			want: NewFixedQueue(1000),
		},
	}
	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			got := NewFixedQueue(tt.args.size)
			if !reflect.DeepEqual(got, tt.want) {
				b.Errorf("NewFixedQueue() = %v, want %v", got, tt.want)
			}
			for j := 0; j < b.N; j++ {
				for i := 0; i < 100; i++ {
					got.Push(i)
				}
				for i := 0; i < 50; i++ {
					got.Pop()
				}
			}
			b.Log(got)
		})
	}
}
