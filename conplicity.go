package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/providers"
)

const labelPrefix string = "io.conplicity"

func main() {
	log.Infof("Starting backup...")

	var err error

	c := &handler.Conplicity{}

	c.GetEnv()

	c.Hostname, err = os.Hostname()
	checkErr(err, "Failed to get hostname: %v", 1)

	endpoint := "unix:///var/run/docker.sock"

	c.Client, err = docker.NewClient(endpoint)
	checkErr(err, "Failed to create Docker client: %v", 1)

	vols, err := c.ListVolumes(docker.ListVolumesOptions{})
	checkErr(err, "Failed to list Docker volumes: %v", 1)

	err = c.PullImage()
	checkErr(err, "Failed to pull image: %v", 1)

	for _, vol := range vols {
		voll, err := c.InspectVolume(vol.Name)
		checkErr(err, "Failed to inspect volume "+vol.Name+": %v", -1)
		p := providers.GetProvider(c, voll)
		err = p.PrepareBackup()
		checkErr(err, "Failed to prepare backup for volume "+vol.Name+": %v", -1)
		err = p.BackupVolume()
		checkErr(err, "Failed to backup volume "+vol.Name+": %v", -1)

		//err = c.backupVolume(voll)
		//checkErr(err, "Failed to process volume "+vol.Name+": %v", -1)
	}

	log.Infof("End backup...")
}

func checkErr(err error, msg string, exit int) {
	if err != nil {
		log.Errorf(msg, err)

		if exit != -1 {
			os.Exit(exit)
		}
	}
}
