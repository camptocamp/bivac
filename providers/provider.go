package providers

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/util"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

const labelPrefix string = "io.conplicity"

// A Provider is an interface for providers
type Provider interface {
	GetName() string
	GetHandler() *handler.Conplicity
	GetBackupDir() string
	PrepareBackup() error
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

// BackupVolume performs the backup of the passed volume
func BackupVolume(p Provider, vol *docker.Volume) (err error) {
	log.Infof("ID: " + vol.Name)
	log.Infof("Driver: " + vol.Driver)
	log.Infof("Mountpoint: " + vol.Mountpoint)

	log.Infof("Creating duplicity container...")

	c := p.GetHandler()

	fullIfOlderThan := getVolumeLabel(vol, ".full_if_older_than")
	if fullIfOlderThan == "" {
		fullIfOlderThan = c.FullIfOlderThan
	}

	pathSeparator := "/"
	if strings.HasPrefix(c.DuplicityTargetURL, "swift://") {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	backupDir := p.GetBackupDir()

	container, err := c.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Cmd: []string{
					"--full-if-older-than", fullIfOlderThan,
					"--s3-use-new-style",
					"--no-encryption",
					"--allow-source-mismatch",
					"--name", vol.Name,
					vol.Mountpoint + "/" + backupDir,
					c.DuplicityTargetURL + pathSeparator + c.Hostname + pathSeparator + vol.Name,
				},
				Env: []string{
					"AWS_ACCESS_KEY_ID=" + c.AWSAccessKeyID,
					"AWS_SECRET_ACCESS_KEY=" + c.AWSSecretAccessKey,
					"SWIFT_USERNAME=" + c.SwiftUsername,
					"SWIFT_PASSWORD=" + c.SwiftPassword,
					"SWIFT_AUTHURL=" + c.SwiftAuthURL,
					"SWIFT_TENANTNAME=" + c.SwiftTenantName,
					"SWIFT_REGIONNAME=" + c.SwiftRegionName,
					"SWIFT_AUTHVERSION=2",
				},
				Image:        c.Image,
				OpenStdin:    true,
				StdinOnce:    true,
				AttachStdin:  true,
				AttachStdout: true,
				AttachStderr: true,
				Tty:          true,
			},
		},
	)

	util.CheckErr(err, "Failed to create container for volume "+vol.Name+": %v", 1)

	defer func() {
		log.Infof("Removing container %v...", container.ID)
		c.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}()

	binds := []string{
		vol.Name + ":" + vol.Mountpoint + ":ro",
		"duplicity_cache:/root/.cache/duplicity",
	}

	err = dockerpty.Start(c.Client, container, &docker.HostConfig{
		Binds: binds,
	})
	util.CheckErr(err, "Failed to start container for volume "+vol.Name+": %v", -1)
	return
	return nil
}

func getVolumeLabel(vol *docker.Volume, key string) (value string) {
	value = vol.Labels[labelPrefix+key]
	return
}

// GetHandler returns the handler associated with the provider
func (p *BaseProvider) GetHandler() *handler.Conplicity {
	return p.handler
}

// GetBackupDir returns the backup directory used by the provider
func (p *BaseProvider) GetBackupDir() string {
	return p.backupDir
}
