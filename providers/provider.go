package providers

import (
	"os"
	"strings"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	conplicity "github.com/camptocamp/conplicity/lib"
	"github.com/docker/engine-api/types"
)

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetPrepareCommand(*types.MountPoint) []string
	GetHandler() *conplicity.Conplicity
	GetVolume() *types.Volume
	GetBackupDir() string
	BackupVolume(*types.Volume) error
}

// BaseProvider is a struct implementing the Provider interface
type BaseProvider struct {
	handler   *conplicity.Conplicity
	vol       *types.Volume
	backupDir string
}

// GetProvider detects which provider suits the passed volume and returns it
func GetProvider(c *conplicity.Conplicity, v *types.Volume) Provider {
	log.WithFields(log.Fields{
		"volume": v.Name,
	}).Info("Detecting provider")
	p := &BaseProvider{
		handler: c,
		vol:     v,
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
	c := p.GetHandler()
	vol := p.GetVolume()
	containers, err := c.ContainerList(context.Background(), types.ContainerListOptions{})
	conplicity.CheckErr(err, "Failed to list containers: %v", "fatal")
	for _, container := range containers {
		container, err := c.ContainerInspect(context.Background(), container.ID)
		conplicity.CheckErr(err, "Failed to inspect container "+container.ID+": %v", "fatal")
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.WithFields(log.Fields{
					"volume":    vol.Name,
					"container": container.ID,
				}).Debug("Container found using volume")

				cmd := p.GetPrepareCommand(&mount)
				if cmd != nil {
					exec, err := c.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
						Cmd: p.GetPrepareCommand(&mount),
					},
					)

					conplicity.CheckErr(err, "Failed to create exec: %v", "fatal")

					err = c.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{})

					conplicity.CheckErr(err, "Failed to start exec: %v", "fatal")
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

// BackupVolume performs the backup of the passed volume
func (p *BaseProvider) BackupVolume(vol *types.Volume) (err error) {
	log.WithFields(log.Fields{
		"volume":     vol.Name,
		"driver":     vol.Driver,
		"mountpoint": vol.Mountpoint,
	}).Info("Creating duplicity container")

	c := p.GetHandler()

	fullIfOlderThan, _ := conplicity.GetVolumeLabel(vol, ".full_if_older_than")
	if fullIfOlderThan == "" {
		fullIfOlderThan = c.Config.Duplicity.FullIfOlderThan
	}

	removeOlderThan, _ := conplicity.GetVolumeLabel(vol, ".remove_older_than")
	if removeOlderThan == "" {
		removeOlderThan = c.Config.Duplicity.RemoveOlderThan
	}

	pathSeparator := "/"
	if strings.HasPrefix(c.Config.Duplicity.TargetURL, "swift://") {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	backupDir := p.GetBackupDir()
	fullTarget := c.Config.Duplicity.TargetURL + pathSeparator + c.Hostname + pathSeparator + vol.Name
	fullBackupDir := vol.Mountpoint + "/" + backupDir
	roMount := vol.Name + ":" + vol.Mountpoint + ":ro"

	volume := &conplicity.Volume{
		Name:            vol.Name,
		Target:          fullTarget,
		BackupDir:       fullBackupDir,
		Mount:           roMount,
		FullIfOlderThan: fullIfOlderThan,
		RemoveOlderThan: removeOlderThan,
		Client:          c,
	}

	var newMetrics []string

	newMetrics, err = volume.Backup()
	conplicity.CheckErr(err, "Failed to backup volume "+vol.Name+" : %v", "fatal")
	c.Metrics = append(c.Metrics, newMetrics...)

	_, err = volume.RemoveOld()
	conplicity.CheckErr(err, "Failed to remove old backups for volume "+vol.Name+" : %v", "fatal")

	_, err = volume.Cleanup()
	conplicity.CheckErr(err, "Failed to cleanup extraneous duplicity files for volume "+vol.Name+" : %v", "fatal")

	noVerifyLbl, _ := conplicity.GetVolumeLabel(vol, ".no_verify")
	noVerify := c.Config.NoVerify || (noVerifyLbl == "true")
	if noVerify {
		log.WithFields(log.Fields{
			"volume": vol.Name,
		}).Info("Skipping verification")
	} else {
		newMetrics, err = volume.Verify()
		conplicity.CheckErr(err, "Failed to verify backup for volume "+vol.Name+" : %v", "fatal")
		c.Metrics = append(c.Metrics, newMetrics...)
	}

	newMetrics, err = volume.Status()
	conplicity.CheckErr(err, "Failed to retrieve last backup info for volume "+vol.Name+" : %v", "fatal")
	c.Metrics = append(c.Metrics, newMetrics...)

	return
}

// GetHandler returns the handler associated with the provider
func (p *BaseProvider) GetHandler() *conplicity.Conplicity {
	return p.handler
}

// GetVolume returns the volume associated with the provider
func (p *BaseProvider) GetVolume() *types.Volume {
	return p.vol
}

// GetBackupDir returns the backup directory used by the provider
func (p *BaseProvider) GetBackupDir() string {
	return p.backupDir
}
