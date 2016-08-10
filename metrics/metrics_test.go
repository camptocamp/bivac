package metrics

import "testing"

func TestEventEquals(t *testing.T) {
	e1 := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "bar",
	}
	if !e1.Equals(e1) {
		t.Fatal("Expected event to equal itself")
	}

	e2 := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "qux",
	}
	if !e1.Equals(e2) {
		t.Fatal("Expected event e1 to equal e2 (different value)")
	}

	e3 := &Event{
		Name: "qux",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "bar",
	}
	if e1.Equals(e3) {
		t.Fatal("Expected event e1 to not equal e3 (different name)")
	}

	e4 := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume": "qux",
		},
		Value: "bar",
	}
	if e1.Equals(e4) {
		t.Fatal("Expected event e1 to not equal e4 (different volume)")
	}
}
