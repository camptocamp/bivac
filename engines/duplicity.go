package engines

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	docker "github.com/docker/engine-api/client"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"golang.org/x/net/context"
)

// DuplicityEngine implements a backup engine with Duplicity
type DuplicityEngine struct {
	Config *config.Config
	Docker *docker.Client
	Volume *volume.Volume
}

// Constants
const cacheMount = "duplicity_cache:/root/.cache/duplicity"
const timeFormat = "Mon Jan 2 15:04:05 2006"

var fullBackupRx = regexp.MustCompile("Last full backup date: (.+)")
var chainEndTimeRx = regexp.MustCompile("Chain end time: (.+)")

// RemoveOld cleans up old backup data
func (d *DuplicityEngine) RemoveOld(v *volume.Volume) (metrics []string, err error) {
	_, _, err = d.LaunchDuplicity(
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
	util.CheckErr(err, "Failed to launch Duplicity: %v", "fatal")
	return
}

// Cleanup removes old index data from duplicity
func (d *DuplicityEngine) Cleanup(v *volume.Volume) (metrics []string, err error) {
	_, _, err = d.LaunchDuplicity(
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
	util.CheckErr(err, "Failed to launch Duplicity: %v", "fatal")
	return
}

// Verify checks that the backup is usable
func (d *DuplicityEngine) Verify(v *volume.Volume) (metrics []string, err error) {
	state, _, err := d.LaunchDuplicity(
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
	util.CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

	metric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"verifyExitCode\"} %v", v.Name, state)
	metrics = []string{
		metric,
	}
	return
}

// Status gets the latest backup date info from duplicity
func (d *DuplicityEngine) Status(v *volume.Volume) (metrics []string, err error) {
	_, stdout, err := d.LaunchDuplicity(
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
	util.CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

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
			util.CheckErr(err, "Failed to parse full backup date: %v", "error")

			if len(chainEndTime) > 0 {
				chainEndTimeDate, err = time.Parse(timeFormat, strings.TrimSpace(chainEndTime[1]))
				util.CheckErr(err, "Failed to parse chain end time date: %v", "error")
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

// LaunchDuplicity starts a duplicity container with given command and binds
func (d *DuplicityEngine) LaunchDuplicity(cmd []string, binds []string) (state int, stdout string, err error) {
	util.PullImage(d.Docker, d.Config.Duplicity.Image)
	util.CheckErr(err, "Failed to pull image: %v", "fatal")

	env := []string{
		"AWS_ACCESS_KEY_ID=" + d.Config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + d.Config.AWS.SecretAccessKey,
		"SWIFT_USERNAME=" + d.Config.Swift.Username,
		"SWIFT_PASSWORD=" + d.Config.Swift.Password,
		"SWIFT_AUTHURL=" + d.Config.Swift.AuthURL,
		"SWIFT_TENANTNAME=" + d.Config.Swift.TenantName,
		"SWIFT_REGIONNAME=" + d.Config.Swift.RegionName,
		"SWIFT_AUTHVERSION=2",
	}

	log.WithFields(log.Fields{
		"image":       d.Config.Duplicity.Image,
		"command":     strings.Join(cmd, " "),
		"environment": strings.Join(env, ", "),
		"binds":       strings.Join(binds, ", "),
	}).Debug("Creating container")

	container, err := d.Docker.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          env,
			Image:        d.Config.Duplicity.Image,
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          true,
		},
		&container.HostConfig{
			Binds: binds,
		}, nil, "",
	)
	util.CheckErr(err, "Failed to create container: %v", "fatal")
	defer util.RemoveContainer(d.Docker, container.ID)

	log.Debugf("Launching 'duplicity %v'...", strings.Join(cmd, " "))
	err = d.Docker.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	util.CheckErr(err, "Failed to start container: %v", "fatal")

	var exited bool

	for !exited {
		cont, err := d.Docker.ContainerInspect(context.Background(), container.ID)
		util.CheckErr(err, "Failed to inspect container: %v", "error")

		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := d.Docker.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Follow:     true,
	})
	util.CheckErr(err, "Failed to retrieve logs: %v", "error")

	defer body.Close()
	content, err := ioutil.ReadAll(body)
	util.CheckErr(err, "Failed to read logs from response: %v", "error")

	stdout = string(content)

	log.Debug(stdout)

	return
}

// GetName returns the engine name
func (*DuplicityEngine) GetName() string {
	return "Duplicity"
}

// Backup performs the backup of the passed volume
func (d *DuplicityEngine) Backup() (metrics []string, err error) {
	v := d.Volume
	vol := v.Volume
	log.WithFields(log.Fields{
		"volume":     vol.Name,
		"driver":     vol.Driver,
		"mountpoint": vol.Mountpoint,
	}).Info("Creating duplicity container")

	fullIfOlderThan, _ := util.GetVolumeLabel(vol, ".full_if_older_than")
	if fullIfOlderThan == "" {
		fullIfOlderThan = v.Config.Duplicity.FullIfOlderThan
	}

	removeOlderThan, _ := util.GetVolumeLabel(vol, ".remove_older_than")
	if removeOlderThan == "" {
		removeOlderThan = v.Config.Duplicity.RemoveOlderThan
	}

	pathSeparator := "/"
	if strings.HasPrefix(v.Config.Duplicity.TargetURL, "swift://") {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	// TODO
	//backupDir := p.GetBackupDir()
	backupDir := v.BackupDir
	hostname, _ := os.Hostname()
	v.Target = v.Config.Duplicity.TargetURL + pathSeparator + hostname + pathSeparator + vol.Name
	v.BackupDir = vol.Mountpoint + "/" + backupDir
	v.Mount = vol.Name + ":" + vol.Mountpoint + ":ro"

	var newMetrics []string

	newMetrics, err = d.DuplicityBackup(v)
	util.CheckErr(err, "Failed to backup volume "+vol.Name+" : %v", "fatal")
	metrics = append(metrics, newMetrics...)

	_, err = d.RemoveOld(v)
	util.CheckErr(err, "Failed to remove old backups for volume "+vol.Name+" : %v", "fatal")

	_, err = d.Cleanup(v)
	util.CheckErr(err, "Failed to cleanup extraneous duplicity files for volume "+vol.Name+" : %v", "fatal")

	noVerifyLbl, _ := util.GetVolumeLabel(vol, ".no_verify")
	noVerify := v.Config.NoVerify || (noVerifyLbl == "true")
	if noVerify {
		log.WithFields(log.Fields{
			"volume": vol.Name,
		}).Info("Skipping verification")
	} else {
		newMetrics, err = d.Verify(v)
		util.CheckErr(err, "Failed to verify backup for volume "+vol.Name+" : %v", "fatal")
		metrics = append(metrics, newMetrics...)
	}

	newMetrics, err = d.Status(v)
	util.CheckErr(err, "Failed to retrieve last backup info for volume "+vol.Name+" : %v", "fatal")
	metrics = append(metrics, newMetrics...)

	return
}

// DuplicityBackup performs the backup of a volume with duplicity
func (d *DuplicityEngine) DuplicityBackup(v *volume.Volume) (metrics []string, err error) {
	log.WithFields(log.Fields{
		"name":               v.Name,
		"backup_dir":         v.BackupDir,
		"full_if_older_than": v.FullIfOlderThan,
		"target":             v.Target,
		"mount":              v.Mount,
	}).Debug("Starting volume backup")

	// TODO
	// Init engine

	state, _, err := d.LaunchDuplicity(
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
	util.CheckErr(err, "Failed to launch Duplicity: %v", "fatal")

	metric := fmt.Sprintf("conplicity{volume=\"%v\",what=\"backupExitCode\"} %v", v.Name, state)
	metrics = []string{
		metric,
	}
	return
}
