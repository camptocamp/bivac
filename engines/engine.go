package engines

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
)

// Engine implements a backup engine interface
type Engine interface {
	Backup() error
	GetName() string
}

// GetEngine returns the engine for passed volume
func GetEngine(c *handler.Conplicity, v *volume.Volume) Engine {
	engine, _ := util.GetVolumeLabel(v.Volume, ".engine")
	if engine == "" {
		engine = c.Config.Engine
	}

	switch engine {
	case "duplicity":
		return &DuplicityEngine{
			Handler: c,
			Volume:  v,
		}
	case "rclone":
		return &RCloneEngine{
			Handler: c,
			Volume:  v,
		}
	}

	log.Fatalf("Unknown engine %s", engine)
	return nil
}
