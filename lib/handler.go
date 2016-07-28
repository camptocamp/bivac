package conplicity

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/util"
	docker "github.com/docker/engine-api/client"
)

// Conplicity is the main handler struct
type Conplicity struct {
	*docker.Client
	Config   *config.Config
	Hostname string
	Metrics  []string
}

// Setup sets up a Conplicity struct
func (c *Conplicity) Setup(version string) (err error) {
	c.Config = config.LoadConfig(version)

	err = c.setupLoglevel()
	util.CheckErr(err, "Failed to setup log level: %v", "fatal")

	c.Hostname, err = os.Hostname()
	util.CheckErr(err, "Failed to get hostname: %v", "fatal")

	err = c.SetupDocker()
	util.CheckErr(err, "Failed to setup docker: %v", "fatal")

	return
}

// SetupDocker for the  client
func (c *Conplicity) SetupDocker() (err error) {
	c.Client, err = docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	util.CheckErr(err, "Failed to create Docker client: %v", "fatal")
	return
}

func (c *Conplicity) setupLoglevel() (err error) {
	switch c.Config.Loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		errMsg := fmt.Sprintf("Wrong log level '%v'", c.Config.Loglevel)
		err = errors.New(errMsg)
	}

	if c.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return
}

// PushToPrometheus sends metrics to a Prometheus push gateway
func (c *Conplicity) PushToPrometheus() (err error) {
	if len(c.Metrics) == 0 || c.Config.Metrics.PushgatewayURL == "" {
		return
	}

	url := c.Config.Metrics.PushgatewayURL + "/metrics/job/conplicity/instance/" + c.Hostname
	data := strings.Join(c.Metrics, "\n") + "\n"

	log.WithFields(log.Fields{
		"data": data,
		"url":  url,
	}).Debug("Sending metrics to Prometheus Pushgateway")

	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
	util.CheckErr(err, "Failed to create HTTP request to send metrics to Prometheus: %v", "error")

	req.Header.Set("Content-Type", "text/plain; version=0.0.4")

	client := &http.Client{}
	resp, err := client.Do(req)
	util.CheckErr(err, "Failed to get HTTP response from sending metrics to Prometheus: %v", "error")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	util.CheckErr(err, "Failed to read HTTP response from sending metrics to Prometheus: %v", "error")

	log.WithFields(log.Fields{
		"resp": body,
	}).Debug("Received Prometheus response")

	return
}
