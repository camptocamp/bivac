package engines

import (
	"fmt"
	"io/ioutil"
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
func (r *RCloneEngine) Backup() (metrics []string, err error) {
	v := r.Volume
	target := r.Handler.Config.RClone.TargetURL + "/" + r.Handler.Hostname + "/" + v.Name
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
	)
	util.CheckErr(err, "Failed to launch RClone: %v", "fatal")

	if state != 0 {
		err = fmt.Errorf("RClone exited with state %v", state)
	}

	return
}

// launchRClone starts an rclone container with a given command and binds
func (r *RCloneEngine) launchRClone(cmd []string, binds []string) (state int, stdout string, err error) {
	util.PullImage(r.Handler.Client, r.Handler.Config.RClone.Image)
	util.CheckErr(err, "Failed to pull image: %v", "fatal")

	env := []string{
		"AWS_ACCESS_KEY_ID=" + r.Handler.Config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + r.Handler.Config.AWS.SecretAccessKey,
		"OS_USERNAME=" + r.Handler.Config.Swift.Username,
		"OS_PASSWORD=" + r.Handler.Config.Swift.Password,
		"OS_AUTH_URL=" + r.Handler.Config.Swift.AuthURL,
		"OS_TENANT_NAME=" + r.Handler.Config.Swift.TenantName,
		"OS_REGION_NAME=" + r.Handler.Config.Swift.RegionName,
	}

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
	util.CheckErr(err, "Failed to create container: %v", "fatal")
	defer util.RemoveContainer(r.Handler.Client, container.ID)

	log.Debugf("Launching 'rclone %v'...", strings.Join(cmd, " "))
	err = r.Handler.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	util.CheckErr(err, "Failed to start container: %v", "fatal")

	var exited bool

	for !exited {
		cont, err := r.Handler.ContainerInspect(context.Background(), container.ID)
		util.CheckErr(err, "Failed to inspect container: %v", "error")

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
	util.CheckErr(err, "Failed to retrieve logs: %v", "error")

	defer body.Close()
	content, err := ioutil.ReadAll(body)
	util.CheckErr(err, "Failed to read logs from response: %v", "error")

	stdout = string(content)

	log.Debug(stdout)

	return
}
