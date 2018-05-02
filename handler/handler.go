package handler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/config"
	"github.com/camptocamp/bivac/util"
	"github.com/camptocamp/bivac/volume"
)

// Bivac is the main handler struct
type Bivac struct {
	Config   *config.Config
	Hostname string
}

// NewBivac returns a new Bivac handler
func NewBivac(version string) (*Bivac, error) {
	c := &Bivac{}
	err := c.Setup(version)
	return c, err
}

// Setup sets up a Bivac struct
func (c *Bivac) Setup(version string) (err error) {
	c.Config = config.LoadConfig(version)

	err = c.setupLoglevel()
	util.CheckErr(err, "Failed to setup log level: %v", "fatal")

	err = c.GetHostname()
	util.CheckErr(err, "Failed to get hostname: %v", "fatal")

	return
}

// GetHostname gets the host name
func (c *Bivac) GetHostname() (err error) {
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

// IsCheckScheduled checks if the backup must be verified
func (c *Bivac) IsCheckScheduled(vol *volume.Volume) (bool, error) {
	logCheckPath := vol.Mountpoint + "/.bivac_last_check"

	if vol.Config.NoVerify {
		log.WithFields(log.Fields{
			"volume": vol.Name,
		}).Info("Skipping verification")

		return false, nil
	}

	if _, err := os.Stat(logCheckPath); os.IsNotExist(err) {
		os.OpenFile(logCheckPath, os.O_RDONLY|os.O_CREATE, 0644)
	}

	info, err := os.Stat(logCheckPath)
	if err != nil {
		log.WithFields(log.Fields{
			"volume": vol.Name,
		}).Warning("Cannot retrieve the last check date, skipping verification")
		return false, nil
	}

	checkEvery, err := time.ParseDuration(c.Config.CheckEvery)
	if err != nil {
		err = fmt.Errorf("failed to parse the parameter 'check-every': %v", err)
		return false, err
	}

	checkExpiration := info.ModTime().Add(checkEvery)
	if time.Now().Before(checkExpiration) {
		return false, nil
	}

	log.WithFields(log.Fields{
		"volume": vol.Name,
	}).Info("Verifying backup")

	return true, nil
}

func (c *Bivac) setupLoglevel() (err error) {
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
