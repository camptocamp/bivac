package orchestrators

import (
	"fmt"

	"github.com/docker/docker/api/types"
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
func (o *DockerOrchestrator) GetVolumes(whitelist, blacklist []string) (volumes []*volume.Volume, err error) {
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
			Mountpoint: voll.Mountpoint,
			Name:       voll.Name,
			Labels:     voll.Labels,
		}

		if b, r, s := o.blacklistedVolume(v); b {
			continue
		}
		volumes = append(volumes, v)
	}
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

func (o *DockerOrchestrator) blacklistedVolume(vol *volume.Volume, whitelist []string) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	// Use whitelist if defined
	if l := whitelist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return false, "", ""
		}
		return true, "blacklisted", "whitelist config"
	}

	i := sort.SearchStrings(blacklist, vol.Name)
	if i < len(blacklist) && blacklist[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}
	return false, "", ""
}
