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

func NewVolume(v *types.Volume) *Volume {
	return &Volume{
		Volume: v,
	}
}
