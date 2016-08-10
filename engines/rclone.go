package engines

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
)

// RCloneEngine implements a backup engine with RClone
type RCloneEngine struct {
	Handler *handler.Conplicity
	Volume  *volume.Volume
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

	target := targetURL.String() + "/" + r.Handler.Hostname + "/" + v.Name
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
	err = util.PullImage(r.Handler.Client, r.Handler.Config.RClone.Image)
	if err != nil {
		err = fmt.Errorf("failed to pull image: %v", err)
		return
	}

	env := []string{
		"AWS_ACCESS_KEY_ID=" + r.Handler.Config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + r.Handler.Config.AWS.SecretAccessKey,
		"OS_USERNAME=" + r.Handler.Config.Swift.Username,
		"OS_PASSWORD=" + r.Handler.Config.Swift.Password,
		"OS_AUTH_URL=" + r.Handler.Config.Swift.AuthURL,
		"OS_TENANT_NAME=" + r.Handler.Config.Swift.TenantName,
		"OS_REGION_NAME=" + r.Handler.Config.Swift.RegionName,
	}
	env = append(env, extraEnv...)

	log.WithFields(log.Fields{
		"image":       r.Handler.Config.RClone.Image,
		"command":     strings.Join(cmd, " "),
		"environment": strings.Join(env, ", "),
		"binds":       strings.Join(binds, ", "),
	}).Debug("Creating container")

	container, err := r.Handler.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          env,
			Image:        r.Handler.Config.RClone.Image,
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
	defer util.RemoveContainer(r.Handler.Client, container.ID)

	log.Debugf("Launching 'rclone %v'...", strings.Join(cmd, " "))
	err = r.Handler.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		err = fmt.Errorf("failed to start container: %v", err)
		return
	}

	var exited bool

	for !exited {
		var cont types.ContainerJSON
		cont, err = r.Handler.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %v", err)
			return
		}
		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := r.Handler.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
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
