package engines

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/volume"
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

	target := targetURL.String() + "/" + r.Orchestrator.GetPath(v)
	backupDir := v.Mountpoint + "/" + v.BackupDir

	state, _, _, err := r.launchRClone(
		[]string{
			"sync",
			backupDir,
			target,
		},
		extraEnv,
		[]*volume.Volume{
			v,
		},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch RClone: %v", err)
	}
	if state != 0 {
		err = fmt.Errorf("RClone exited with state %v", state)
	}
	return
}

func formatURL(u *url.URL) (env map[string]string) {
	// We have no way but to assume fqdns contain "."
	// which is arguable very ugly
	env = make(map[string]string)
	if strings.Contains(u.Host, ".") && strings.HasPrefix(u.Scheme, "s3") {
		u.Opaque = strings.TrimPrefix(u.Path, "/")
		env["AWS_ENDPOINT"] = u.Host
	} else {
		u.Opaque = strings.TrimPrefix(u.Host+u.Path, "/")
	}

	plusIndex := strings.Index(u.Scheme, "+")
	if plusIndex >= 0 {
		u.Scheme = u.Scheme[0:plusIndex]
	}
	return
}

// launchRClone starts an rclone container with a given command
func (r *RCloneEngine) launchRClone(cmd []string, extraEnv map[string]string, volumes []*volume.Volume) (state int, stdout string, stderr string, err error) {
	config := r.Orchestrator.GetHandler().Config
	image := config.RClone.Image

	env := map[string]string{
		"AWS_ACCESS_KEY_ID":                 config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY":             config.AWS.SecretAccessKey,
		"RCLONE_CONFIG_SWIFT_TYPE":          "swift",
		"RCLONE_CONFIG_SWIFT_AUTH":          config.Swift.AuthURL,
		"RCLONE_CONFIG_SWIFT_USER":          config.Swift.Username,
		"RCLONE_CONFIG_SWIFT_KEY":           config.Swift.Password,
		"RCLONE_CONFIG_SWIFT_REGION":        config.Swift.RegionName,
		"RCLONE_CONFIG_SWIFT_TENANT":        config.Swift.TenantName,
		"RCLONE_CONFIG_SWIFT_DOMAIN":        config.Swift.UserDomainName,
		"RCLONE_CONFIG_SWIFT_TENANT_DOMAIN": config.Swift.ProjectDomainName,
	}
	for en, ev := range extraEnv {
		env[en] = ev
	}
	for k, v := range config.ExtraEnv {
		env[k] = v
	}

	return r.Orchestrator.LaunchContainer(image, env, cmd, volumes)
}
