package orchestrators

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"unicode/utf8"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	docker "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"

	"github.com/camptocamp/bivac/pkg/volume"
)

// DockerConfig stores Docker configuration
type DockerConfig struct {
	Endpoint string
}

// DockerOrchestrator implements a container orchestrator for Docker
type DockerOrchestrator struct {
	client docker.CommonAPIClient
}

// NewDockerOrchestrator creates a Docker client
func NewDockerOrchestrator(config *DockerConfig) (o *DockerOrchestrator, err error) {
	o = &DockerOrchestrator{}
	o.client, err = docker.NewClient(config.Endpoint, "", nil, nil)
	if err != nil {
		err = fmt.Errorf("failed to create a Docker client: %s", err)
	}
	return
}

// GetName returns the orchestrator's name
func (o *DockerOrchestrator) GetName() string {
	return "docker"
}

// GetPath returns the backup path
func (*DockerOrchestrator) GetPath(v *volume.Volume) string {
	return v.Hostname
}

// GetVolumes returns the Docker volumes, inspected and filtered
func (o *DockerOrchestrator) GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error) {

	info, err := o.client.Info(context.Background())
	if err != nil {
		err = fmt.Errorf("failed to retrieve Docker engine info: %s", err)
		return
	}

	vols, err := o.client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		err = fmt.Errorf("failed to list Docker volumes: %v", err)
		return
	}

	var voll types.Volume
	for _, vol := range vols.Volumes {
		voll, err = o.client.VolumeInspect(context.Background(), vol.Name)
		if err != nil {
			err = fmt.Errorf("failed to inspect volume `%s': %v", vol.Name, err)
			return
		}

		v := &volume.Volume{
			ID:         voll.Name,
			Name:       voll.Name,
			Mountpoint: voll.Mountpoint,
			HostBind:   info.Name,
			Hostname:   info.Name,
			Labels:     voll.Labels,
		}

		if b, _, _ := o.blacklistedVolume(v, volumeFilters); b {
			continue
		}
		volumes = append(volumes, v)
	}
	return
}

// DeployAgent creates a `bivac agent` container
func (o *DockerOrchestrator) DeployAgent(image string, cmd []string, envs []string, v *volume.Volume) (success bool, output string, err error) {
	success = false
	err = o.PullImage(image)
	if err != nil {
		err = fmt.Errorf("failed to pull image: %s", err)
		return
	}

	mounts := []mount.Mount{
		mount.Mount{
			Type:     "volume",
			Target:   v.Mountpoint,
			Source:   v.Name,
			ReadOnly: v.ReadOnly,
		},
	}

	container, err := o.client.ContainerCreate(
		context.Background(),
		&containertypes.Config{
			Cmd:          cmd,
			Env:          envs,
			Image:        image,
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
		},
		&containertypes.HostConfig{
			Mounts: mounts,
		}, nil, "")

	if err != nil {
		err = fmt.Errorf("failed to create container: %s", err)
		return
	}
	defer o.RemoveContainer(container.ID)

	err = o.client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		err = fmt.Errorf("failed to start container: %s", err)
		return
	}

	var exited bool

	for !exited {
		var cont types.ContainerJSON
		cont, err = o.client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %s", err)
			return
		}

		if cont.State.Status == "exited" {
			exited = true
		}
	}

	body, err := o.client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Follow:     true,
	})
	if err != nil {
		err = fmt.Errorf("failed to retrieve logs: %s", err)
		return
	}

	defer body.Close()
	stdout := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(stdout, ioutil.Discard, body)
	if err != nil {
		err = fmt.Errorf("failed to read logs from response: %s", err)
		return
	}
	output = stdout.String()
	success = true
	return
}

// RemoveContainer
func (o *DockerOrchestrator) RemoveContainer(containerID string) (err error) {
	err = o.client.ContainerRemove(context.Background(), containerID, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	return
}

// DetectDocker tries to detect a Docker orchestrator by connecting to the endpoint
func DetectDocker(config *DockerConfig) bool {
	client, err := docker.NewClient(config.Endpoint, "", nil, nil)
	if err != nil {
		return false
	}
	_, err = client.Ping(context.Background())
	if err != nil {
		return false
	}
	return true
}

func (o *DockerOrchestrator) blacklistedVolume(vol *volume.Volume, volumeFilters volume.Filters) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	// Use whitelist if defined
	if l := volumeFilters.Whitelist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return false, "", ""
		}
		return true, "blacklisted", "whitelist config"
	}

	i := sort.SearchStrings(volumeFilters.Blacklist, vol.Name)
	if i < len(volumeFilters.Blacklist) && volumeFilters.Blacklist[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}
	return false, "", ""
}

func (o *DockerOrchestrator) PullImage(image string) (err error) {
	if _, _, err = o.client.ImageInspectWithRaw(context.Background(), image); err != nil {
		resp, err := o.client.ImagePull(context.Background(), image, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer resp.Close()

		_, err = ioutil.ReadAll(resp)
		if err != nil {
			err = fmt.Errorf("failed to read ImagePull response: %s", err)
			return err
		}
	}
	return nil
}

// GetContainersMountingVolume returns mounted volumes
func (o *DockerOrchestrator) GetContainersMountingVolume(v *volume.Volume) (mountedVolumes []*volume.MountedVolume, err error) {
	c, err := o.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = fmt.Errorf("failed to list containers: %s", err)
		return
	}

	for _, container := range c {
		for _, mount := range container.Mounts {
			if mount.Name == v.Name && mount.Type == "volume" {
				mv := &volume.MountedVolume{
					ContainerID: container.ID,
					Volume:      v,
					Path:        mount.Destination,
				}
				mountedVolumes = append(mountedVolumes, mv)
			}
		}
	}
	return
}

// ContainerExec executes a command in a container
func (o *DockerOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolume, command []string) (stdout string, err error) {
	exec, err := o.client.ContainerExecCreate(context.Background(), mountedVolumes.ContainerID, types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	})
	if err != nil {
		err = fmt.Errorf("failed to create exec: %s", err)
		return
	}

	conn, err := o.client.ContainerExecAttach(context.Background(), exec.ID, types.ExecStartCheck{})
	if err != nil {
		err = fmt.Errorf("failed to attach: %s", err)
		return
	}
	defer conn.Close()

	err = o.client.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{})
	if err != nil {
		err = fmt.Errorf("failed to start exec: %s", err)
		return
	}

	stdoutput := new(bytes.Buffer)
	stdcopy.StdCopy(stdoutput, ioutil.Discard, conn.Reader)

	stdout = stdoutput.String()
	return
}
