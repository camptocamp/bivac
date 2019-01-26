package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/internal/utils"
	"github.com/camptocamp/bivac/pkg/volume"
)

func backupVolume(m *Manager, v *volume.Volume) (err error) {
	_, output, err := m.Orchestrator.DeployAgent(
		"cryptobioz/bivac:2.0.0",
		[]string{"agent"},
		os.Environ(),
		v,
	)
	if err != nil {
		err = fmt.Errorf("failed to deploy agent: %s", err)
		return
	}

	var agentOutput utils.MsgFormat
	err = json.Unmarshal([]byte(output), &agentOutput)
	if err != nil {
		log.Warningf("failed to unmarshal agent output: %s", err)
	}

	if agentOutput.Type == "error" {
		v.LastBackupStatus = "Failed"
	} else {
		v.LastBackupStatus = "Success"
		v.Logs = make(map[string]string)
		for stepKey, stepValue := range agentOutput.Content.(map[string]interface{}) {
			v.Logs[stepKey] = stepValue.(map[string]interface{})["stdout"].(string)
		}

	}
	v.LastBackupDate = time.Now().Format("2006-01-02 15:04:05")

	return
}
