package providers

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
func (p *PostgreSQLProvider) GetPrepareCommand(volDestination string) []string {
	return []string{
		"sh",
		"-c",
		"mkdir -p " + volDestination + "/backups && pg_dumpall --clean -Upostgres > " + volDestination + "/backups/all.sql",
	}
}

// GetBackupDir returns the backup directory used by the provider
func (p *PostgreSQLProvider) GetBackupDir() string {
	return "backups"
}

// SetVolumeBackupDir sets the backup dir for the volume
func (p *PostgreSQLProvider) SetVolumeBackupDir() {
	p.vol.BackupDir = p.GetBackupDir()
}
