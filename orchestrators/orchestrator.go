package orchestrators

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
	GetHandler() *handler.Bivac
	GetVolumes() ([]*volume.Volume, error)
	LaunchContainer(image string, env map[string]string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error)
	GetMountedVolumes() ([]*volume.MountedVolumes, error)
	ContainerExec(mountedVolumes *volume.MountedVolumes, command []string) error
	ContainerPrepareBackup(mountedVolumes *volume.MountedVolumes, command []string) (backupVolume *volume.Volume, err error)
}

// GetOrchestrator returns the Orchestrator based on configuration or environment if not defined
func GetOrchestrator(c *handler.Bivac) (orch Orchestrator, err error) {
	if c.Config.Orchestrator != "" {
		log.Debugf("Choosing orchestrator based on configuration...")
		switch c.Config.Orchestrator {
		case "docker":
			orch = NewDockerOrchestrator(c)
		case "kubernetes":
			orch = NewKubernetesOrchestrator(c)
		case "cattle":
			orch = NewCattleOrchestrator(c)
		default:
			err = fmt.Errorf("'%s' is not a valid orchestrator", c.Config.Orchestrator)
			return
		}
	} else {
		log.Debugf("Detecting orchestrator based on environment...")
		if detectKubernetes() {
			orch = NewKubernetesOrchestrator(c)
		} else if detectCattle() {
			orch = NewCattleOrchestrator(c)
		} else if detectDocker(c) {
			orch = NewDockerOrchestrator(c)
		} else {
			err = fmt.Errorf("no orchestrator detected")
			return
		}
	}
	log.Debugf("Using orchestrator: %s", orch.GetName())
	return
}
