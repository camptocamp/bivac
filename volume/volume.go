package volume

import (
	"github.com/camptocamp/conplicity/config"
	"github.com/docker/engine-api/types"
)

// Handler is an interface providing a mean of launching Duplicity
type Handler interface {
	LaunchDuplicity([]string, []string) (int, string, error)
}

// Volume provides backup methods for a single Docker volume
type Volume struct {
	Name            string
	Volume          *types.Volume
	Target          string
	BackupDir       string
	Mount           string
	FullIfOlderThan string
	RemoveOlderThan string
	Config          *config.Config
}
