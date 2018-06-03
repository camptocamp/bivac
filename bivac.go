package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/engines"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/providers"
	"github.com/camptocamp/bivac/util"
	"github.com/camptocamp/bivac/volume"
)

var version = "undefined"

func main() {
	var err error
	var exitCode int

	c, err := handler.NewBivac(version)
	util.CheckErr(err, "Failed to setup Bivac handler: %v", "fatal")

	log.Infof("Bivac v%s starting backup...", version)

	orch, err := orchestrators.GetOrchestrator(c)
	if err != nil {
		log.Fatalf("Failed to get an orchestrator: %s", err)
	}

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

	e := engines.GetEngine(o, vol)
	log.WithFields(log.Fields{
		"volume": vol.Name,
		"engine": e.GetName(),
	}).Info("Found backup engine")

	p := providers.GetProvider(o, vol)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found data provider")

	if cmd := p.GetPrepareCommand(""); cmd != nil {
		if e.StdinSupport() {
			pbErr := make(chan error)
			log.Debugf("preparing backup...")
			go providers.PrepareBackup(p, pbErr)
			log.Debugf("waiting for assignment")
			err = <-pbErr
			log.Debugf("error assigned")
			if err != nil {
				err = fmt.Errorf("failed to prepare backup: %s", err)
				return
			}
			log.Debugf("Backuping...")
			err = e.Backup()
		} else {
			_, err = providers.PrepareBackup(p, nil)
			if err != nil {
				err = fmt.Errorf("failed to prepare backup: %s", err)
				return
			}
			err = e.Backup()
		}
		if err != nil {
			err = fmt.Errorf("failed to backup volume: %s", err)
			return
		}
	} else {
		err = e.Backup()
		if err != nil {
			err = fmt.Errorf("failed to backup volume: %s", err)
			return
		}
	}
	/*

		if e.StdinSupport {
			go providers.PrepareBackupToPipe(p)
		} else {
			vp, err := providers.PrepareBackupToVolume(p)
			err = e.BackupVolume()
			if err != nil {
				err = fmt.Errorf("failed to backup volume: %v", err)
				return
			}
		}
		if err != nil {
			err = fmt.Errorf("failed to prepare backup: %v", err)
			return
		}

	*/
	return
}
