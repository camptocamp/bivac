package orchestrators

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/context"

	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"

	log "github.com/Sirupsen/logrus"
)

// DockerOrchestrator implements a container orchestrator for Docker
type DockerOrchestrator struct {
	handler *handler.Conplicity
}

// GetName returns the orchestrator name
func (*DockerOrchestrator) GetName() string {
	return "Docker"
}

// GetHandler returns the Orchestrator's handler
func (o *DockerOrchestrator) GetHandler() *handler.Conplicity {
	return o.handler
}

// GetVolumes returns the Docker volumes, inspected and filtered
func (o *DockerOrchestrator) GetVolumes() (volumes []*volume.Volume, err error) {
	c := o.handler
	vols, err := c.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		err = fmt.Errorf("Failed to list Docker volumes: %v", err)
		return
	}
	for _, vol := range vols.Volumes {
		var voll types.Volume
		voll, err = c.VolumeInspect(context.Background(), vol.Name)
		if err != nil {
			err = fmt.Errorf("Failed to inspect volume %s: %v", vol.Name, err)
			return
		}
		v := volume.NewVolume(&voll, c.Config, c.Hostname)
		if b, r, s := o.blacklistedVolume(v); b {
			log.WithFields(log.Fields{
				"volume": vol.Name,
				"reason": r,
				"source": s,
			}).Info("Ignoring volume")
			continue
		}
		volumes = append(volumes, v)
	}
	return
}

// LaunchContainer starts a container using the Docker orchestrator
func (o *DockerOrchestrator) LaunchContainer(image string, env []string, cmd []string, binds []string) (state int, stdout string, err error) {
	err = util.PullImage(o.handler.Client, image)
	if err != nil {
		err = fmt.Errorf("failed to pull image: %v", err)
		return
	}

	log.WithFields(log.Fields{
		"image":       image,
		"command":     strings.Join(cmd, " "),
		"environment": strings.Join(env, ", "),
		"binds":       strings.Join(binds, ", "),
	}).Debug("Creating container")

	container, err := o.handler.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          env,
			Image:        image,
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
	defer util.RemoveContainer(o.handler.Client, container.ID)

	log.Debugf("Launching with '%v'...", strings.Join(cmd, " "))
	err = o.handler.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		err = fmt.Errorf("failed to start container: %v", err)
	}

	var exited bool

	for !exited {
		var cont types.ContainerJSON
		cont, err = o.handler.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %v", err)
			return
		}

		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := o.handler.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
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

func (o *DockerOrchestrator) blacklistedVolume(vol *volume.Volume) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	list := o.handler.Config.VolumesBlacklist
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}

	if vol.Config.Ignore {
		return true, "blacklisted", "volume config"
	}

	return false, "", ""
}
