package providers

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/volume"
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
func GetProvider(o orchestrators.Orchestrator, v *volume.Volume) Provider {
	log.WithFields(log.Fields{
		"volume": v.Name,
	}).Info("Detecting provider")
	p := &BaseProvider{
		orchestrator: o,
		vol:          v,
	}
	shell := `([[ -d ` + v.Mountpoint + `/mysql ]] && echo 'mysql') || ` +
		`([[ -f ` + v.Mountpoint + `/PG_VERSION ]] && echo 'postgresql') || ` +
		`([[ -f ` + v.Mountpoint + `/DB_CONFIG ]] && echo 'openldap'); ` +
		`return 0`

	cmd := []string{
		"sh",
		"-c",
		shell,
	}
	_, stdout, err := o.LaunchContainer("busybox", map[string]string{}, cmd, []*volume.Volume{v})
	if err != nil && err.Error() != "EOF" {
		log.Errorf("failed to run provider detection: %s", err)
	}

	switch strings.TrimSpace(stdout) {
	case "mysql":
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("mysql directory found, this should be MySQL datadir")
		return &MySQLProvider{
			BaseProvider: p,
		}
	case "postgresql":
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("PG_VERSION file found, this should be a PostgreSQL datadir")
		return &PostgreSQLProvider{
			BaseProvider: p,
		}
	case "openldap":
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("DB_CONFIG file found, this should be and OpenLDAP datadir")
		return &OpenLDAPProvider{
			BaseProvider: p,
		}
	default:
		return &DefaultProvider{
			BaseProvider: p,
		}
	}
}

// PrepareBackup sets up the data before backup
func PrepareBackup(p Provider) (err error) {
	p.SetVolumeBackupDir()

	o := p.GetOrchestrator()

	vol := p.GetVolume()

	containers, err := o.GetMountedVolumes()
	if err != nil {
		err = fmt.Errorf("failed to list containers: %v", err)
		return
	}

	for _, container := range containers {
		for volName, volDestination := range container.Volumes {
			if volName == vol.Name {
				log.WithFields(log.Fields{
					"volume":    volName,
					"container": container.ContainerID,
				}).Debug("Container found using volume")

				cmd := p.GetPrepareCommand(volDestination)
				if cmd != nil {
					err = o.ContainerExec(container, cmd)
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
