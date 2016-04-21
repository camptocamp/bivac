package providers

import (
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

type Provider interface {
	GetName() string
	PrepareBackup() error
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

func BackupVolume(c *handler.Conplicity, vol *docker.Volume) (err error) {
	// TODO: detect if it's a Database volume (PostgreSQL, MySQL, OpenLDAP...) and launch DUPLICITY_PRECOMMAND instead of backuping the volume
	log.Infof("ID: " + vol.Name)
	log.Infof("Driver: " + vol.Driver)
	log.Infof("Mountpoint: " + vol.Mountpoint)

	backupDir := ""
	// p.backupDir?

	log.Infof("Creating duplicity container...")

	fullIfOlderThan := getVolumeLabel(vol, ".full_if_older_than")
	if fullIfOlderThan == "" {
		fullIfOlderThan = c.FullIfOlderThan
	}

	pathSeparator := "/"
	if strings.HasPrefix(c.DuplicityTargetURL, "swift://") {
		// Looks like I'm not the one to fall on this issue: http://stackoverflow.com/questions/27991960/upload-to-swift-pseudo-folders-using-duplicity
		pathSeparator = "_"
	}

	container, err := c.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Cmd: []string{
					"--full-if-older-than", fullIfOlderThan,
					"--s3-use-new-style",
					"--no-encryption",
					"--allow-source-mismatch",
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

	checkErr(err, "Failed to create container for volume "+vol.Name+": %v", 1)

	defer func() {
		c.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
	}()

	binds := []string{
		vol.Name + ":" + vol.Mountpoint + ":ro",
	}

	err = dockerpty.Start(c.Client, container, &docker.HostConfig{
		Binds: binds,
	})
	checkErr(err, "Failed to start container for volume "+vol.Name+": %v", -1)
	return
	return nil
}

func checkErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
		}
	}
}

func getVolumeLabel(vol *docker.Volume, key string) (value string) {
	value = vol.Labels[labelPrefix+key]
	return
}
