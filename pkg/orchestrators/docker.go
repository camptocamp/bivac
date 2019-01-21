package orchestrators

import (
	"fmt"
	"io/ioutil"
	"sort"
	"unicode/utf8"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"
	"golang.org/x/net/context"

	"github.com/camptocamp/bivac/pkg/volume"
)

// DockerConfig stores Docker configuration
type DockerConfig struct {
	Endpoint string
}

// DockerOrchestrator implements a container orchestrator for Docker
type DockerOrchestrator struct {
	client *docker.Client
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
			Mountpoint: voll.Mountpoint,
			Name:       voll.Name,
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
func (o *DockerOrchestrator) DeployAgent(cmd []string, envs []string, volume *volume.Volume) (success bool, output string, err error) {
	return
	/*
		err = o.PullImage()
		if err != nil {
			err = fmt.Errorf("failed to pull image: %s", err)
			return
		}

		container, err := o.client.ContainerCreate(
			context.Background(),
			&container.Config{
				Cmd:          cmd,
				Env:          envs,
				Image:        "camptocamp/bivac:v2",
				OpenStdin:    true,
				StdinOnce:    true,
				AttachStdin:  true,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          false,
			},
			&container.HostConfig{
				Mounts: mounts,
			}, nil, "",
		)
		if err != nil {
			err = fmt.Errorf("failed to create container: %s", err)
			return
		}
		//defer o.RemoveContainer(container.ID)

		return
	*/
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

func (o *DockerOrchestrator) PullImage() (err error) {
	if _, _, err = o.client.ImageInspectWithRaw(context.Background(), "camptocamp/bivac:v2"); err != nil {
		resp, err := o.client.ImagePull(context.Background(), "camptocamp/bivac:v2", types.ImagePullOptions{})
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
	return
}
