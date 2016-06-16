package metrics

import (
	"bytes"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
)

func PushToPrometheus(c *handler.Conplicity) error {
	url := c.Config.Metrics.PushgatewayURL + "/metrics/job/conplicity/instance/" + c.Hostname
	data := strings.Join(c.Metrics, "\n") + "\n"

	log.Infof("Sending metrics to Prometheus Pushgateway: %v", data)
	log.Debugf("URL=%v", url)

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "text/plain; version=0.0.4")

	client := &http.Client{}
	resp, err := client.Do(req)

	log.Debugf("resp = %v", resp)

	return err
}
