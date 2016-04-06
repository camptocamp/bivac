package main

import (
	"os"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

type environment struct {
	Image              string
	DuplicityTargetURL string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	SwiftUsername      string
	SwiftPassword      string
	SwiftAuthURL       string
	SwiftTenantName    string
	SwiftRegionName    string
}

type conplicity struct {
	*docker.Client
	*environment
	Hostname string
}

func main() {
	log.Infof("Starting backup...")

	var err error

	c := &conplicity{}

	c.getEnv()

	c.Hostname, err = os.Hostname()
	checkErr(err, "Failed to get hostname: %v", 1)

	endpoint := "unix:///var/run/docker.sock"

	c.Client, err = docker.NewClient(endpoint)
	checkErr(err, "Failed to create Docker client: %v", 1)

	vols, err := c.ListVolumes(docker.ListVolumesOptions{})
	checkErr(err, "Failed to list Docker volumes: %v", 1)

	err = c.pullImage()
	checkErr(err, "Failed to pull image: %v", 1)

	for _, vol := range vols {
		err = c.backupVolume(vol)
		checkErr(err, "Failed to process volume "+vol.Name+": %v", -1)
	}

	log.Infof("End backup...")
}

func (c *conplicity) getEnv() (err error) {
	c.environment = &environment{
		Image:              os.Getenv("DUPLICITY_DOCKER_IMAGE"),
		DuplicityTargetURL: os.Getenv("DUPLICITY_TARGET_URL"),
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		SwiftUsername:      os.Getenv("SWIFT_USERNAME"),
		SwiftPassword:      os.Getenv("SWIFT_PASSWORD"),
		SwiftAuthURL:       os.Getenv("SWIFT_AUTHURL"),
		SwiftTenantName:    os.Getenv("SWIFT_TENANTNAME"),
		SwiftRegionName:    os.Getenv("SWIFT_REGIONNAME"),
	}

	if c.Image == "" {
		c.Image = "camptocamp/duplicity:latest"
	}

	return
}

func (c *conplicity) backupVolume(vol docker.Volume) (err error) {
	if utf8.RuneCountInString(vol.Name) == 64 {
		log.Infof("Ignoring volume " + vol.Name)
		return
	}

	// TODO: detect if it's a Database volume (PostgreSQL, MySQL, OpenLDAP...) and launch DUPLICITY_PRECOMMAND instead of backuping the volume
	log.Infof("ID: " + vol.Name)
	log.Infof("Driver: " + vol.Driver)
	log.Infof("Mountpoint: " + vol.Mountpoint)
	log.Infof("Creating duplicity container...")
	container, err := c.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Cmd: []string{
					"--full-if-older-than", "15D",
					"--s3-use-new-style",
					"--no-encryption",
					"--allow-source-mismatch",
					"/var/backups",
					c.DuplicityTargetURL + "/" + c.Hostname + "/" + vol.Name,
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
		vol.Mountpoint + ":/var/backups:ro",
	}

	err = dockerpty.Start(c.Client, container, &docker.HostConfig{
		Binds: binds,
	})
	checkErr(err, "Failed to start container for volume "+vol.Name+": %v", -1)
	return
}

func (c *conplicity) pullImage() (err error) {
	// TODO: output pull to logs
	log.Infof("Pulling image %v", c.Image)
	err = c.PullImage(docker.PullImageOptions{
		Repository: c.Image,
	}, docker.AuthConfiguration{})

	return err
}

func checkErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
		}
	}
}
