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
	p := NewMetrics("foo", "bar", "http://foo:9091")

	if p.Instance != "foo" {
		t.Fatalf("Expected instance to be foo, got %s", p.Instance)
	}

	if p.Volume != "bar" {
		t.Fatalf("Expected volume to be bar, got %s", p.Volume)
	}

	if p.PushgatewayURL != "http://foo:9091" {
		t.Fatalf("Expected URL to be http://foo:9091, got %s", p.PushgatewayURL)
	}

	if len(p.Metrics) != 0 {
		t.Fatalf("Expected empty Metrics array, got size %v", len(p.Metrics))
	}
}

func TestNewMetric(t *testing.T) {
	p := NewMetrics("foo", "baz", "http://foo:9091")
	m := p.NewMetric("bar", "qux")

	if len(p.Metrics) != 1 {
		t.Fatalf("Expected 1 metric, got %v", len(p.Metrics))
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

func TestMetricUpdateEvent(t *testing.T) {
	var err error
	m := &Metric{
		Name: "foo",
	}
	if len(m.Events) != 0 {
		t.Fatalf("Expected no events, got %v", len(m.Events))
	}

	// Add event
	e1 := &Event{
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "bar",
	}
	err = m.UpdateEvent(e1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(m.Events) != 1 {
		t.Fatalf("Expected one event, got %v", len(m.Events))
	}
	if m.Events[0].Name != "foo" {
		t.Fatalf("Expected event name to be foo, got %s", m.Events[0].Name)
	}
	if m.Events[0].Value != "bar" {
		t.Fatalf("Expected event value to be bar, got %s", m.Events[0].Value)
	}

	// Update event
	e2 := &Event{
		Labels: map[string]string{
			"volume": "baz",
		},
		Value: "qux",
	}
	err = m.UpdateEvent(e2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(m.Events) != 1 {
		t.Fatalf("Expected one event, got %v", len(m.Events))
	}
	if m.Events[0].Name != "foo" {
		t.Fatalf("Expected event name to be foo, got %s", m.Events[0].Name)
	}
	if m.Events[0].Value != "qux" {
		t.Fatalf("Expected event value to be qux, got %s", m.Events[0].Value)
	}

	// Add new event
	e3 := &Event{
		Labels: map[string]string{
			"volume": "foo",
		},
		Value: "quxx",
	}
	err = m.UpdateEvent(e3)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(m.Events) != 2 {
		t.Fatalf("Expected two events, got %v", len(m.Events))
	}
	if m.Events[1].Name != "foo" {
		t.Fatalf("Expected event name to be foo, got %s", m.Events[1].Name)
	}
	if m.Events[1].Value != "quxx" {
		t.Fatalf("Expected event value to be quxx, got %s", m.Events[1].Value)
	}

	// Add event with wrong name
	e4 := &Event{
		Name: "bar",
		Labels: map[string]string{
			"volume": "fooddd",
		},
		Value: "quxx",
	}
	err = m.UpdateEvent(e4)
	if err == nil {
		t.Fatal("Expected an error, got nil")
	}
	if len(m.Events) != 2 {
		t.Fatalf("Expected two events, got %v", len(m.Events))
	}
}
