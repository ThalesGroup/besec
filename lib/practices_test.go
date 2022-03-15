package lib

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/afero"
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

func TestDeltas(t *testing.T) {
	fs := afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), ".."))
	parser := NewPracticeParser("/demo/practices", "/demo/local-practices", "/practices/schema.json", fs)
	practices, err := parser.ParsePracticesDir()
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, p := range practices {
		count++
		switch p.ID {
		case "gibson":
			t.Error("Delta application didn't remove a practice it should have")
		case "demoPractice":
			{
				if !strings.Contains(p.Notes, "local modifications") {
					t.Error("Delta application didn't update practice notes")
				}

				task, _ := p.TaskFromID("triageNew")
				if task.Level != 4 {
					t.Error("Delta application didn't preserve task level")
				}
				if len(task.Questions) != 1 {
					t.Error("Delta application didn't preserve task questions")
				}
				if strings.Contains(task.Description, "little lacking") {
					t.Error("Delta application didn't update task description")
				}

				_, found := p.TaskFromID("removed")
				if found {
					t.Error("Delta application didn't exclude a task it should have")
				}
				_, found = p.TaskFromID("demo")
				if !found {
					t.Error("Delta application didn't add a task it should have")
				}
			}
		case "local":
		default:
			t.Errorf("Unexpected demo practice found '%v'", p.ID)
		}
	}
	if count != 2 {
		t.Errorf("Delta application resulted in %v practices, expected 2", count)
	}
}
