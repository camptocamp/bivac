package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

// DefaultProvider implements a BaseProvider struct
// for simple filesystem backups
type DefaultProvider struct {
	*BaseProvider
}

// GetName returns the provider name
func (*DefaultProvider) GetName() string {
	return "Default"
}

// PrepareBackup sets up the data before backup
func (p *DefaultProvider) PrepareBackup() error {
	log.WithFields(log.Fields{
		"provider": p.GetName(),
	}).Debug("Provider does not implement a prepare method")
	return nil
}

// GetPrepareCommand returns the command to be executed before backup
func (p *DefaultProvider) GetPrepareCommand(mount *docker.Mount) []string {
	return nil
}
