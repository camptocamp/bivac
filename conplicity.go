package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/engines"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/metrics"
	"github.com/camptocamp/conplicity/providers"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
)

var version = "undefined"

func main() {
	var err error
	var exitCode int

	c, err := handler.NewConplicity(version)
	util.CheckErr(err, "Failed to setup Conplicity handler: %v", "fatal")

	log.Infof("Conplity v%s starting backup...", version)

	vols, err := c.GetVolumes()
	util.CheckErr(err, "Failed to get Docker volumes: %v", "fatal")

	for _, vol := range vols {
		events, err := backupVolume(c, vol)
		if err != nil {
			log.Errorf("Failed to backup volume %s: %v", vol.Name, err)
			exitCode = 1
			continue
		}
		for _, e := range events {
			c.UpdateEvent(e)
		}
	}

	err = c.MetricsHandler.Push()
	if err != nil {
		log.Errorf("Failed to post data to Prometheus Pushgateway: %v", err)
		exitCode = 2
	}

	log.Infof("End backup...")
	os.Exit(exitCode)
}

func backupVolume(c *handler.Conplicity, vol *volume.Volume) (metrics []*metrics.Event, err error) {
	p := providers.GetProvider(c, vol)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found data provider")
	err = providers.PrepareBackup(p)
	if err != nil {
		err = fmt.Errorf("failed to prepare backup: %v", err)
		return
	}

	e := engines.GetEngine(c, vol)
	log.WithFields(log.Fields{
		"volume": vol.Name,
		"engine": e.GetName(),
	}).Info("Found backup engine")

	metrics, err = e.Backup()
	if err != nil {
		err = fmt.Errorf("failed to backup volume: %v", err)
		return
	}
	return
}
