package orchestrators

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetHandler() *handler.Bivac
	GetVolumes() ([]*volume.Volume, error)
	LaunchContainer(image string, env map[string]string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error)
	GetMountedVolumes() ([]*volume.MountedVolumes, error)
	ContainerExec(mountedVolumes *volume.MountedVolumes, command []string) error
}

// GetOrchestrator returns the Orchestrator as specified in configuration
func GetOrchestrator(c *handler.Bivac) Orchestrator {
	orch := c.Config.Orchestrator
	log.Debugf("Using orchestrator: %s", orch)

	switch orch {
	case "docker":
		return NewDockerOrchestrator(c)
	case "kubernetes":
		return NewKubernetesOrchestrator(c)
	case "cattle":
		return NewCattleOrchestrator(c)
	}

	log.Fatalf("Unknown orchestrator %s", orch)
	return nil
}
