package conplicity

import "github.com/docker/engine-api/types"

// PostgreSQLProvider implements a BaseProvider struct
// for PostgreSQL backups
type PostgreSQLProvider struct {
	*BaseProvider
}

// GetName returns the provider name
func (p *PostgreSQLProvider) GetName() string {
	return "PostgreSQL"
}

// GetPrepareCommand returns the command to be executed before backup
func (p *PostgreSQLProvider) GetPrepareCommand(mount *types.MountPoint) []string {
	return []string{
		"sh",
		"-c",
		"mkdir -p " + mount.Destination + "/backups && pg_dumpall -Upostgres > " + mount.Destination + "/backups/all.sql",
	}
}

// GetBackupDir returns the backup directory used by the provider
func (p *PostgreSQLProvider) GetBackupDir() string {
	return "backups"
}
