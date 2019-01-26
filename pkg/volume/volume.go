package volume

// Volume provides backup methods for a single volume
type Volume struct {
	ID         string
	Name       string
	BackupDir  string
	Mount      string
	Mountpoint string
	Driver     string
	Labels     map[string]string
	ReadOnly   bool
	HostBind   string
	Hostname   string
	Namespace  string

	LastBackupDate   string
	LastBackupStatus string
	Logs             map[string]string
}

// Filters hfcksdghfvd
type Filters struct {
	Whitelist []string
	Blacklist []string
}
