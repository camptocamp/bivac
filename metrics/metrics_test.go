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

func TestEventString(t *testing.T) {
	e := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume":   "baz",
			"instance": "qux",
		},
		Value: "bar",
	}
	expected := "foo{volume=\"baz\",instance=\"qux\"} bar"
	if e.String() != expected {
		t.Fatal("Expected %s, got %s", expected, e.String())
	}
}

func TestNewMetric(t *testing.T) {
	p := NewMetrics("foo", "http://foo:9091")
	m := p.NewMetric("bar", "qux")

	if len(p.Metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %i", len(p.Metrics))
	}

	if p.Metrics["bar"] != m {
		t.Fatal("Expected to find metric in handler")
	}

	if m.Name != "bar" {
		t.Fatalf("Expected name to be bar, got %s", m.Name)
	}

	if m.Type != "qux" {
		t.Fatalf("Expected type to be qux, got %s", m.Name)
	}
}
