package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

const labelPrefix string = "io.conplicity"

type DefaultProvider struct {
	handler *handler.Conplicity
	vol     *docker.Volume
}

func (*DefaultProvider) GetName() string {
	return "Default"
}

func (p *DefaultProvider) GetHandler() *handler.Conplicity {
	return p.handler
}

func (*DefaultProvider) PrepareBackup() error {
	log.Infof("Nothing to do to prepare backup for default provider")
	return nil
}
