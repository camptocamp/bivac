package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

type Provider interface {
  PrepareBackup() error
  BackupVolume() error
}

func GetProvider(c *handler.Conplicity, v *docker.Volume) Provider {
  log.Infof("Detecting provider for volume %v", v.Name)
  return &DefaultProvider{
    handler: c,
    vol: v,
  }
}
