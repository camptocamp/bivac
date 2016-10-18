package util

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	docker "github.com/docker/docker/client"
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
	//log.Debugf("Getting value for label %s of volume %s", key, vol.Name)
	value, ok := vol.Labels[labelPrefix+"."+key]
	if !ok {
		errMsg := fmt.Sprintf("Key %v not found in labels for volume %v", key, vol.Name)
		err = errors.New(errMsg)
	}
	return
}

// PullImage pulls an image from the registry
func PullImage(c *docker.Client, image string) (err error) {
	if _, _, err = c.ImageInspectWithRaw(context.Background(), image); err != nil {
		// TODO: output pull to logs
		log.WithFields(log.Fields{
			"image": image,
		}).Info("Pulling image")
		resp, err := c.ImagePull(context.Background(), image, types.ImagePullOptions{})
		if err != nil {
			log.Errorf("ImagePull returned an error: %v", err)
			return err
		}
		defer resp.Close()
		body, err := ioutil.ReadAll(resp)
		if err != nil {
			log.Errorf("Failed to read from ImagePull response: %v", err)
			return err
		}
		log.Debugf("Pull image response body: %v", string(body))
	} else {
		log.WithFields(log.Fields{
			"image": image,
		}).Debug("Image already pulled, not pulling")
	}

	return nil
}

// RemoveContainer removes a container
func RemoveContainer(c *docker.Client, id string) {
	log.WithFields(log.Fields{
		"container": id,
	}).Debug("Removing container")
	err := c.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	CheckErr(err, "Failed to remove container "+id+": %v", "error")
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
