package handler

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/caarlos0/env"
	"github.com/fsouza/go-dockerclient"
)

type Environment struct {
	Image              string `env:"DUPLICITY_DOCKER_IMAGE" envDefault:"camptocamp/duplicity:latest"`
	DuplicityTargetURL string `env:"DUPLICITY_TARGET_URL"`
	AWSAccessKeyID     string `env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string `env:"AWS_SECRET_ACCESS_KEY"`
	SwiftUsername      string `env:"SWIFT_USERNAME"`
	SwiftPassword      string `env:"SWIFT_PASSWORD"`
	SwiftAuthURL       string `env:"SWIFT_AUTHURL"`
	SwiftTenantName    string `env:"SWIFT_TENANTNAME"`
	SwiftRegionName    string `env:"SWIFT_REGIONNAME"`
	FullIfOlderThan    string `env:"FULL_IF_OLDER_THAN" envDefault:"15D"`
}

type Conplicity struct {
	*docker.Client
	*Environment
	Hostname string
}

func (c *Conplicity) Setup() (err error) {
	c.GetEnv()

	c.Hostname, err = os.Hostname()
	checkErr(err, "Failed to get hostname: %v", 1)

	endpoint := "unix:///var/run/docker.sock"

	c.Client, err = docker.NewClient(endpoint)
	checkErr(err, "Failed to create Docker client: %v", 1)

	err = c.PullImage()
	checkErr(err, "Failed to pull image: %v", 1)

	return
}

func (c *Conplicity) GetEnv() (err error) {
	c.Environment = &Environment{}
	env.Parse(c.Environment)

	return
}

func (c *Conplicity) PullImage() (err error) {
	if _, err = c.InspectImage(c.Image); err != nil {
		// TODO: output pull to logs
		log.Infof("Pulling image %v", c.Image)
		err = c.Client.PullImage(docker.PullImageOptions{
			Repository: c.Image,
		}, docker.AuthConfiguration{})
	}

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
