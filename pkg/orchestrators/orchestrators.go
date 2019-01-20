package orchestrators

import (
	"github.com/camptocamp/bivac/pkg/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
	GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error)
}
