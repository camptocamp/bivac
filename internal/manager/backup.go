package manager

import (
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/pkg/volume"
)

func backupVolume(m *Manager, v *volume.Volume) (err error) {
	success, output, err := m.Orchestrator.DeployAgent(
		"cryptobioz/bivac:2.0.0",
		[]string{"agent"},
		os.Environ(),
		v,
	)
	if err != nil {
		err = fmt.Errorf("failed to deploy agent: %s", err)
		return
	}

	log.Debugf("## %s", output)

	if success {
		v.LastBackupStatus = "Success"
	} else {
		v.LastBackupStatus = "Failed"
	}
	v.LastBackupDate = time.Now().Format("2006.01.02 15:04:05")

	return
}
