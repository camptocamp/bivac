package volume

import "github.com/docker/engine-api/types"

// Volume provides backup methods for a single Docker volume
type Volume struct {
	*types.Volume
	Target          string
	BackupDir       string
	Mount           string
	FullIfOlderThan string
	RemoveOlderThan string
}

// NewVolume returns a new Volume for a given types.Volume struct
func NewVolume(v *types.Volume) *Volume {
	return &Volume{
		Volume: v,
	}
}
