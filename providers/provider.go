package providers

import (
	"os"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	conplicity "github.com/camptocamp/conplicity/lib"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetPrepareCommand(*types.MountPoint) []string
	GetHandler() *conplicity.Conplicity
	GetVolume() *volume.Volume
	GetBackupDir() string
}

// BaseProvider is a struct implementing the Provider interface
type BaseProvider struct {
	handler   *conplicity.Conplicity
	vol       *volume.Volume
	backupDir string
}

// GetProvider detects which provider suits the passed volume and returns it
func GetProvider(c *conplicity.Conplicity, vol *volume.Volume) Provider {
	v := vol
	log.WithFields(log.Fields{
		"volume": v.Name,
	}).Info("Detecting provider")
	p := &BaseProvider{
		handler: c,
		vol:     v,
	}
	if f, err := os.Stat(v.Volume.Mountpoint + "/PG_VERSION"); err == nil && f.Mode().IsRegular() {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("PG_VERSION file found, this should be a PostgreSQL datadir")
		return &PostgreSQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Volume.Mountpoint + "/mysql"); err == nil && f.Mode().IsDir() {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("mysql directory found, this should be MySQL datadir")
		return &MySQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Volume.Mountpoint + "/DB_CONFIG"); err == nil && f.Mode().IsRegular() {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("DB_CONFIG file found, this should be and OpenLDAP datadir")
		return &OpenLDAPProvider{
			BaseProvider: p,
		}
	}

	return &DefaultProvider{
		BaseProvider: p,
	}
}

// PrepareBackup sets up the data before backup
func PrepareBackup(p Provider) (err error) {
	c := p.GetHandler()
	vol := p.GetVolume()
	containers, err := c.ContainerList(context.Background(), types.ContainerListOptions{})
	util.CheckErr(err, "Failed to list containers: %v", "fatal")

	// Work around https://github.com/docker/engine-api/issues/303
	client, err := docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	util.CheckErr(err, "Failed to create new Docker client: %v", "fatal")

	for _, container := range containers {
		container, err := client.ContainerInspect(context.Background(), container.ID)
		util.CheckErr(err, "Failed to inspect container "+container.ID+": %v", "fatal")
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.WithFields(log.Fields{
					"volume":    vol.Name,
					"container": container.ID,
				}).Debug("Container found using volume")

				cmd := p.GetPrepareCommand(&mount)
				if cmd != nil {
					exec, err := client.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
						Cmd: p.GetPrepareCommand(&mount),
					},
					)

					util.CheckErr(err, "Failed to create exec: %v", "fatal")

					err = client.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{})

					util.CheckErr(err, "Failed to start exec: %v", "fatal")
				} else {
					log.WithFields(log.Fields{
						"volume":    vol.Name,
						"container": container.ID,
					}).Info("No prepare command to execute in container")
				}
			}
		}
	}
	return
}

// GetHandler returns the handler associated with the provider
func (p *BaseProvider) GetHandler() *conplicity.Conplicity {
	return p.handler
}

// GetVolume returns the volume associated with the provider
func (p *BaseProvider) GetVolume() *volume.Volume {
	return p.vol
}

// GetBackupDir returns the backup directory used by the provider
func (p *BaseProvider) GetBackupDir() string {
	return p.backupDir
}
