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
	Backup() ([]string, error)
	GetName() string
}

// GetEngine returns the engine for passed volume
func GetEngine(c *conplicity.Conplicity, vol *types.Volume) Engine {
	v := &volume.Volume{
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
