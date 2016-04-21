package providers

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

type Provider interface {
	GetName() string
	PrepareBackup() error
	BackupVolume() error
}

func GetProvider(c *handler.Conplicity, v *docker.Volume) Provider {
	log.Infof("Detecting provider for volume %v", v.Name)
	if f, err := os.Stat(v.Mountpoint + "/PG_VERSION"); err == nil && f.Mode().IsRegular() {
		return &PostgreSQLProvider{
			handler: c,
			vol:     v,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/mysql"); err == nil && f.Mode().IsDir() {
		return &MySQLProvider{
			handler: c,
			vol:     v,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/DB_CONFIG"); err == nil && f.Mode().IsRegular() {
		return &OpenLDAPProvider{
			handler: c,
			vol:     v,
		}
	} else {
		return &DefaultProvider{
			handler: c,
			vol:     v,
		}
	}
}
