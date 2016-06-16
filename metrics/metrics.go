package metrics

import (
	"bytes"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

func PushToPrometheus(url, data string) error {
	log.Infof("Sending metrics to Prometheus Pushgateway: %v", data)
	log.Debugf("URL=%v", url)

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "text/plain; version=0.0.4")

	client := &http.Client{}
	resp, err := client.Do(req)

	log.Debugf("resp = %v", resp)

	return err
}
