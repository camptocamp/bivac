package engines

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/metrics"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"golang.org/x/net/context"
)

// DuplicityEngine implements a backup engine with Duplicity
type DuplicityEngine struct {
	Handler *handler.Conplicity
	Volume  *volume.Volume
}

// Constants
const cacheMount = "duplicity_cache:/root/.cache/duplicity"
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

	pathSeparator := "/"
	if targetURL.Scheme == "swift" {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	backupDir := vol.BackupDir
	vol.Target = targetURL.String() + pathSeparator + d.Handler.Hostname + pathSeparator + vol.Name
	vol.BackupDir = vol.Mountpoint + "/" + backupDir
	vol.Mount = vol.Name + ":" + vol.Mountpoint + ":ro"

	err = d.duplicityBackup()
	if err != nil {
		err = fmt.Errorf("failed to backup volume with duplicity: %v", err)
		return
	}

	err = d.removeOld()
	if err != nil {
		err = fmt.Errorf("failed to remove old backups: %v", err)
		return
	}

	err = d.cleanup()
	if err != nil {
		err = fmt.Errorf("failed to cleanup extraneous duplicity files: %v", err)
		return
	}

	if vol.Config.NoVerify {
		log.WithFields(log.Fields{
			"volume": vol.Name,
		}).Info("Skipping verification")
	} else {
		err = d.verify()
		if err != nil {
			err = fmt.Errorf("failed to verify backup: %v", err)
			return
		}
	}

	err = d.status()
	if err != nil {
		err = fmt.Errorf("failed to retrieve last backup info: %v", err)
		return
	}

	return
}

// removeOld cleans up old backup data
func (d *DuplicityEngine) removeOld() (err error) {
	v := d.Volume
	_, _, err = d.launchDuplicity(
		[]string{
			"remove-older-than", v.Config.Duplicity.RemoveOlderThan,
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
	if err != nil {
		err = fmt.Errorf("failed to launch Duplicity: %v", err)
		return
	}
	return
}

// cleanup removes old index data from duplicity
func (d *DuplicityEngine) cleanup() (err error) {
	v := d.Volume
	_, _, err = d.launchDuplicity(
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
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
	}
	return
}

// verify checks that the backup is usable
func (d *DuplicityEngine) verify() (err error) {
	v := d.Volume
	state, _, err := d.launchDuplicity(
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
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
		return
	}

	metric := d.Handler.MetricsHandler.NewMetric("conplicity_verifyExitCode", "gauge")
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

// status gets the latest backup date info from duplicity
func (d *DuplicityEngine) status() (err error) {
	v := d.Volume
	_, stdout, err := d.launchDuplicity(
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
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
		return
	}

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
			if err != nil {
				err = fmt.Errorf("failed to parse full backup data: %v", err)
				return
			}

			if len(chainEndTime) > 0 {
				chainEndTimeDate, err = time.Parse(timeFormat, strings.TrimSpace(chainEndTime[1]))
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

	lastBackupMetric := d.Handler.MetricsHandler.NewMetric("conplicity_lastBackup", "counter")
	lastBackupMetric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(chainEndTimeDate.Unix(), 10),
		},
	)

	lastFullBackupMetric := d.Handler.MetricsHandler.NewMetric("conplicity_lastFullBackup", "counter")
	lastFullBackupMetric.UpdateEvent(
		&metrics.Event{
			Labels: map[string]string{
				"volume": v.Name,
			},
			Value: strconv.FormatInt(fullBackupDate.Unix(), 10),
		},
	)

	return
}

// launchDuplicity starts a duplicity container with given command and binds
func (d *DuplicityEngine) launchDuplicity(cmd []string, binds []string) (state int, stdout string, err error) {
	err = util.PullImage(d.Handler.Client, d.Handler.Config.Duplicity.Image)
	if err != nil {
		err = fmt.Errorf("failed to pull image: %v", err)
		return
	}

	env := []string{
		"AWS_ACCESS_KEY_ID=" + d.Handler.Config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + d.Handler.Config.AWS.SecretAccessKey,
		"SWIFT_USERNAME=" + d.Handler.Config.Swift.Username,
		"SWIFT_PASSWORD=" + d.Handler.Config.Swift.Password,
		"SWIFT_AUTHURL=" + d.Handler.Config.Swift.AuthURL,
		"SWIFT_TENANTNAME=" + d.Handler.Config.Swift.TenantName,
		"SWIFT_REGIONNAME=" + d.Handler.Config.Swift.RegionName,
		"SWIFT_AUTHVERSION=2",
	}

	log.WithFields(log.Fields{
		"image":       d.Handler.Config.Duplicity.Image,
		"command":     strings.Join(cmd, " "),
		"environment": strings.Join(env, ", "),
		"binds":       strings.Join(binds, ", "),
	}).Debug("Creating container")

	container, err := d.Handler.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          env,
			Image:        d.Handler.Config.Duplicity.Image,
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
	if err != nil {
		err = fmt.Errorf("failed to create container: %v", err)
		return
	}
	defer util.RemoveContainer(d.Handler.Client, container.ID)

	log.Debugf("Launching 'duplicity %v'...", strings.Join(cmd, " "))
	err = d.Handler.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		err = fmt.Errorf("failed to start container: %v", err)
	}

	var exited bool

	for !exited {
		var cont types.ContainerJSON
		cont, err = d.Handler.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %v", err)
			return
		}

		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := d.Handler.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Follow:     true,
	})
	if err != nil {
		err = fmt.Errorf("failed to retrieve logs: %v", err)
		return
	}

	defer body.Close()
	content, err := ioutil.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("failed to read logs from response: %v", err)
		return
	}

	stdout = string(content)
	log.Debug(stdout)

	return
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
	}).Debug("Starting volume backup")

	// TODO
	// Init engine

	state, _, err := d.launchDuplicity(
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
		[]string{
			v.Mount,
			cacheMount,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch duplicity: %v", err)
		return
	}

	metric := d.Handler.MetricsHandler.NewMetric("conplicity_backupExitCode", "gauge")
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
