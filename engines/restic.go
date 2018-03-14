package engines

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/camptocamp/conplicity/metrics"
	"github.com/camptocamp/conplicity/orchestrators"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
)

// ResticEngine implements a backup engine with Restic
type ResticEngine struct {
	Orchestrator orchestrators.Orchestrator
	Volume       *volume.Volume
}

// GetName returns the engine name
func (*ResticEngine) GetName() string {
	return "Restic"
}

// Backup performs the backup of the passed volume
func (r *ResticEngine) Backup() (err error) {

	v := r.Volume

	targetURL, err := url.Parse(v.Config.TargetURL)
	if err != nil {
		err = fmt.Errorf("failed to parse target URL: %v", err)
		return
	}

	v.Target = targetURL.String()
	v.BackupDir = v.Mountpoint + "/" + v.BackupDir
	v.Mount = v.Name + ":" + v.Mountpoint + ":ro"

	err = util.Retry(3, r.init)
	if err != nil {
		err = fmt.Errorf("failed to create a secure bucket: %v", err)
		return
	}

	err = util.Retry(3, r.resticBackup)
	if err != nil {
		err = fmt.Errorf("failed to backup the volume: %v", err)
		return
	}

	c := r.Orchestrator.GetHandler()
	if _, err := c.IsCheckScheduled(v); err == nil {
		err = util.Retry(3, r.verify)
		if err != nil {
			err = fmt.Errorf("failed to verify backup: %v", err)
			return err
		}
	}
	return
}

// init initialize a secure bucket
func (r *ResticEngine) init() (err error) {
	v := r.Volume
	state, stdout, err := r.launchRestic(
		[]string{
			"-r",
			v.Target,
			"init",
		},
		[]string{
			v.Mount,
		},
	)
	if strings.Contains(stdout, "already initialized") {
		err = nil
		return
	}
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to initialize the repository: %v", err)
		return
	}
	if state != 0 {
		err = fmt.Errorf("Restic existed with state %v while initializing repository", state)
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
			v.Target,
			"backup",
			v.BackupDir,
		},
		[]string{
			v.Mount,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to backup the volume: %v", err)
	}
	if state != 0 {
		err = fmt.Errorf("Restic exited with state %v while backuping the volume", state)
	}
	return
}

// verify checks that the backup is usable
func (r *ResticEngine) verify() (err error) {
	v := r.Volume
	state, _, err := r.launchRestic(
		[]string{
			"-r",
			v.Target,
			"check",
		},
		[]string{
			v.Mount,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch Restic to check the backup: %v", err)
		return
	}
	if state == 0 {
		now := time.Now().Local()
		os.Chtimes(v.Mountpoint+"/.conplicity_last_check", now, now)
	} else {
		err = fmt.Errorf("Restic exited with state %v while checking the backup", state)
	}

	metric := r.Volume.MetricsHandler.NewMetric("conplicity_verifyExitCode", "gauge")
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

// launchRestic starts a restic container with the given command and binds
func (r *ResticEngine) launchRestic(cmd, binds []string) (state int, stdout string, err error) {
	config := r.Orchestrator.GetHandler().Config
	image := config.Restic.Image

	env := []string{
		"AWS_ACCESS_KEY_ID=" + config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + config.AWS.SecretAccessKey,
		"OS_USERNAME=" + config.Swift.Username,
		"OS_PASSWORD=" + config.Swift.Password,
		"OS_AUTH_URL=" + config.Swift.AuthURL,
		"OS_TENANT_NAME=" + config.Swift.TenantName,
		"OS_REGION_NAME=" + config.Swift.RegionName,
		"RESTIC_PASSWORD=" + config.Restic.Password,
	}

	return r.Orchestrator.LaunchContainer(image, env, cmd, binds)
}