package orchestrators

import (
	"github.com/camptocamp/bivac/pkg/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
	GetVolumes() ([]*volume.Volume, error)
}
