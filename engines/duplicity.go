package engines

import (
	"fmt"
	"net/url"
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

// Backup performs the backup of the passed volume
func (d *DuplicityEngine) Backup() (err error) {
	vol := d.Volume
	log.WithFields(log.Fields{
		"volume":     vol.Name,
		"driver":     vol.Driver,
		"mountpoint": vol.Mountpoint,
	}).Info("Creating duplicity container")

	targetURL, err := url.Parse(vol.Config.TargetURL)
	if err != nil {
		err = fmt.Errorf("failed to parse target URL: %v", err)
		return
	}

	backupDir := vol.BackupDir
	c := d.Orchestrator.GetHandler()
	vol.Target = targetURL.String() + "/" + d.Orchestrator.GetPath(vol)
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
	v := d.Volume
	_, _, _, err = d.launchDuplicity(
		[]string{
			"remove-older-than", v.Config.RemoveOlderThan,
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--force",
			"--name", v.Name,
			v.Target,
		},
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
	v := d.Volume
	_, _, _, err = d.launchDuplicity(
		[]string{
			"cleanup",
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--force",
			"--extra-clean",
			"--name", v.Name,
			v.Target,
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
	state, _, _, err := d.launchDuplicity(
		[]string{
			"verify",
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--allow-source-mismatch",
			"--name", v.Name,
			v.Target,
			v.BackupDir,
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
	for i := 0; i < attempts; i++ {
		_, stdout, _, err = d.launchDuplicity(
			[]string{
				"collection-status",
				"--s3-use-new-style",
				"--ssh-options", "-oStrictHostKeyChecking=no",
				"--no-encryption",
				"--name", v.Name,
				v.Target,
			},
			[]*volume.Volume{
				v,
			},
		)
		if err != nil {
			err = fmt.Errorf("failed to launch duplicity: %v", err)
			return
		}
		// TODO Check where the following message is send : stdout or stderr
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
func (d *DuplicityEngine) launchDuplicity(cmd []string, volumes []*volume.Volume) (state int, stdout string, stderr string, err error) {
	config := d.Orchestrator.GetHandler().Config
	image := config.Duplicity.Image

	env := map[string]string{
		"AWS_ACCESS_KEY_ID":     config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY": config.AWS.SecretAccessKey,
		"SWIFT_USERNAME":        config.Swift.Username,
		"SWIFT_PASSWORD":        config.Swift.Password,
		"SWIFT_AUTHURL":         config.Swift.AuthURL,
		"SWIFT_TENANTNAME":      config.Swift.TenantName,
		"SWIFT_REGIONNAME":      config.Swift.RegionName,
		"SWIFT_AUTHVERSION":     "2",
	}

	for k, v := range config.ExtraEnv {
		env[k] = v
	}

	return d.Orchestrator.LaunchContainer(image, env, cmd, volumes)
}

// duplicityBackup performs the backup of a volume with duplicity
func (d *DuplicityEngine) duplicityBackup() (err error) {
	v := d.Volume
	log.WithFields(log.Fields{
		"name":               v.Name,
		"backup_dir":         v.BackupDir,
		"full_if_older_than": v.Config.Duplicity.FullIfOlderThan,
		"target":             v.Target,
		"mount":              v.Mount,
	}).Info("Starting volume backup")

	// TODO
	// Init engine

	state, _, _, err := d.launchDuplicity(
		[]string{
			"--full-if-older-than", v.Config.Duplicity.FullIfOlderThan,
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--allow-source-mismatch",
			"--name", v.Name,
			v.BackupDir,
			v.Target,
		},
		[]*volume.Volume{
			v,
		},
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
