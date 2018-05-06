package providers

// MySQLProvider implements a BaseProvider struct
// for MySQL backups
type MySQLProvider struct {
	*BaseProvider
}

// GetName returns the provider name
func (*MySQLProvider) GetName() string {
	return "MySQL"
}

// GetPrepareCommand returns the command to be executed before backup
func (p *MySQLProvider) GetPrepareCommand(volDestination string) []string {
	return []string{
		"sh",
		"-c",
		"mkdir -p " + volDestination + "/backups && mysqldump --all-databases --extended-insert --password=$MYSQL_ROOT_PASSWORD",
	}
}

// GetBackupDir returns the backup directory used by the provider
func (p *MySQLProvider) GetBackupDir() string {
	return "backups"
}

// SetVolumeBackupDir sets the backup dir for the volume
func (p *MySQLProvider) SetVolumeBackupDir() {
	p.vol.BackupDir = p.GetBackupDir()
}
