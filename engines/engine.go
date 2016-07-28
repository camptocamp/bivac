package engines

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/lib"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
)

// Engine implements a backup engine interface
type Engine interface {
	// RemoveOld cleans up old backup data
	RemoveOld(*volume.Volume) ([]string, error)

	// Cleanup removes old index data
	Cleanup(*volume.Volume) ([]string, error)

	// Verify checks that the backup is usable
	Verify(*volume.Volume) ([]string, error)

	// Status gets the latest backup date info
	Status(*volume.Volume) ([]string, error)

	Backup() ([]string, error)

	GetName() string
}

// GetEngine returns the engine for passed volume
func GetEngine(c *conplicity.Conplicity, vol *types.Volume) Engine {
	v := &volume.Volume{
		Config: c.Config,
		Volume: vol,
	}

	engine, _ := util.GetVolumeLabel(vol, ".engine")
	if engine == "" {
		engine = c.Config.Engine
	}

	switch engine {
	case "duplicity":
		return &DuplicityEngine{
			Config: c.Config,
			Docker: c.Client,
			Volume: v,
		}
	}

	log.Fatalf("Unknown engine %s", engine)
	return nil
}
