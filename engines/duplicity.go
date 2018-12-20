package engines

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/metrics"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/util"
	"github.com/camptocamp/bivac/volume"
)

// DuplicityEngine implements a backup engine with Duplicity
type DuplicityEngine struct {
	Orchestrator orchestrators.Orchestrator
	Volume       *volume.Volume
}

// Constants
const timeFormat = "Mon Jan 2 15:04:05 2006"

var fullBackupRx = regexp.MustCompile("Last full backup date: (.+)")
var chainEndTimeRx = regexp.MustCompile("Chain end time: (.+)")

// GetName returns the engine name
func (*DuplicityEngine) GetName() string {
	return "Duplicity"
}

// replaceArgs replace arguments with their values
func (d *DuplicityEngine) replaceArgs(args []string) (newArgs []string) {
	log.Debugf("Replacing args, Input: %v", args)
	for _, arg := range args {
		arg = strings.Replace(arg, "%B", d.Volume.Config.TargetURL, -1)
		arg = strings.Replace(arg, "%D", d.Volume.BackupDir, -1)
		arg = strings.Replace(arg, "%H", d.Volume.Hostname, -1)
		arg = strings.Replace(arg, "%N", d.Volume.Namespace, -1)
		arg = strings.Replace(arg, "%P", d.Orchestrator.GetPath(d.Volume), -1)
		arg = strings.Replace(arg, "%V", d.Volume.Name, -1)
		newArgs = append(newArgs, arg)
	}
	log.Debugf("Replacing args, Output: %v", newArgs)
	return
}

// Backup performs the backup of the passed volume
func (d *DuplicityEngine) Backup() (err error) {
	vol := d.Volume
	log.WithFields(log.Fields{
		"volume":     vol.Name,
		"driver":     vol.Driver,
		"mountpoint": vol.Mountpoint,
	}).Info("Creating duplicity container")

	backupDir := vol.BackupDir
	c := d.Orchestrator.GetHandler()
	vol.BackupDir = vol.Mountpoint + "/" + backupDir
	vol.Mount = vol.Name + ":" + vol.Mountpoint + ":ro"

	err = util.Retry(3, d.duplicityBackup)
	if err != nil {
		err = fmt.Errorf("failed to backup volume with duplicity: %v", err)
		return
	}

	err = util.Retry(3, d.removeOld)
	if err != nil {
		err = fmt.Errorf("failed to remove old backups: %v", err)
		return
	}

	err = util.Retry(3, d.cleanup)
	if err != nil {
		err = fmt.Errorf("failed to cleanup extraneous duplicity files: %v", err)
		return
	}

	if c.IsCheckScheduled(vol) {
		err = util.Retry(3, d.verify)
		if err != nil {
			err = fmt.Errorf("failed to verify backup: %v", err)
			return err
		}
	} else {
		return err
	}

	err = util.Retry(3, d.status)
	if err != nil {
		err = fmt.Errorf("failed to retrieve last backup info: %v", err)
		return
	}

	return
}

// removeOld cleans up old backup data
func (d *DuplicityEngine) removeOld() (err error) {
	config := d.Orchestrator.GetHandler().Config
	_, _, err = d.launchDuplicity(
		append(
			[]string{"remove-older-than"},
			strings.Split(config.Duplicity.RemoveOlderThanArgs, " ")...,
		),
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Duplicity: %v", err)
		return
	}
	return
}

// cleanup removes old index data from duplicity
func (d *DuplicityEngine) cleanup() (err error) {
	config := d.Orchestrator.GetHandler().Config
	_, _, err = d.launchDuplicity(
		[]string{
			"cleanup",
			"--force",
			"--extra-clean",
			"--name", "%V",
			config.Duplicity.BackupArgs,
		},
		[]*volume.Volume{},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
	}
	return
}

// verify checks that the backup is usable
func (d *DuplicityEngine) verify() (err error) {
	v := d.Volume
	config := d.Orchestrator.GetHandler().Config
	state, _, err := d.launchDuplicity(
		[]string{
			"verify",
			"--allow-source-mismatch",
			"--name", "%V",
			config.Duplicity.BackupArgs,
			"%D",
		},
		[]*volume.Volume{
			v,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
		return
	}

	if state == 0 {
		now := time.Now().Local()
		os.Chtimes(v.Mountpoint+"/.bivac_last_check", now, now)
	} else {
		err = fmt.Errorf("Duplicity exited with state %v while checking the backup", state)
	}

	metric := d.Volume.MetricsHandler.NewMetric("bivac_verifyExitCode", "gauge")
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

// status gets the latest backup date info from duplicity
func (d *DuplicityEngine) status() (err error) {
	var stdout string
	collectionComplete := false
	attempts := 3
	v := d.Volume
	config := d.Orchestrator.GetHandler().Config
	for i := 0; i < attempts; i++ {
		_, stdout, err = d.launchDuplicity(
			[]string{
				"collection-status",
				"--name", "%V",
				config.Duplicity.BackupArgs,
			},
			[]*volume.Volume{
				v,
			},
		)
		if err != nil {
			err = fmt.Errorf("failed to launch duplicity: %v", err)
			return
		}
		if strings.Contains(stdout, "No orphaned or incomplete backup sets found.") {
			collectionComplete = true
			break
		} else {
			log.Debug("No end string, found the collection-status output may be wrong, retrying ...")
		}
	}

	if !collectionComplete {
		err = fmt.Errorf("failed to retrieve full output from collection-status after %v attempts", attempts)
		return
	}

	fullBackup := fullBackupRx.FindStringSubmatch(stdout)
	var fullBackupDate time.Time
	var chainEndTimeDate time.Time

	if len(fullBackup) > 0 {
		chainEndTime := chainEndTimeRx.FindAllStringSubmatch(stdout, -1)
		if strings.TrimSpace(fullBackup[1]) == "none" {
			fullBackupDate = time.Unix(0, 0)
			chainEndTimeDate = time.Unix(0, 0)
		} else {
			fullBackupDate, err = time.Parse(timeFormat, strings.TrimSpace(fullBackup[1]))
			if err != nil {
				err = fmt.Errorf("failed to parse full backup data: %v", err)
				return
			}

			if len(chainEndTime) > 0 {
				chainEndTimeDate, err = time.Parse(timeFormat, strings.TrimSpace(chainEndTime[len(chainEndTime)-1][1]))
				if err != nil {
					err = fmt.Errorf("failed to parse chain end time date: %v", err)
					return
				}
			} else {
				err = fmt.Errorf("failed to parse Duplicity output for chain end time of %v", v.Name)
				return
			}

		}
	} else {
		err = fmt.Errorf("failed to parse Duplicity output for last full backup date of %v", v.Name)
		return
	}

	lastBackupMetric := d.Volume.MetricsHandler.NewMetric("bivac_lastBackup", "counter")
	lastBackupMetric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{},
			Value:  strconv.FormatInt(chainEndTimeDate.Unix(), 10),
		},
	)

	lastFullBackupMetric := d.Volume.MetricsHandler.NewMetric("bivac_lastFullBackup", "counter")
	lastFullBackupMetric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{},
			Value:  strconv.FormatInt(fullBackupDate.Unix(), 10),
		},
	)

	return
}

// launchDuplicity starts a duplicity container with given command
func (d *DuplicityEngine) launchDuplicity(cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	config := d.Orchestrator.GetHandler().Config
	image := config.Duplicity.Image

	return d.Orchestrator.LaunchContainer(image, d.replaceArgs(append(cmd, strings.Split(config.Duplicity.CommonArgs, " ")...)), volumes)
}

// duplicityBackup performs the backup of a volume with duplicity
func (d *DuplicityEngine) duplicityBackup() (err error) {
	config := d.Orchestrator.GetHandler().Config
	v := d.Volume
	log.WithFields(log.Fields{
		"name":       v.Name,
		"backup_dir": v.BackupDir,
		"mount":      v.Mount,
	}).Info("Starting volume backup")

	// TODO
	// Init engine

	state, _, err := d.launchDuplicity(
		strings.Split(config.Duplicity.BackupArgs, " "),
		[]*volume.Volume{v},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
		return
	}

	metric := d.Volume.MetricsHandler.NewMetric("bivac_backupExitCode", "gauge")
	metric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{},
			Value:  strconv.Itoa(state),
		},
	)
	return
}
