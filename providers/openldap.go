package providers

import "github.com/docker/docker/api/types"

// OpenLDAPProvider implements a BaseProvider struct
// for OpenLDAP backups
type OpenLDAPProvider struct {
	*BaseProvider
}

// GetName returns the provider name
func (p *OpenLDAPProvider) GetName() string {
	return "OpenLDAP"
}

// GetPrepareCommand returns the command to be executed before backup
func (p *OpenLDAPProvider) GetPrepareCommand(mount *types.MountPoint) []string {
	return []string{
		"sh",
		"-c",
		"mkdir -p " + mount.Destination + "/backups && slapcat > " + mount.Destination + "/backups/all.ldif",
	}
}

// GetBackupDir returns the backup directory used by the provider
func (p *OpenLDAPProvider) GetBackupDir() string {
	return "backups"
}
