package providers

import "github.com/docker/engine-api/types"

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
func (p *MySQLProvider) GetPrepareCommand(mount *types.MountPoint) []string {
	return []string{
		"sh",
		"-c",
		"mkdir -p " + mount.Destination + "/backups && mysqldump --all-databases --extended-insert --password=$MYSQL_ROOT_PASSWORD > " + mount.Destination + "/backups/all.sql",
	}
}

// GetBackupDir returns the backup directory used by the provider
func (p *MySQLProvider) GetBackupDir() string {
	return "backups"
}
