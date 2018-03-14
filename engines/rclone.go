package engines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/camptocamp/conplicity/orchestrators"
	"github.com/camptocamp/conplicity/volume"
)

// RCloneEngine implements a backup engine with RClone
type RCloneEngine struct {
	Orchestrator orchestrators.Orchestrator
	Volume       *volume.Volume
}

// GetName returns the engine name
func (*RCloneEngine) GetName() string {
	return "RClone"
}

// Backup performs the backup of the passed volume
func (r *RCloneEngine) Backup() (err error) {
	v := r.Volume

	targetURL, err := url.Parse(v.Config.TargetURL)
	if err != nil {
		err = fmt.Errorf("failed to parse target URL: %v", err)
		return
	}

	// Format targetURL for RClone
	extraEnv := formatURL(targetURL)
	c := r.Orchestrator.GetHandler()

	target := targetURL.String() + "/" + c.Hostname + "/" + v.Name
	backupDir := v.Mountpoint + "/" + v.BackupDir

	state, _, err := r.launchRClone(
		[]string{
			"sync",
			backupDir,
			target,
		},
		[]string{
			v.Name + ":" + v.Mountpoint + ":ro",
		},
		extraEnv,
	)
	if err != nil {
		err = fmt.Errorf("failed to launch RClone: %v", err)
	}
	if state != 0 {
		err = fmt.Errorf("RClone exited with state %v", state)
	}
	return
}

func formatURL(u *url.URL) (env []string) {
	// We have no way but to assume fqdns contain "."
	// which is arguable very ugly
	if strings.Contains(u.Host, ".") && strings.HasPrefix(u.Scheme, "s3") {
		u.Opaque = strings.TrimPrefix(u.Path, "/")
		env = append(env, "AWS_ENDPOINT="+u.Host)
	} else {
		u.Opaque = strings.TrimPrefix(u.Host+u.Path, "/")
	}

	plusIndex := strings.Index(u.Scheme, "+")
	if plusIndex >= 0 {
		u.Scheme = u.Scheme[0:plusIndex]
	}
	return
}

// launchRClone starts an rclone container with a given command and binds
func (r *RCloneEngine) launchRClone(cmd, binds, extraEnv []string) (state int, stdout string, err error) {
	config := r.Orchestrator.GetHandler().Config
	image := config.RClone.Image

	env := []string{
		"AWS_ACCESS_KEY_ID=" + config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + config.AWS.SecretAccessKey,
		"OS_USERNAME=" + config.Swift.Username,
		"OS_PASSWORD=" + config.Swift.Password,
		"OS_AUTH_URL=" + config.Swift.AuthURL,
		"OS_TENANT_NAME=" + config.Swift.TenantName,
		"OS_REGION_NAME=" + config.Swift.RegionName,
	}
	env = append(env, extraEnv...)

	return r.Orchestrator.LaunchContainer(image, env, cmd, binds)
}
