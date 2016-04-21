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

func (*DefaultProvider) GetName() string {
	return "Default"
}

func (p *DefaultProvider) GetBackupDir() string {
	return ""
}

func (*DefaultProvider) PrepareBackup() error {
	log.Infof("Nothing to do to prepare backup for default provider")
	return nil
}
