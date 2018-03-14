package providers

import (
	"fmt"
	"os"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/orchestrators"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
)

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetPrepareCommand(*types.MountPoint) []string
	GetOrchestrator() orchestrators.Orchestrator
	GetVolume() *volume.Volume
	GetBackupDir() string
	SetVolumeBackupDir()
}

// BaseProvider is a struct implementing the Provider interface
type BaseProvider struct {
	orchestrator orchestrators.Orchestrator
	vol          *volume.Volume
	backupDir    string
}

// GetProvider detects which provider suits the passed volume and returns it
func GetProvider(o orchestrators.Orchestrator, vol *volume.Volume) Provider {
	v := vol
	log.WithFields(log.Fields{
		"volume": v.Name,
	}).Info("Detecting provider")
	p := &BaseProvider{
		orchestrator: o,
		vol:          v,
	}
	if f, err := os.Stat(v.Mountpoint + "/PG_VERSION"); err == nil && f.Mode().IsRegular() {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("PG_VERSION file found, this should be a PostgreSQL datadir")
		return &PostgreSQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/mysql"); err == nil && f.Mode().IsDir() {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("mysql directory found, this should be MySQL datadir")
		return &MySQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/DB_CONFIG"); err == nil && f.Mode().IsRegular() {
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
	p.SetVolumeBackupDir()

	o := p.GetOrchestrator()
	c := o.GetHandler()
	vol := p.GetVolume()
	containers, err := c.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list containers: %v", err)
	}

	// Work around https://github.com/docker/engine-api/issues/303
	client, err := docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create new Docker client: %v", err)
	}

	for _, container := range containers {
		container, err := client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			return fmt.Errorf("failed to inspect container %v: %v", container.ID, err)
		}
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
					if err != nil {
						return fmt.Errorf("failed to create exec: %v", err)
					}

					err = client.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{})
					if err != nil {
						return fmt.Errorf("failed to start exec: %v", err)
					}

					inspect, err := client.ContainerExecInspect(context.Background(), exec.ID)
					if err != nil {
						return fmt.Errorf("failed to check prepare command exit code: %v", err)
					}
					if c := inspect.ExitCode; c != 0 {
						return fmt.Errorf("prepare command exited with code %v", c)
					}
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

// GetOrchestrator returns the orchestrator associated with the provider
func (p *BaseProvider) GetOrchestrator() orchestrators.Orchestrator {
	return p.orchestrator
}

// GetVolume returns the volume associated with the provider
func (p *BaseProvider) GetVolume() *volume.Volume {
	return p.vol
}

// GetBackupDir returns the backup directory used by the provider
func (p *BaseProvider) GetBackupDir() string {
	return p.backupDir
}

// SetVolumeBackupDir sets the backup dir for the volume
func (p *BaseProvider) SetVolumeBackupDir() {
	p.vol.BackupDir = p.GetBackupDir()
}
