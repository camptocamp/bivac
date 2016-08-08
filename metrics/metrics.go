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

// Push sends metrics to a Prometheus push gateway
func (p *PrometheusMetrics) Push() (err error) {
	metrics := p.Metrics
	if len(metrics) == 0 || p.PushgatewayURL == "" {
		return
	}

	url := p.PushgatewayURL + "/metrics/job/conplicity/instance/" + p.Instance

	var data string
	for _, m := range metrics {
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
