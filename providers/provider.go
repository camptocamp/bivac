package providers

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/orchestrators"
	"github.com/camptocamp/conplicity/volume"
)

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetPrepareCommand(string) []string
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

	containers, err := o.GetMountedVolumes()
	if err != nil {
		err = fmt.Errorf("failed to list containers: %v", err)
		return
	}

	for _, container := range containers {
		for volName, volDestination := range container.Volumes {
			log.WithFields(log.Fields{
				"volume":    volName,
				"container": container.ContainerID,
			}).Debug("Container found using volume")

			cmd := p.GetPrepareCommand(volDestination)
			if cmd != nil {
				err = o.ContainerExec(container.ContainerID, cmd)
				if err != nil {
					return fmt.Errorf("failed to execute command in container: %v", err)
				}
			} else {
				log.WithFields(log.Fields{
					"volume":    volName,
					"container": container.ContainerID,
				}).Info("No prepare command to execute in container")
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
