package orchestrators

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
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

// GetVolumes returns the Docker volumes, inspected and filtered
func (o *DockerOrchestrator) GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error) {
	vols, err := o.client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		err = fmt.Errorf("failed to list Docker volumes: %v", err)
		return
	}

	for _, vol := range vols.Volumes {
		var voll types.Volume
		voll, err = o.client.VolumeInspect(context.Background(), vol.Name)
		if err != nil {
			err = fmt.Errorf("failed to inspect volume `%s': %v", vol.Name, err)
			return
		}

		v := &volume.Volume{
			ID:         voll.Name,
			Name:       voll.Name,
			Mountpoint: voll.Mountpoint,
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

	managerHostname, err := os.Hostname()
	if err != nil {
		err = fmt.Errorf("failed to get hostname: %s", err)
		return
	}
	envs = append(envs, fmt.Sprintf("RESTIC_HOSTNAME=%s", managerHostname))
	envs = append(envs, fmt.Sprintf("RESTIC_REPOSITORY=%s/%s/%s", os.Getenv("BIVAC_TARGET_URL"), managerHostname, v.Name))
	envs = append(envs, fmt.Sprintf("RESTIC_BACKUP_PATH=%s", v.Mountpoint))

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
