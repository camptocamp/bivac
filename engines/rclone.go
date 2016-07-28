package engines

import (
	"io/ioutil"
	"strings"

	"golang.org/x/net/context"

	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/vmware/harbor/utils/log"
)

// RCloneEngine implements a backup engine with RClone
type RCloneEngine struct {
	Config *config.Config
	Docker *docker.Client
	Volume *volume.Volume
}

// GetName returns the engine name
func (*RCloneEngine) GetName() string {
	return "RClone"
}

// Backup performs the backup of the passed volume
func (r *RCloneEngine) Backup() (metrics []string, err error) {
	vol := r.Volume

	_, _, err = r.launchRClone(
		[]string{
			"sync",
			"/data",
			"swift:backup",
		},
		[]string{
			vol.Name + ":" + vol.Mountpoint + ":ro",
		},
	)
	util.CheckErr(err, "Failed to launch RClone: %v", "fatal")

	return
}

// launchRClone starts an rclone container with a given command and binds
func (r *RCloneEngine) launchRClone(cmd []string, binds []string) (state int, stdout string, err error) {
	util.PullImage(r.Docker, r.Config.RClone.Image)
	util.CheckErr(err, "Failed to pull image: %v", "fatal")

	env := []string{
		"AWS_ACCESS_KEY_ID=" + r.Config.AWS.AccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + r.Config.AWS.SecretAccessKey,
		"OS_USERNAME=" + r.Config.Swift.Username,
		"OS_PASSWORD=" + r.Config.Swift.Password,
		"OS_AUTH_URL=" + r.Config.Swift.AuthURL,
		"OS_TENANT_NAME=" + r.Config.Swift.TenantName,
		"OS_REGION_NAME=" + r.Config.Swift.RegionName,
	}

	container, err := r.Docker.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          env,
			Image:        r.Config.RClone.Image,
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
	defer util.RemoveContainer(r.Docker, container.ID)

	log.Debugf("Launching 'rclone %v'...", strings.Join(cmd, " "))
	err = r.Docker.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	util.CheckErr(err, "Failed to start container: %v", "fatal")

	var exited bool

	for !exited {
		cont, err := r.Docker.ContainerInspect(context.Background(), container.ID)
		util.CheckErr(err, "Failed to inspect container: %v", "error")

		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := r.Docker.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
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
