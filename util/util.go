package util

import (
	"fmt"
	"regexp"
	"strconv"
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

// GetDurationFromInterval takes a formated interval and returns a datetime
func GetDurationFromInterval(interval string) (duration time.Duration, err error) {

	r, err := regexp.Compile("([0-9]{1,}[mhDWMY])")
	if err != nil {
		return
	}

	for _, v := range r.FindAllStringSubmatch(interval, -1) {

		i := v[0]

		iValue, err := strconv.Atoi(i[:len(i)-1])
		if err != nil {
			return duration, err
		}
		iUnit := i[len(i)-1:]

		switch iUnit {
		case "D":
			duration += time.Hour * 24 * time.Duration(iValue)
		case "W":
			duration += time.Hour * 24 * 7 * time.Duration(iValue)
		case "M":
			duration += time.Hour * 24 * 30 * time.Duration(iValue)
		case "Y":
			duration += time.Hour * 24 * 365 * time.Duration(iValue)
		default:
			d, err := time.ParseDuration(i)
			if err != nil {
				return duration, err
			}
			duration += d
		}
	}
	return
}
