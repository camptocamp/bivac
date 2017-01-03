package metrics

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// PrometheusMetrics is a struct to push metrics to Prometheus
type PrometheusMetrics struct {
	Instance       string
	Volume         string
	PushgatewayURL string
	Metrics        map[string]*Metric
}

// Metric is a Prometheus Metric
type Metric struct {
	Name   string
	Events []*Event
	Type   string
}

// Event is a Prometheus Metric Event
type Event struct {
	Name   string
	Labels map[string]string
	Value  string
}

// NewMetrics returns a new metrics struct
func NewMetrics(instance, volume, pushgatewayURL string) *PrometheusMetrics {
	return &PrometheusMetrics{
		Instance:       instance,
		Volume:         volume,
		PushgatewayURL: pushgatewayURL,
		Metrics:        make(map[string]*Metric),
	}
}

// String formats an event for printing
func (e *Event) String() string {
	var labels []string
	for l, v := range e.Labels {
		labels = append(labels, fmt.Sprintf("%s=\"%s\"", l, v))
	}
	return fmt.Sprintf("%s{%s} %s", e.Name, strings.Join(labels, ","), e.Value)
}

// Equals checks if two Events refer to the same Prometheus event
func (e *Event) Equals(newEvent *Event) bool {
	if e.Name != newEvent.Name {
		return false
	}

	if e.Labels["volume"] != newEvent.Labels["volume"] {
		return false
	}

	return true
}

// UpdateEvent adds an event, or updates it if the event already exists
func (m *Metric) UpdateEvent(event *Event) error {
	if event.Name == "" {
		event.Name = m.Name
	}
	if event.Name != m.Name {
		return fmt.Errorf("cannot add event %s to metric %s", event.Name, m.Name)
	}
	for i, e := range m.Events {
		if e.Equals(event) {
			log.WithFields(log.Fields{
				"metric":    m.Name,
				"old_event": e.String(),
				"new_event": event.String(),
			}).Debug("Replacing event")
			m.Events[i] = event
			return nil
		}
	}
	log.WithFields(log.Fields{
		"metric": m.Name,
		"event":  event.String(),
	}).Debug("Adding event")
	m.Events = append(m.Events, event)

	return nil
}

// NewMetric adds a new metric if it doesn't exist yet
// or returns the existing matching metric otherwise
func (p *PrometheusMetrics) NewMetric(name, mType string) (m *Metric) {
	m, ok := p.Metrics[name]
	if !ok {
		m = &Metric{
			Name: name,
		}
		p.Metrics[name] = m
	}
	m.Type = mType
	return
}

// Push sends metrics to a Prometheus push gateway
func (p *PrometheusMetrics) Push() (err error) {
	if p.PushgatewayURL == "" {
		log.Debug("No Pushgateway URL specified, not pushing metrics")
		return
	}
	metrics := p.Metrics
	url := p.PushgatewayURL + "/metrics/job/conplicity/instance/" + p.Instance + "/volume/" + p.Volume

	var data string
	for _, m := range metrics {
		if m.Type != "" {
			data += fmt.Sprintf("# TYPE %s %s\n", m.Name, m.Type)
		}
		for _, e := range m.Events {
			data += fmt.Sprintf("%s\n", e)
		}
	}
	data += "\n"

	log.WithFields(log.Fields{
		"data": data,
		"url":  url,
	}).Debug("Sending metrics to Prometheus Pushgateway")

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
	if err != nil {
		err = fmt.Errorf("failed to create HTTP request: %v", err)
		return
	}

	req.Header.Set("Content-Type", "text/plain; version=0.0.4")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("failed to get HTTP response: %v", err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read HTTP response: %v", err)
		return
	}

	log.WithFields(log.Fields{
		"resp": string(body),
	}).Debug("Received Prometheus response")

	return
}
