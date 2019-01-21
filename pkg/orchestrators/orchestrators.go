package orchestrators

import (
	"github.com/camptocamp/bivac/pkg/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
	GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error)
	DeployAgent(cmd []string, envs []string, volume *volume.Volume) (success bool, output string, err error)
}
