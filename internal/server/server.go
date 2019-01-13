package server

import (
	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/pkg/orchestrators"
)

// Server kzejgd
func Start(o orchestrators.Orchestrator) (err error) {
	log.Infof(o.GetName())
	return
}
