package providers

import (
	log "github.com/Sirupsen/logrus"
)

const labelPrefix string = "io.conplicity"

// DefaultProvider implements a BaseProvider struct
// for simple filesystem backups
type DefaultProvider struct {
	*BaseProvider
}

// GetName returns the provider name
func (*DefaultProvider) GetName() string {
	return "Default"
}

// GetBackupDir returns the backup directory
func (p *DefaultProvider) GetBackupDir() string {
	return ""
}

// PrepareBackup sets up the data before backup
func (*DefaultProvider) PrepareBackup() error {
	log.Infof("Nothing to do to prepare backup for default provider")
	return nil
}
