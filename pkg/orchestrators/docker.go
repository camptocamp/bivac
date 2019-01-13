package orchestrators

import (
	"fmt"

	docker "github.com/docker/docker/client"
	"golang.org/x/net/context"
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

func detectDocker(config *DockerConfig) bool {
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
