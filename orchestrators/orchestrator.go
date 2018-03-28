package orchestrators

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetHandler() *handler.Conplicity
	GetVolumes() ([]*volume.Volume, error)
	LaunchContainer(image string, env map[string]string, cmd []string, v *volume.Volume) (state int, stdout string, err error)
	GetMountedVolumes() ([]*volume.MountedVolumes, error)
	ContainerExec(containerID string, command []string) error
}

// GetOrchestrator returns the Orchestrator as specified in configuration
func GetOrchestrator(c *handler.Conplicity) Orchestrator {
	orch := c.Config.Orchestrator
	log.Debugf("orchestrator=%s", orch)

	switch orch {
	case "docker":
		return NewDockerOrchestrator(c)
	}

	log.Fatalf("Unknown orchestrator %s", orch)
	return nil
}
