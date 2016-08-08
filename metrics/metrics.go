package metrics

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
)

// PrometheusMetrics is a struct to push metrics to Prometheus
type PrometheusMetrics struct {
	Handler *handler.Conplicity
}

// NewMetrics returns a new metrics struct
func NewMetrics(c *handler.Conplicity) *PrometheusMetrics {
	return &PrometheusMetrics{
		Handler: c,
	}
}

// Push sends metrics to a Prometheus push gateway
func (p *PrometheusMetrics) Push(replace bool) (err error) {
	var verb string
	if replace {
		verb = "PUT"
	} else {
		verb = "POST"
	}

	c := p.Handler
	if len(c.Metrics) == 0 || c.Config.Metrics.PushgatewayURL == "" {
		return
	}

	url := c.Config.Metrics.PushgatewayURL + "/metrics/job/conplicity/instance/" + c.Hostname
	data := strings.Join(c.Metrics, "\n") + "\n"

	log.WithFields(log.Fields{
		"data": data,
		"verb": verb,
		"url":  url,
	}).Debug("Sending metrics to Prometheus Pushgateway")

	req, err := http.NewRequest(verb, url, bytes.NewBufferString(data))
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
		"resp": body,
	}).Debug("Received Prometheus response")

	return
}
