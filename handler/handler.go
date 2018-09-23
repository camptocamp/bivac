package handler

import (
	"errors"
	"fmt"
	"os"

	"github.com/camptocamp/bivac/config"

	log "github.com/Sirupsen/logrus"
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
	//c.Config = config.LoadConfig(version)

	//err = c.setupLogLevel()
	//log.Fatalf("Failed to setup log level: %v", err)

	//err = c.GetHostname()
	//log.Fatalf("Failed to get hostname: %v", err)
	return
}

// GetHostname gets the host name
func (c *Bivac) GetHostname() (err error) {
	c.Hostname, err = os.Hostname()
	return
}

func (c *Bivac) setupLogLevel() (err error) {
	switch c.Config.LogLevel {
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
		errMsg := fmt.Sprintf("Wrong log level '%v'", c.Config.LogLevel)
		err = errors.New(errMsg)
	}

	if c.Config.JSON {
		log.SetFormatter(&log.JSONFormatter{})
	}
	return
}
