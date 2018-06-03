package engines

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/volume"
)

// Engine implements a backup engine interface
type Engine interface {
	Backup() error
	GetName() string
	StdinSupport() bool
}

// GetEngine returns the engine for passed volume
func GetEngine(o orchestrators.Orchestrator, v *volume.Volume) Engine {
	engine := v.Config.Engine
	log.Debugf("engine=%s", engine)

	switch engine {
	case "duplicity":
		return &DuplicityEngine{
			Orchestrator: o,
			Volume:       v,
		}
	case "rclone":
		return &RCloneEngine{
			Orchestrator: o,
			Volume:       v,
		}
	case "restic":
		return &ResticEngine{
			Orchestrator: o,
			Volume:       v,
		}
	}

	log.Fatalf("Unknown engine %s", engine)
	return nil
}
