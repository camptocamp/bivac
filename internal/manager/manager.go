package manager

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/internal/server"
	"github.com/camptocamp/bivac/pkg/orchestrators"
)

type Orchestrators struct {
	Docker orchestrators.DockerConfig
}

// Start starts a Bivac manager which handle backups management
func Start(o orchestrators.Orchestrator, s server.Server) (err error) {
	err = server.Start(o, &s)
	if err != nil {
		log.Errorf("failed to start server: %s", err)
	}
	return
}

func GetOrchestrator(name string, orchs Orchestrators) (o orchestrators.Orchestrator, err error) {
	if name != "" {
		log.Debugf("Choosing orchestrator based on configuration...")
		switch name {
		case "docker":
			o, err = orchestrators.NewDockerOrchestrator(&orchs.Docker)
		default:
			err = fmt.Errorf("'%s' is not a valid orchestrator")
			return
		}
	} else {
		log.Debugf("Trying to detect orchestrator based on environment...")
		if orchestrators.DetectDocker(&orchs.Docker) {
			o, err = orchestrators.NewDockerOrchestrator(&orchs.Docker)
		} else {
			err = fmt.Errorf("no orchestrator detected")
			return
		}
	}
	if err != nil {
		log.Infof("Using orchestrator: %s", o.GetName())
	}
	return
}
