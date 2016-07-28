package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"unicode/utf8"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

// Conplicity is the main handler struct
type Conplicity struct {
	*docker.Client
	Config   *config.Config
	Hostname string
	Metrics  []string
}

// NewConplicity returns a new Conplicity handler
func NewConplicity(version string) (*Conplicity, error) {
	c := &Conplicity{}
	err := c.Setup(version)
	return c, err
}

// Setup sets up a Conplicity struct
func (c *Conplicity) Setup(version string) (err error) {
	c.Config = config.LoadConfig(version)

	err = c.setupLoglevel()
	util.CheckErr(err, "Failed to setup log level: %v", "fatal")

	err = c.GetHostname()
	util.CheckErr(err, "Failed to get hostname: %v", "fatal")

	err = c.SetupDocker()
	util.CheckErr(err, "Failed to setup docker: %v", "fatal")

	return
}

// GetHostname gets the host name
func (c *Conplicity) GetHostname() (err error) {
	if c.Config.HostnameFromRancher {
		resp, err := http.Get("http://rancher-metadata/latest/self/host/name")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		c.Hostname = string(body)
	} else {
		c.Hostname, err = os.Hostname()
	}
	return
}

// SetupDocker for the  client
func (c *Conplicity) SetupDocker() (err error) {
	c.Client, err = docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	util.CheckErr(err, "Failed to create Docker client: %v", "fatal")
	return
}

// GetVolumes returns the Docker volumes, inspected and filtered
func (c *Conplicity) GetVolumes() (volumes []*volume.Volume, err error) {
	vols, err := c.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		err = fmt.Errorf("Failed to list Docker volumes: %v", err)
		return
	}
	for _, vol := range vols.Volumes {
		var voll types.Volume
		voll, err = c.VolumeInspect(context.Background(), vol.Name)
		if err != nil {
			err = fmt.Errorf("Failed to inspect volume %s: %v", vol.Name, err)
			return
		}
		if b, r, s := c.blacklistedVolume(&voll); b {
			log.WithFields(log.Fields{
				"volume": vol.Name,
				"reason": r,
				"source": s,
			}).Info("Ignoring volume")
			continue
		}
		v := volume.NewVolume(vol)
		volumes = append(volumes, v)
	}
	return
}

func (c *Conplicity) blacklistedVolume(vol *types.Volume) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	list := c.Config.VolumesBlacklist
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}

	if ignoreLbl, _ := util.GetVolumeLabel(vol, ".ignore"); ignoreLbl == "true" {
		return true, "blacklisted", "volume label"
	}

	return false, "", ""
}

func (c *Conplicity) setupLoglevel() (err error) {
	switch c.Config.Loglevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		errMsg := fmt.Sprintf("Wrong log level '%v'", c.Config.Loglevel)
		err = errors.New(errMsg)
	}

	if c.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return
}
