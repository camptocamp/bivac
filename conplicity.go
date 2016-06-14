package main

import (
	"bytes"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"

	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/providers"
	"github.com/camptocamp/conplicity/util"
)

const labelPrefix string = "io.conplicity"

func main() {
	log.Infof("Starting backup...")

	var err error

	c := &handler.Conplicity{}
	err = c.Setup()
	util.CheckErr(err, "Failed to setup Conplicity handler: %v", 1)

	vols, err := c.ListVolumes(docker.ListVolumesOptions{})
	util.CheckErr(err, "Failed to list Docker volumes: %v", 1)

	var metrics []string
	for _, vol := range vols {
		voll, err := c.InspectVolume(vol.Name)
		util.CheckErr(err, "Failed to inspect volume "+vol.Name+": %v", -1)

		_metrics, err := backupVolume(c, voll)
		util.CheckErr(err, "Failed to process volume "+vol.Name+": %v", -1)
		metrics = append(metrics, _metrics...)
	}
	if c.PushgatewayURL != "" {
		data := strings.Join(metrics, "\n") + "\n"
		log.Infof("Sending metrics to Prometheus Pushgateway: %v", data)

		url := c.PushgatewayURL + "/metrics/job/conplicity/instance/" + c.Hostname
		log.Infof("URL=%v", url)

		req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data))
		req.Header.Set("Content-Type", "text/plain; version=0.0.4")

		client := &http.Client{}
		resp, err := client.Do(req)

		log.Infof("resp = %v", resp)
		util.CheckErr(err, "Failed post data to Prometheus Pushgateway: %v", 1)
	}

	log.Infof("End backup...")
}

func backupVolume(c *handler.Conplicity, vol *docker.Volume) (metrics []string, err error) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" {
		log.Infof("Ignoring unnamed volume " + vol.Name)
		return
	}

	list := c.VolumesBlacklist
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		log.Infof("Ignoring blacklisted volume " + vol.Name)
		return
	}

	if getVolumeLabel(vol, ".ignore") == "true" {
		log.Infof("Ignoring blacklisted volume " + vol.Name)
		return
	}

	p := providers.GetProvider(c, vol)
	log.Infof("Using provider %v to backup %v", p.GetName(), vol.Name)
	err = providers.PrepareBackup(p)
	util.CheckErr(err, "Failed to prepare backup for volume "+vol.Name+": %v", -1)
	metrics, err = providers.BackupVolume(p, vol)
	util.CheckErr(err, "Failed to backup volume "+vol.Name+": %v", -1)
	return
}

func getVolumeLabel(vol *docker.Volume, key string) (value string) {
	value = vol.Labels[labelPrefix+key]
	return
}
