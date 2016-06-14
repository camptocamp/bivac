package providers

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/fsouza/go-dockerclient"
)

const labelPrefix string = "io.conplicity"

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetPrepareCommand(*docker.Mount) []string
	GetHandler() *handler.Conplicity
	GetVolume() *docker.Volume
	GetBackupDir() string
}

// BaseProvider is a struct implementing the Provider interface
type BaseProvider struct {
	handler   *handler.Conplicity
	vol       *docker.Volume
	backupDir string
}

// GetProvider detects which provider suits the passed volume and returns it
func GetProvider(c *handler.Conplicity, v *docker.Volume) Provider {
	log.Infof("Detecting provider for volume %v", v.Name)
	p := &BaseProvider{
		handler: c,
		vol:     v,
	}
	if f, err := os.Stat(v.Mountpoint + "/PG_VERSION"); err == nil && f.Mode().IsRegular() {
		log.Infof("PG_VERSION file found, this should be a PostgreSQL datadir")
		return &PostgreSQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/mysql"); err == nil && f.Mode().IsDir() {
		log.Infof("mysql directory found, this should be MySQL datadir")
		return &MySQLProvider{
			BaseProvider: p,
		}
	} else if f, err := os.Stat(v.Mountpoint + "/DB_CONFIG"); err == nil && f.Mode().IsRegular() {
		log.Infof("DB_CONFIG file found, this should be and OpenLDAP datadir")
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
	containers, err := c.ListContainers(docker.ListContainersOptions{})
	util.CheckErr(err, "Failed to list containers: %v", -1)
	for _, container := range containers {
		container, err := c.InspectContainer(container.ID)
		util.CheckErr(err, "Failed to inspect container "+container.ID+": %v", -1)
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.Infof("Volume %v is used by container %v", vol.Name, container.ID)

				cmd := p.GetPrepareCommand(&mount)
				if cmd != nil {

					exec, err := c.CreateExec(
						docker.CreateExecOptions{
							Container: container.ID,
							Cmd:       p.GetPrepareCommand(&mount),
						},
					)

					util.CheckErr(err, "Failed to create exec", 1)

					err = c.StartExec(
						exec.ID,
						docker.StartExecOptions{},
					)

					util.CheckErr(err, "Failed to create exec", 1)
				} else {
					log.Infof("Not executing command for volume %v in container %v", vol.Name, container.ID)
				}
			}
		}
	}
	return
}

// BackupVolume performs the backup of the passed volume
func BackupVolume(p Provider, vol *docker.Volume) (metrics []string, err error) {
	log.Infof("ID: " + vol.Name)
	log.Infof("Driver: " + vol.Driver)
	log.Infof("Mountpoint: " + vol.Mountpoint)

	log.Infof("Creating duplicity container...")

	c := p.GetHandler()

	fullIfOlderThan := getVolumeLabel(vol, ".full_if_older_than")
	if fullIfOlderThan == "" {
		fullIfOlderThan = c.FullIfOlderThan
	}

	removeOlderThan := getVolumeLabel(vol, ".remove_older_than")
	if removeOlderThan == "" {
		removeOlderThan = c.RemoveOlderThan
	}

	pathSeparator := "/"
	if strings.HasPrefix(c.DuplicityTargetURL, "swift://") {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	backupDir := p.GetBackupDir()
	fullTarget := c.DuplicityTargetURL + pathSeparator + c.Hostname + pathSeparator + vol.Name
	fullBackupDir := vol.Mountpoint + "/" + backupDir
	roMount := vol.Name + ":" + vol.Mountpoint + ":ro"

	volume := &volume.Volume{
		Name:            vol.Name,
		Target:          fullTarget,
		BackupDir:       fullBackupDir,
		Mount:           roMount,
		FullIfOlderThan: fullIfOlderThan,
		RemoveOlderThan: removeOlderThan,
		Client:          c,
	}

	var newMetrics []string

	_, err = volume.Backup()
	util.CheckErr(err, "Failed to backup volume "+vol.Name+" : %v", -1)

	_, err = volume.RemoveOld()
	util.CheckErr(err, "Failed to remove old backups for volume "+vol.Name+" : %v", -1)

	_, err = volume.Cleanup()
	util.CheckErr(err, "Failed to cleanup extraneous duplicity files for volume "+vol.Name+" : %v", -1)

	newMetrics, err = volume.Verify()
	util.CheckErr(err, "Failed to verify backup for volume "+vol.Name+" : %v", -1)
	metrics = append(metrics, newMetrics...)

	newMetrics, err = volume.Status()
	util.CheckErr(err, "Failed to retrieve last backup info for volume "+vol.Name+" : %v", -1)
	metrics = append(metrics, newMetrics...)

	return
}

func getVolumeLabel(vol *docker.Volume, key string) (value string) {
	value = vol.Labels[labelPrefix+key]
	return
}

// GetHandler returns the handler associated with the provider
func (p *BaseProvider) GetHandler() *handler.Conplicity {
	return p.handler
}

// GetVolume returns the volume associated with the provider
func (p *BaseProvider) GetVolume() *docker.Volume {
	return p.vol
}

// GetBackupDir returns the backup directory used by the provider
func (p *BaseProvider) GetBackupDir() string {
	return p.backupDir
}
