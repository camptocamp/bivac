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
	ReadOnly   string
	HostBind   string
	Hostname   string
	Namespace  string
}
