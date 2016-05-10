package util

import (
  "os"
  "io/ioutil"
  "time"
  "fmt"

	log "github.com/Sirupsen/logrus"
)

// CheckErr checks for error, logs and optionally exits the program
func CheckErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
		}
	}
}

// MonitoringStatus generates status from backup runs, which the prometheus
// node exporter can pick up
func MonitoringStatus(volume string) {
	ts := time.Now().Unix()
	stamp := fmt.Sprint(ts)
	labels := fmt.Sprintf("volume=\"%s\",what=\"lastruntimestamp\"", volume)
	metric := fmt.Sprintf("conplicity{%s} %s\n", labels, stamp)
	text := []byte(metric)
	err := ioutil.WriteFile(volume + ".prom", text, 0644)
	CheckErr(err, "Failed writing to monitoring file: " + volume, 1)
}
