package orchestrators

import (
	"github.com/camptocamp/bivac/pkg/volume"
)

// Orchestrator implements a container Orchestrator interface
type Orchestrator interface {
	GetName() string
	GetPath(v *volume.Volume) string
	GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error)
	DeployAgent(image string, cmd []string, envs []string, volume *volume.Volume) (success bool, output string, err error)
	GetContainersMountingVolume(v *volume.Volume) (mountedVolumes []*volume.MountedVolume, err error)
	ContainerExec(mountedVolumes *volume.MountedVolume, command []string) (stdout string, err error)
	IsNodeAvailable(hostID string) (ok bool, err error)
}
