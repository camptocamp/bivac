package util

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
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
func GetVolumeLabel(vol *docker.Volume, key string) (value string) {
	value = vol.Labels[labelPrefix+key]
	return
}
