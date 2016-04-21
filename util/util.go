package util

import (
  "os"

	log "github.com/Sirupsen/logrus"
)

// CheckErr checks for error, logs and optionally exits the program
func CheckErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
		}
	}
}
