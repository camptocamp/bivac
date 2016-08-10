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
		t.Fatalf("Expected %s, got %s", expected, e.String())
	}
}

func TestNewMetrics(t *testing.T) {
	p := NewMetrics("foo", "http://foo:9091")

	if p.Instance != "foo" {
		t.Fatalf("Expected instance to be foo, got %s", p.Instance)
	}

	if p.PushgatewayURL != "http://foo:9091" {
		t.Fatalf("Expected URL to be http://foo:9091, got %s", p.PushgatewayURL)
	}

	if len(p.Metrics) != 0 {
		t.Fatal("Expected empty Metrics array, got size %i", len(p.Metrics))
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

func TestParseEvent(t *testing.T) {
	if e := parseEvent(""); e != nil {
		t.Fatalf("Expected empty line to return nil, got %v", e)
	}

	if e := parseEvent("# HELP foo Some foo metric"); e != nil {
		t.Fatalf("Expected help line to return nil, got %v", e)
	}

	if e := parseEvent("# TYPE foo gauge"); e != nil {
		t.Fatalf("Expected type line to return nil, got %v", e)
	}

	if e := parseEvent("foo{bar=\"qux\"} 0"); e != nil {
		t.Fatalf("Expected non-conplicity event to return nil, got %v", e)
	}

	e := parseEvent("conplicity_foo{bar=\"qux\",baz=\"abc\"} 0")
	if e == nil {
		t.Fatal("Expected an event, got nil")
	}
	if e.Name != "conplicity_foo" {
		t.Fatalf("Expected event name to be conplicity_foo, got %s", e.Name)
	}
	if e.Value != "0" {
		t.Fatalf("Expected event value to be 0, got %s", e.Value)
	}
	if len(e.Labels) != 2 {
		t.Fatalf("Expected event to have two labels, got %s", len(e.Labels))
	}
	if e.Labels["bar"] != "qux" {
		t.Fatalf("Expected event's bar label to be \"qux\", got %s", e.Labels["bar"])
	}
}

func TestMetricUpdateEvent(t *testing.T) {
	m := &Metric{
		Name: "foo",
	}
	if len(m.Events) != 0 {
		t.Fatalf("Expected no events, got %i", len(m.Events))
	}

	// Add event
	e1 := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "bar",
	}
	m.UpdateEvent(e1)
	if len(m.Events) != 1 {
		t.Fatalf("Expected one event, got %i", len(m.Events))
	}
	if m.Events[0].Name != "foo" {
		t.Fatalf("Expected event name to be foo, got %s", m.Events[0].Name)
	}
	if m.Events[0].Value != "bar" {
		t.Fatalf("Expected event value to be bar, got %s", m.Events[0].Value)
	}

	// Update event
	e2 := &Event{
		Name: "foo",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "qux",
	}
	m.UpdateEvent(e2)
	if len(m.Events) != 1 {
		t.Fatalf("Expected one event, got %i", len(m.Events))
	}
	if m.Events[0].Name != "foo" {
		t.Fatalf("Expected event name to be foo, got %s", m.Events[0].Name)
	}
	if m.Events[0].Value != "qux" {
		t.Fatalf("Expected event value to be qux, got %s", m.Events[0].Value)
	}

	// Add new event
	e3 := &Event{
		Name: "bar",
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "quxx",
	}
	m.UpdateEvent(e3)
	if len(m.Events) != 2 {
		t.Fatalf("Expected two events, got %i", len(m.Events))
	}
	if m.Events[1].Name != "bar" {
		t.Fatalf("Expected event name to be bar, got %s", m.Events[1].Name)
	}
	if m.Events[1].Value != "quxx" {
		t.Fatalf("Expected event value to be quxx, got %s", m.Events[1].Value)
	}
}
