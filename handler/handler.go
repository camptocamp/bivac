package handler

import (
	"bytes"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/util"
	"github.com/fgrehm/go-dockerpty"
	"github.com/fsouza/go-dockerclient"
	"github.com/jessevdk/go-flags"
)

type environment struct {
	Image              string   `short:"i" long:"image" description:"The duplicity docker image." env:"DUPLICITY_DOCKER_IMAGE" default:"camptocamp/duplicity:latest"`
	DuplicityTargetURL string   `short:"u" long:"url" description:"The duplicity target URL to push to." env:"DUPLICITY_TARGET_URL"`
	PushgatewayURL     string   `short:"g" long:"gateway-url" description:"The prometheus push gateway URL to use." env:"PUSHGATEWAY_URL"`
	AWSAccessKeyID     string   `long:"aws-access-key-id" description:"The AWS access key ID." env:"AWS_ACCESS_KEY_ID"`
	AWSSecretAccessKey string   `long:"aws-secret-key-id" description:"The AWS secret access key." env:"AWS_SECRET_ACCESS_KEY"`
	SwiftUsername      string   `long:"swift-username" description:"The Swift user name." env:"SWIFT_USERNAME"`
	SwiftPassword      string   `long:"swift-password" description:"The Swift password." env:"SWIFT_PASSWORD"`
	SwiftAuthURL       string   `long:"swift-auth_url" description:"The Swift auth URL." env:"SWIFT_AUTHURL"`
	SwiftTenantName    string   `long:"swift-tenant-name" description:"The Swift tenant name." env:"SWIFT_TENANTNAME"`
	SwiftRegionName    string   `long:"swift-region-name" description:"The Swift region name." env:"SWIFT_REGIONNAME"`
	FullIfOlderThan    string   `long:"full-if-older-than" description:"The number of days after which a full backup must be performed." env:"CONPLICITY_FULL_IF_OLDER_THAN" default:"15D"`
	RemoveOlderThan    string   `long:"remove-older-than" description:"The number days after which backups must be removed." env:"CONPLICITY_REMOVE_OLDER_THAN" default:"30D"`
	VolumesBlacklist   []string `short:"b" long:"blacklist" description:"Volumes to blacklist in backups." env:"CONPLICITY_VOLUMES_BLACKLIST"`
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
	parser := flags.NewParser(c.environment, flags.Default)
	if _, err = parser.Parse(); err != nil {
		os.Exit(1)
	}
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
func (c *Conplicity) LaunchDuplicity(cmd []string, binds []string) (state docker.State, stdout string, err error) {
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

	var stdoutBuffer bytes.Buffer
	opts := docker.LogsOptions{
		Container:    container.ID,
		OutputStream: &stdoutBuffer,
		Stdout:       true,
		RawTerminal:  true,
	}
	err = c.Logs(opts)
	util.CheckErr(err, "Failed to retrieve logs: %v", -1)

	state = container.State
	stdout = stdoutBuffer.String()

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
