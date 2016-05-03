package handler

import (
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"github.com/camptocamp/conplicity/util"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
)

type environment struct {
	Image              string   `env:"DUPLICITY_DOCKER_IMAGE" envDefault:"camptocamp/duplicity:latest"`
	DuplicityTargetURL string   `env:"DUPLICITY_TARGET_URL"`
	AWSAccessKeyID     string   `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string   `env:"AWS_SECRET_ACCESS_KEY"`
	SwiftUsername      string   `env:"SWIFT_USERNAME"`
	SwiftPassword      string   `env:"SWIFT_PASSWORD"`
	SwiftAuthURL       string   `env:"SWIFT_AUTHURL"`
	SwiftTenantName    string   `env:"SWIFT_TENANTNAME"`
	SwiftRegionName    string   `env:"SWIFT_REGIONNAME"`
	FullIfOlderThan    string   `env:"CONPLICITY_FULL_IF_OLDER_THAN" envDefault:"15D"`
	RemoveOlderThan    string   `env:"CONPLICITY_REMOVE_OLDER_THAN" envDefault:"30D"`
	VolumesBlacklist   []string `env:"CONPLICITY_VOLUMES_BLACKLIST"`
}

// Conplicity is the main handler struct
type Conplicity struct {
	*docker.Client
	*environment
	Hostname string
}

// Setup sets up a Conplicity struct
func (c *Conplicity) Setup() (err error) {
	c.getEnv()

	c.Hostname, err = os.Hostname()
	util.CheckErr(err, "Failed to get hostname: %v", 1)

	endpoint := "unix:///var/run/docker.sock"

	c.Client, err = docker.NewClient(endpoint)
	util.CheckErr(err, "Failed to create Docker client: %v", 1)

	err = c.pullImage()
	util.CheckErr(err, "Failed to pull image: %v", 1)

	return
}

func (c *Conplicity) getEnv() (err error) {
	c.environment = &environment{}
	env.Parse(c.environment)
	sort.Strings(c.VolumesBlacklist)
	return
}

func (c *Conplicity) pullImage() (err error) {
	if _, err = c.InspectImage(c.Image); err != nil {
		// TODO: output pull to logs
		log.Infof("Pulling image %v", c.Image)
		err = c.Client.PullImage(docker.PullImageOptions{
			Repository: c.Image,
		}, docker.AuthConfiguration{})
	}

	return err
}

// LaunchDuplicity starts a duplicity container with given command and binds
func (c *Conplicity) LaunchDuplicity(cmd []string, binds []string) (err error) {
	env := []string{
		"AWS_ACCESS_KEY_ID=" + c.AWSAccessKeyID,
		"AWS_SECRET_ACCESS_KEY=" + c.AWSSecretAccessKey,
		"SWIFT_USERNAME=" + c.SwiftUsername,
		"SWIFT_PASSWORD=" + c.SwiftPassword,
		"SWIFT_AUTHURL=" + c.SwiftAuthURL,
		"SWIFT_TENANTNAME=" + c.SwiftTenantName,
		"SWIFT_REGIONNAME=" + c.SwiftRegionName,
		"SWIFT_AUTHVERSION=2",
	}

	container, err := c.CreateContainer(
		docker.CreateContainerOptions{
			Config: &docker.Config{
				Cmd:          cmd,
				Env:          env,
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
	util.CheckErr(err, "Failed to create container: %v", 1)
	defer c.removeContainer(container)

	log.Infof("Launching 'duplicity %v'...", strings.Join(cmd, " "))
	err = dockerpty.Start(c.Client, container, &docker.HostConfig{
		Binds: binds,
	})
	util.CheckErr(err, "Failed to start container: %v", -1)
	return
}

func (c *Conplicity) removeContainer(cont *docker.Container) {
	log.Infof("Removing container %v...", cont.ID)
	c.RemoveContainer(docker.RemoveContainerOptions{
		ID:            cont.ID,
		Force:         true,
		RemoveVolumes: true,
	})
}
