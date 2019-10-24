package manager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/internal/engine"
	"github.com/camptocamp/bivac/internal/utils"
	"github.com/camptocamp/bivac/pkg/volume"
)

func backupVolume(m *Manager, v *volume.Volume, force bool) (err error) {

	v.BackingUp = true
	defer func() { v.BackingUp = false }()

	v.Mux.Lock()
	defer v.Mux.Unlock()

	useLogReceiver := false
	if m.LogServer != "" {
		useLogReceiver = true
	}

	v.LastBackupStartDate = time.Now().Format("2006-01-02 15:04:05")

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
		decodedOutput, err := base64.StdEncoding.DecodeString(strings.Replace(output, " ", "", -1))
		if err != nil {
			log.Errorf("failed to decode agent output of `%s` : %s -> `%s`", v.Name, err, strings.Replace(output, " ", "", -1))
		} else {
			var agentOutput utils.MsgFormat
			err = json.Unmarshal(decodedOutput, &agentOutput)
			if err != nil {
				log.WithFields(log.Fields{
					"volume":   v.Name,
					"hostname": v.Hostname,
				}).Warningf("failed to unmarshal agent output: %s -> `%s`", err, strings.TrimSpace(output))
			}

			m.updateBackupLogs(v, agentOutput)
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

func (m *Manager) attachOrphanAgent(containerID string, v *volume.Volume) {
	defer func() { v.BackingUp = false }()

	p, err := m.Providers.GetProvider(m.Orchestrator, v)
	if err != nil {
		err = fmt.Errorf("failed to get provider: %s", err)
		return
	}
	useLogReceiver := false
	if m.LogServer != "" {
		useLogReceiver = true
	}

	_, output, err := m.Orchestrator.AttachOrphanAgent(containerID, v.Namespace)
	if err != nil {
		log.WithFields(log.Fields{
			"volume":   v.Name,
			"hostname": v.Hostname,
		}).Errorf("failed to attach orphan agent: %s", err)
		return
	}

	if !useLogReceiver {
		decodedOutput, err := base64.StdEncoding.DecodeString(strings.TrimSpace(output))
		if err != nil {
			log.Errorf("failed to decode agent output of `%s` : %s -> `%s`", v.Name, err, output)
		} else {
			var agentOutput utils.MsgFormat
			err = json.Unmarshal(decodedOutput, &agentOutput)
			if err != nil {
				log.WithFields(log.Fields{
					"volume":   v.Name,
					"hostname": v.Hostname,
				}).Warningf("failed to unmarshal agent output: %s -> `%s`", err, output)
			}

			m.updateBackupLogs(v, agentOutput)
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
			stdout, _ := base64.StdEncoding.DecodeString(stepValue.(map[string]interface{})["stdout"].(string))
			v.Logs[stepKey] = fmt.Sprintf("[%d] %s", int(stepValue.(map[string]interface{})["rc"].(float64)), stdout)
		}
		if success {
			v.LastBackupStatus = "Success"
			v.Metrics.LastBackupStatus.Set(0.0)
			err := m.setOldestBackupDate(v)
			if err != nil {
				log.Errorf("failed to set oldest backup date: %s", err)
			}
		} else {
			v.LastBackupStatus = "Failed"
			v.Metrics.LastBackupStatus.Set(1.0)
		}
	}

	v.LastBackupDate = time.Now().Format("2006-01-02 15:04:05")
	return
}

func (m *Manager) setOldestBackupDate(v *volume.Volume) (err error) {
	r, err := regexp.Compile(`\S{3} (.*)`)
	stdout := r.FindStringSubmatch(v.Logs["snapshots"])[1]

	var snapshots []engine.Snapshot

	err = json.Unmarshal([]byte(stdout), &snapshots)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal: %s", err)
		return
	}

	if len(snapshots) > 0 {
		v.Metrics.OldestBackupDate.Set(float64(snapshots[0].Time.Unix()))
		v.Metrics.LastBackupDate.Set(float64(snapshots[len(snapshots)-1].Time.Unix()))
		v.Metrics.BackupCount.Set(float64(len(snapshots)))
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
