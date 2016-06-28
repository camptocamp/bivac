package util

import (
	"errors"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
)

const labelPrefix string = "io.conplicity"

// CheckErr checks for error, logs and optionally exits the program
func CheckErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
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
