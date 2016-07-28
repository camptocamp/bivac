package engines

import (
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/volume"
	docker "github.com/docker/engine-api/client"
)

// RCloneEngine implements a backup engine with RClone
type RCloneEngine struct {
	Config *config.Config
	Docker *docker.Client
	Volume *volume.Volume
}

// GetName returns the engine name
func (*RCloneEngine) GetName() string {
	return "RClone"
}

// Backup performs the backup of the passed volume
func (r *RCloneEngine) Backup() (metrics []string, err error) {
	return
}
