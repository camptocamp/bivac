package manager

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/internal/engine"
	"github.com/camptocamp/bivac/internal/utils"
	"github.com/camptocamp/bivac/pkg/volume"
)

func backupVolume(m *Manager, v *volume.Volume, force bool) (err error) {

	v.Mux.Lock()
	defer v.Mux.Unlock()

	useLogReceiver := false
	if m.LogServer != "" {
		useLogReceiver = true
	}

	p, err := m.Providers.GetProvider(m.Orchestrator, v)
	if err != nil {
		err = fmt.Errorf("failed to get provider: %s", err)
		return
	}

	if p.PreCmd != "" {
		err = RunCmd(p, m.Orchestrator, v, p.PreCmd, "precmd")
		if err != nil {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Warningf("failed to run pre-command: %s", err)
		}
	}

	cmd := []string{
		"agent",
		"backup",
		"-p",
		v.Mountpoint + v.SubPath + "/" + v.BackupDir,
		"-r",
		m.TargetURL + "/" + m.Orchestrator.GetPath(v) + "/" + v.RepoName,
		"--host",
		m.Orchestrator.GetPath(v),
	}

	if force {
		cmd = append(cmd, "--force")
	}

	if useLogReceiver {
		cmd = append(cmd, []string{"--log.receiver", m.LogServer + "/backup/" + v.ID + "/logs"}...)
	}

	_, output, err := m.Orchestrator.DeployAgent(
		m.AgentImage,
		cmd,
		os.Environ(),
		v,
	)
	if err != nil {
		err = fmt.Errorf("failed to deploy agent: %s", err)
		return
	}

	if !useLogReceiver {
		var agentOutput utils.MsgFormat
		err = json.Unmarshal([]byte(output), &agentOutput)
		if err != nil {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Warningf("failed to unmarshal agent output: %s -> `%s`", err, output)
		} else {
			m.updateBackupLogs(v, agentOutput)
		}
	} else {
		if output != "" {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Errorf("failed to send output: %s", output)
		}
	}

	if p.PostCmd != "" {
		err = RunCmd(p, m.Orchestrator, v, p.PostCmd, "postcmd")
		if err != nil {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Warningf("failed to run post-command: %s", err)
		}
	}
	return
}

func (m *Manager) updateBackupLogs(v *volume.Volume, agentOutput utils.MsgFormat) {
	if agentOutput.Type != "success" {
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
			m.setOldestBackupDate(v)
		} else {
			v.LastBackupStatus = "Failed"
			v.Metrics.LastBackupStatus.Set(1.0)
		}
	}

	v.LastBackupDate = time.Now().Format("2006-01-02 15:04:05")
	v.Metrics.LastBackupDate.SetToCurrentTime()
	return
}

func (m *Manager) setOldestBackupDate(v *volume.Volume) (err error) {
	// TODO: use regex
	stdout := strings.Split(v.Logs["snapshots"], "]")[1]

	var snapshots []engine.Snapshot

	err = json.Unmarshal([]byte(stdout), &snapshots)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal: %s", err)
		return
	}

	if len(snapshots) > 0 {
		v.Metrics.OldestBackupDate.Set(float64(snapshots[0].Time.Unix()))
	}

	return
}

// RunResticCommand runs a custom Restic command
func (m *Manager) RunResticCommand(v *volume.Volume, cmd []string) (output string, err error) {
	e := &engine.Engine{
		DefaultArgs: []string{
			"--no-cache",
			"-r",
			m.TargetURL + "/" + m.Orchestrator.GetPath(v) + "/" + v.RepoName,
		},
		Output: make(map[string]utils.OutputFormat),
	}

	err = e.RawCommand(cmd)

	output = e.Output["raw"].Stdout
	return
}
