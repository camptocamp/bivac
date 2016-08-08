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
func NewMetrics(instance, pushgatewayURL string) *PrometheusMetrics {
	return &PrometheusMetrics{
		Instance:       instance,
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
	if e.Name == newEvent.Name {
		return false
	}

	if e.Labels["volume"] != newEvent.Labels["volume"] {
		return false
	}

	return true
}

// UpdateEvent adds an event, or updates it if the event already exists
func (m *Metric) UpdateEvent(event *Event) {
	event.Name = m.Name
	var found bool
	for _, e := range m.Events {
		if e.Equals(event) {
			log.WithFields(log.Fields{
				"metric":    m.Name,
				"old_event": e.String(),
				"new_event": event.String(),
			}).Debug("Replacing event")
			e = event
			found = true
			break
		}
	}
	if !found {
		log.WithFields(log.Fields{
			"metric": m.Name,
			"event":  event.String(),
		}).Debug("Adding event")
		m.Events = append(m.Events, event)
	}
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

// GetMetrics returns a map of existing metrics
func (p *PrometheusMetrics) GetMetrics() (err error) {
	if p.PushgatewayURL == "" {
		log.Debug("No Pushgateway URL specified, not retrieving metrics")
		return
	}
	log.Debug("Retrieving existing metrics")
	url := p.PushgatewayURL + "/metrics"
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("failed to get existing metrics from Prometheus: %v", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read HTTP response: %v", err)
		return
	}
	log.Debug(string(body))
	for _, l := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(l, "# TYPE ") {
			lSplit := strings.Fields(l)
			mName := lSplit[2]
			mType := lSplit[3]
			m, ok := p.Metrics[mName]
			if !ok {
				m = &Metric{
					Name: mName,
				}
			}
			m.Type = mType
			continue
		}
		e := parseEvent(l)
		if e == nil {
			continue
		}
		if e.Labels["instance"] != p.Instance {
			log.WithFields(log.Fields{
				"event":    e.Name,
				"instance": e.Labels["instance"],
			}).Debug("Ignoring event from wrong instance")
			continue
		}
		log.WithFields(log.Fields{
			"event":  e.Name,
			"value":  e.Value,
			"labels": e.Labels,
		}).Debug("Found event")
		m, ok := p.Metrics[e.Name]
		if !ok {
			m = &Metric{
				Name: e.Name,
			}
			p.Metrics[e.Name] = m
		}
		m.Events = append(m.Events, e)
	}

	return
}

func parseEvent(line string) (event *Event) {
	// Filter out metrics and comments
	if !strings.HasPrefix(line, "conplicity") {
		return
	}

	spaceSplit := strings.Fields(line)
	BraceSplit := strings.Split(spaceSplit[0], "{")
	name := BraceSplit[0]
	labelsStr := strings.TrimSuffix(strings.TrimPrefix(spaceSplit[0], fmt.Sprintf("%s{", name)), "}")
	labels := make(map[string]string)
	for _, l := range strings.Split(labelsStr, ",") {
		lParse := strings.Split(l, "=")
		lName := lParse[0]
		lValue := strings.TrimSuffix(strings.TrimPrefix(l, fmt.Sprintf("%s=\"", lName)), "\"")
		labels[lName] = lValue
	}
	event = &Event{
		Name:   name,
		Value:  spaceSplit[1],
		Labels: labels,
	}
	return
}

// Push sends metrics to a Prometheus push gateway
func (p *PrometheusMetrics) Push() (err error) {
	if p.PushgatewayURL == "" {
		log.Debug("No Pushgateway URL specified, not pushing metrics")
		return
	}
	metrics := p.Metrics
	url := p.PushgatewayURL + "/metrics/job/conplicity/instance/" + p.Instance

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
