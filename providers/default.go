package providers

import (
	log "github.com/Sirupsen/logrus"
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
	log.Infof("Provider %v does not implement a prepare method", p.GetName())
	return nil
}
