package engines

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/metrics"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/util"
	"github.com/camptocamp/bivac/volume"
)

// ResticEngine implements a backup engine with Restic
type ResticEngine struct {
	Orchestrator orchestrators.Orchestrator
	Volume       *volume.Volume
}

// Snapshot is a struct returned by the function snapshots()
type Snapshot struct {
	Time     time.Time `json:"time"`
	Parent   string    `json:"parent"`
	Tree     string    `json:"tree"`
	Path     []string  `json:"path"`
	Hostname string    `json:"hostname"`
	ID       string    `json:"id"`
	ShortID  string    `json:"short_id"`
}

// GetName returns the engine name
func (*ResticEngine) GetName() string {
	return "Restic"
}

// replaceArgs replace arguments with their values
func (r *ResticEngine) replaceArgs(args []string) (newArgs []string) {
	log.Debugf("Replacing args, Input: %v", args)
	for _, arg := range args {
		arg = strings.Replace(arg, "%D", r.Volume.BackupDir, -1)
		arg = strings.Replace(arg, "%T", r.Volume.Target, -1)
		arg = strings.Replace(arg, "%V", r.Volume.Name, -1)
		newArgs = append(newArgs, arg)
	}
	log.Debugf("Replacing args, Output: %v", newArgs)
	return
}

// Backup performs the backup of the passed volume
func (r *ResticEngine) Backup() (err error) {

	v := r.Volume

	targetURL, err := url.Parse(v.Config.TargetURL)
	if err != nil {
		err = fmt.Errorf("failed to parse target URL: %v", err)
		return
	}

	c := r.Orchestrator.GetHandler()
	v.Target = targetURL.String() + "/" + r.Orchestrator.GetPath(v)
	v.BackupDir = v.Mountpoint + "/" + v.BackupDir
	v.Mount = v.Name + ":" + v.Mountpoint + ":ro"

	err = util.Retry(3, r.init)
	if err != nil {
		err = fmt.Errorf("failed to create a secure repository: %v", err)
		r.sendBackupStatus(1, v.Name)
		return
	}

	err = util.Retry(3, r.resticBackup)
	if err != nil {
		err = fmt.Errorf("failed to backup the volume: %v", err)
		r.sendBackupStatus(1, v.Name)
		return
	}

	err = util.Retry(3, r.forget)
	if err != nil {
		err = fmt.Errorf("failed to forget the oldest snapshots: %v", err)
		r.sendBackupStatus(1, v.Name)
		return
	}

	if c.IsCheckScheduled(v) {
		err = util.Retry(3, r.verify)
		if err != nil {
			err = fmt.Errorf("failed to verify backup: %v", err)
			r.sendBackupStatus(1, v.Name)
			return
		}
	}

	r.sendBackupStatus(0, v.Name)

	return
}

// init initialize a secure repository
func (r *ResticEngine) init() (err error) {
	v := r.Volume

	// Check if the repository already exists
	state, _, err := r.launchRestic(
		[]string{
			"-r",
			"%T",
			"snapshots",
		},
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to verify the existence of the repository: %v", err)
		return
	}
	if state == 0 {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Info("The repository already exists, skipping initialization.")
		return nil
	}

	// Initialize the repository
	state, _, err = r.launchRestic(
		[]string{
			"-r",
			"%T",
			"init",
		},
		[]*volume.Volume{
			v,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to initialize the repository: %v", err)
		return
	}
	if state != 0 {
		err = fmt.Errorf("Restic exited with state %v while initializing the repository", state)
		return
	}
	return
}

// resticBackup performs the backup of a volume with Restic
func (r *ResticEngine) resticBackup() (err error) {
	c := r.Orchestrator.GetHandler()
	v := r.Volume
	state, _, err := r.launchRestic(
		[]string{
			"--hostname",
			c.Hostname,
			"-r",
			"%T",
			"backup",
			"%D",
		},
		[]*volume.Volume{
			v,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to backup the volume: %v", err)
	}
	if state != 0 {
		err = fmt.Errorf("Restic exited with state %v while backuping the volume", state)
	}

	metric := r.Volume.MetricsHandler.NewMetric("bivac_backupExitCode", "gauge")
	metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.Itoa(state),
		},
	)
	return
}

// verify checks that the backup is usable
func (r *ResticEngine) verify() (err error) {
	v := r.Volume
	state, _, err := r.launchRestic(
		[]string{
			"-r",
			"%T",
			"check",
		},
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to check the backup: %v", err)
		return
	}
	if state == 0 {
		now := time.Now().Local()
		os.Chtimes(v.Mountpoint+"/.bivac_last_check", now, now)
	} else {
		err = fmt.Errorf("Restic exited with state %v while checking the backup", state)
	}

	metric := r.Volume.MetricsHandler.NewMetric("bivac_verifyExitCode", "gauge")
	err = metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.Itoa(state),
		},
	)
	return
}

// forget removes a snapshot
func (r *ResticEngine) forget() (err error) {

	v := r.Volume

	snapshots, err := r.snapshots()
	if err != nil {
		return
	}
	if len(snapshots) == 0 {
		err = errors.New("No snapshots found but bucket should contains at least current backup")
		return
	}

	// Send last backup date to pushgateway
	metric := r.Volume.MetricsHandler.NewMetric("bivac_lastBackup", "counter")
	metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(snapshots[len(snapshots)-1].Time.Unix(), 10),
		},
	)

	// Send oldest backup date to pushgateway
	metric = r.Volume.MetricsHandler.NewMetric("bivac_oldestBackup", "counter")
	metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(snapshots[0].Time.Unix(), 10),
		},
	)

	// Send snapshots count to pushgateway
	metric = r.Volume.MetricsHandler.NewMetric("bivac_backupCount", "gauge")
	metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(int64(len(snapshots)), 10),
		},
	)

	duration, err := util.GetDurationFromInterval(v.Config.RemoveOlderThan)
	if err != nil {
		return err
	}

	validSnapshots := 0
	now := time.Now()
	for _, snapshot := range snapshots {
		expiration := snapshot.Time.Add(duration)
		if now.Before(expiration) {
			validSnapshots++
		}
	}

	state, output, err := r.launchRestic(
		[]string{
			"-r",
			"%T",
			"forget",
			"--prune",
			"--keep-last",
			fmt.Sprintf("%d", validSnapshots),
		},
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to forget the snapshot: %v", err)
		return err
	}

	if state != 0 {
		err = fmt.Errorf("restic failed to forget old snapshots: %v", output)
		return err
	}
	return
}

// snapshots lists snapshots
func (r *ResticEngine) snapshots() (snapshots []Snapshot, err error) {
	_, output, err := r.launchRestic(
		[]string{
			"-r",
			"%T",
			"snapshots",
			"--json",
		},
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to check the backup: %v", err)
		return
	}

	if err := json.Unmarshal([]byte(output), &snapshots); err != nil {
		err = fmt.Errorf("failed to parse JSON output: %v", err)
		return snapshots, err
	}
	return
}

// sendBackupStatus creates a metric which represents the backup status. 0 == OK / 1 == KO
func (r *ResticEngine) sendBackupStatus(status int, volume string) {
	metric := r.Volume.MetricsHandler.NewMetric("bivac_backupExitCode", "gauge")
	err := metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": volume,
			},
			Value: strconv.Itoa(status),
		},
	)
	if err != nil {
		log.Errorf("failed to send metric: %v", err)
	}
}

// launchRestic starts a restic container with the given command
func (r *ResticEngine) launchRestic(cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	config := r.Orchestrator.GetHandler().Config
	image := config.Restic.Image

	// Disable cache to avoid volume issues with Kubernetes
	cmd = append(cmd, "--no-cache")

	env := map[string]string{
		"AWS_ACCESS_KEY_ID":      config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY":  config.AWS.SecretAccessKey,
		"OS_USERNAME":            config.Swift.Username,
		"OS_PASSWORD":            config.Swift.Password,
		"OS_AUTH_URL":            config.Swift.AuthURL,
		"OS_TENANT_NAME":         config.Swift.TenantName,
		"OS_REGION_NAME":         config.Swift.RegionName,
		"OS_USER_DOMAIN_NAME":    config.Swift.UserDomainName,
		"OS_PROJECT_NAME":        config.Swift.ProjectName,
		"OS_PROJECT_DOMAIN_NAME": config.Swift.ProjectDomainName,
		"RESTIC_PASSWORD":        config.Restic.Password,
	}

	for k, v := range config.ExtraEnv {
		env[k] = v
	}

	return r.Orchestrator.LaunchContainer(image, env, r.replaceArgs(cmd), volumes)
}
