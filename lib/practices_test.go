package lib

import (
	"reflect"
	"testing"
)

func TestTasksByLevel(t *testing.T) {
	one := Task{ID: "one", Level: 1}
	two := Task{ID: "two", Level: 2}
	alsotwo := Task{ID: "two_again", Level: 2}
	three := Task{ID: "three", Level: 3}

	cases := []struct {
		in   Practice
		want map[uint8][]Task
	}{
		{Practice{ID: "p", Tasks: []Task{one}}, map[uint8][]Task{1: {one}}},
		{Practice{ID: "p", Tasks: []Task{one, two}}, map[uint8][]Task{1: {one}, 2: {two}}},
		{Practice{ID: "p", Tasks: []Task{two, one}}, map[uint8][]Task{1: {one}, 2: {two}}},
		{Practice{ID: "p", Tasks: []Task{three, one, two, alsotwo}}, map[uint8][]Task{1: {one}, 2: {two, alsotwo}, 3: {three}}},
	}

	for _, c := range cases {
		got := c.in.TasksByLevel()
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%v.TasksByLevel() == %v, want %v", c.in, got, c.want)
		}
	}
}
