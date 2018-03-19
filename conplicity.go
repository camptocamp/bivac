package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/engines"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/orchestrators"
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

	log.Infof("Conplicity v%s starting backup...", version)

	orch := orchestrators.GetOrchestrator(c)

	vols, err := orch.GetVolumes()
	util.CheckErr(err, "Failed to get Docker volumes: %v", "fatal")

	for _, vol := range vols {
		vol.LogTime("backupStartTime")
		err = backupVolume(orch, vol)
		vol.LogTime("backupEndTime")
		if err != nil {
			log.Errorf("Failed to backup volume %s: %v", vol.Name, err)
			exitCode = 1
			continue
		}
	}

	log.Infof("End backup...")
	os.Exit(exitCode)
}

func backupVolume(o orchestrators.Orchestrator, vol *volume.Volume) (err error) {
	p := providers.GetProvider(o, vol)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found data provider")

	err = providers.PrepareBackup(p)
	if err != nil {
		err = fmt.Errorf("failed to prepare backup: %v", err)
		return
	}

	e := engines.GetEngine(o, vol)
	log.WithFields(log.Fields{
		"volume": vol.Name,
		"engine": e.GetName(),
	}).Info("Found backup engine")

	err = e.Backup()
	if err != nil {
		err = fmt.Errorf("failed to backup volume: %v", err)
		return
	}
	return
}
