package engines

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/lib"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
)

// Engine implements a backup engine interface
type Engine interface {
	Backup() ([]string, error)
	GetName() string
}

// GetEngine returns the engine for passed volume
func GetEngine(c *conplicity.Conplicity, v *volume.Volume) Engine {
	engine, _ := util.GetVolumeLabel(v.Volume, ".engine")
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
	case "rclone":
		return &RCloneEngine{
			Config: c.Config,
			Docker: c.Client,
			Volume: v,
		}
	}

	log.Fatalf("Unknown engine %s", engine)
	return nil
}
