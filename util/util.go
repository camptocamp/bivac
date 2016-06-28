package util

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
)

const labelPrefix string = "io.conplicity"

// CheckErr checks for error, logs and optionally exits the program
func CheckErr(err error, msg string, level string) {
	if err != nil {
		switch level {
		case "debug":
			log.Debugf(msg, err)
		case "info":
			log.Infof(msg, err)
		case "warn":
			log.Warnf(msg, err)
		case "error":
			log.Errorf(msg, err)
		case "fatal":
			log.Fatalf(msg, err)
		case "panic":
			log.Panicf(msg, err)
		default:
			log.Panicf("Wrong loglevel '%v', please report this bug", level)
		}
	}
}

// GetVolumeLabel retrieves the value of given key in the io.conplicity
// namespace of the volume labels
func GetVolumeLabel(vol *types.Volume, key string) (value string, err error) {
	value, ok := vol.Labels[labelPrefix+key]
	if !ok {
		errMsg := fmt.Sprintf("Key %v not found in labels for volume %v", key, vol.Name)
		err = errors.New(errMsg)
	}
	return
}
