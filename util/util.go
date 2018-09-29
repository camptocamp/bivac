package util

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
)

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

// Retry retry on error
func Retry(attempts int, callback func() error) (err error) {
	for i := 0; ; i++ {
		err = callback()
		if err == nil {
			return nil
		}

		if i >= (attempts - 1) {
			break
		}

		time.Sleep(2 * time.Second)

		log.Println("retrying...")
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}
