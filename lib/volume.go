package conplicity

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Handler is an interface providing a mean of launching Duplicity
type Handler interface {
	LaunchDuplicity([]string, []string) (int, string, error)
}

// Volume provides backup methods for a single Docker volume
type Volume struct {
	Name            string
	Target          string
	BackupDir       string
	Mount           string
	FullIfOlderThan string
	RemoveOlderThan string
	Client          Handler
}

// Constants
const cacheMount = "duplicity_cache:/root/.cache/duplicity"
const timeFormat = "Mon Jan 2 15:04:05 2006"

var fullBackupRx = regexp.MustCompile("Last full backup date: (.+)")
var chainEndTimeRx = regexp.MustCompile("Chain end time: (.+)")

// Backup performs the backup of a volume with duplicity
func (v *Volume) Backup() (metrics []string, err error) {
	log.WithFields(log.Fields{
		"name":               v.Name,
		"backup_dir":         v.BackupDir,
		"full_if_older_than": v.FullIfOlderThan,
		"target":             v.Target,
		"mount":              v.Mount,
	}).Debug("Starting volume backup")

	state, _, err := v.Client.LaunchDuplicity(
		[]string{
			"--full-if-older-than", v.FullIfOlderThan,
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--allow-source-mismatch",
			"--name", v.Name,
			v.BackupDir,
			v.Target,
		},
		[]string{
			v.Mount,
			cacheMount,
		},
	)
	CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

	metric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"backupExitCode\"} %v", v.Name, state)
	metrics = []string{
		metric,
	}
	return
}

// RemoveOld cleans up old backup data from duplicity
func (v *Volume) RemoveOld() (metrics []string, err error) {
	_, _, err = v.Client.LaunchDuplicity(
		[]string{
			"remove-older-than", v.RemoveOlderThan,
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--force",
			"--name", v.Name,
			v.Target,
		},
		[]string{
			cacheMount,
		},
	)
	CheckErr(err, "Failed to launch Duplicity: %v", "fatal")
	return
}

// Cleanup removes old index data from duplicity
func (v *Volume) Cleanup() (metrics []string, err error) {
	_, _, err = v.Client.LaunchDuplicity(
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
		[]string{
			cacheMount,
		},
	)
	CheckErr(err, "Failed to launch Duplicity: %v", "fatal")
	return
}

// Verify checks that the backup is usable
func (v *Volume) Verify() (metrics []string, err error) {
	state, _, err := v.Client.LaunchDuplicity(
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
		[]string{
			v.Mount,
			cacheMount,
		},
	)
	CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

	metric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"verifyExitCode\"} %v", v.Name, state)
	metrics = []string{
		metric,
	}
	return
}

// Status gets the latest backup date info from duplicity
func (v *Volume) Status() (metrics []string, err error) {
	_, stdout, err := v.Client.LaunchDuplicity(
		[]string{
			"collection-status",
			"--s3-use-new-style",
			"--ssh-options", "-oStrictHostKeyChecking=no",
			"--no-encryption",
			"--name", v.Name,
			v.Target,
		},
		[]string{
			v.Mount,
			cacheMount,
		},
	)
	CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

	fullBackup := fullBackupRx.FindStringSubmatch(stdout)
	var fullBackupDate time.Time
	chainEndTime := chainEndTimeRx.FindStringSubmatch(stdout)
	var chainEndTimeDate time.Time

	if len(fullBackup) > 0 {
		if strings.TrimSpace(fullBackup[1]) == "none" {
			fullBackupDate = time.Unix(0, 0)
			chainEndTimeDate = time.Unix(0, 0)
		} else {
			fullBackupDate, err = time.Parse(timeFormat, strings.TrimSpace(fullBackup[1]))
			CheckErr(err, "Failed to parse full backup date: %v", "error")

			if len(chainEndTime) > 0 {
				chainEndTimeDate, err = time.Parse(timeFormat, strings.TrimSpace(chainEndTime[1]))
				CheckErr(err, "Failed to parse chain end time date: %v", "error")
			} else {
				errMsg := fmt.Sprintf("Failed to parse Duplicity output for chain end time of %v", v.Name)
				err = errors.New(errMsg)
				return
			}

		}
	} else {
		errMsg := fmt.Sprintf("Failed to parse Duplicity output for last full backup date of %v", v.Name)
		err = errors.New(errMsg)
		return
	}

	lastBackupMetric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"lastBackup\"} %v", v.Name, chainEndTimeDate.Unix())

	lastFullBackupMetric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"lastFullBackup\"} %v", v.Name, fullBackupDate.Unix())

	metrics = []string{
		lastBackupMetric,
		lastFullBackupMetric,
	}

	return
}
