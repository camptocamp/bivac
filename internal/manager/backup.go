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

func backupVolume(m *Manager, v *volume.Volume, force bool) (err error) {
	p, err := m.Providers.GetProvider(m.Orchestrator, v)
	if err != nil {
		err = fmt.Errorf("failed to get provider: %s", err)
		return
	}

	if p.PreCmd != "" {
		err = RunCmd(p, m.Orchestrator, v, p.PreCmd)
		if err != nil {
			log.Warningf("failed to run pre-command: %s", err)
		}
	}
	cmd := []string{
		"agent",
		"backup",
		"-p",
		v.Mountpoint + "/" + v.BackupDir,
		"-r",
		m.TargetURL + "/" + m.Orchestrator.GetPath(v) + "/" + v.Name,
		"--host",
		v.Hostname,
	}

	if force {
		cmd = append(cmd, "--force")
	}

	_, output, err := m.Orchestrator.DeployAgent(
		"cryptobioz/bivac:2.0.0",
		cmd,
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
		v.Metrics.LastBackupStatus.Set(1.0)
	} else {
		success := true
		v.Logs = make(map[string]string)
		for stepKey, stepValue := range agentOutput.Content.(map[string]interface{}) {
			if stepKey != "testInit" && stepValue.(map[string]interface{})["rc"].(float64) > 0.0 {
				success = false
			}
			v.Logs[stepKey] = fmt.Sprintf("[%d] %s", int(stepValue.(map[string]interface{})["rc"].(float64)), stepValue.(map[string]interface{})["stdout"].(string))
		}
		if success {
			v.LastBackupStatus = "Success"
			v.Metrics.LastBackupStatus.Set(0.0)
		} else {
			v.LastBackupStatus = "Failed"
			v.Metrics.LastBackupStatus.Set(1.0)
		}

	}
	v.LastBackupDate = time.Now().Format("2006-01-02 15:04:05")
	v.Metrics.LastBackupDate.SetToCurrentTime()

	if p.PostCmd != "" {
		err = RunCmd(p, m.Orchestrator, v, p.PostCmd)
		if err != nil {
			log.Warningf("failed to run post-command: %s", err)
		}
	}
	return
}
