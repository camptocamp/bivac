package main

import (
	"sort"
	"unicode/utf8"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"

	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/providers"
	"github.com/camptocamp/conplicity/util"
)

var version = "undefined"

func main() {
	var err error

	c := &handler.Conplicity{}
	err = c.Setup(version)
	util.CheckErr(err, "Failed to setup Conplicity handler: %v", "panic")

	log.Infof("Starting backup...")

	vols, err := c.VolumeList(context.Background(), filters.NewArgs())
	util.CheckErr(err, "Failed to list Docker volumes: %v", "panic")

	for _, vol := range vols.Volumes {
		voll, err := c.VolumeInspect(context.Background(), vol.Name)
		util.CheckErr(err, "Failed to inspect volume "+vol.Name+": %v", "fatal")

		err = backupVolume(c, voll)
		util.CheckErr(err, "Failed to process volume "+vol.Name+": %v", "fatal")
	}

	err = c.PushToPrometheus()
	util.CheckErr(err, "Failed post data to Prometheus Pushgateway: %v", "fatal")

	log.Infof("End backup...")
}

func backupVolume(c *handler.Conplicity, vol types.Volume) (err error) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" {
		log.WithFields(log.Fields{
			"volume": vol.Name,
			"reason": "unnamed",
		}).Info("Ignoring volume")
		return
	}

	list := c.Config.VolumesBlacklist
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		log.WithFields(log.Fields{
			"volume": vol.Name,
			"reason": "blacklisted",
			"source": "blacklist config",
		}).Info("Ignoring volume")
		return
	}

	if ignoreLbl, _ := util.GetVolumeLabel(&vol, ".ignore"); ignoreLbl == "true" {
		log.WithFields(log.Fields{
			"volume": vol.Name,
			"reason": "blacklisted",
			"source": "volume label",
		}).Info("Ignoring volume")
		return
	}

	p := providers.GetProvider(c, &vol)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found provider")
	err = providers.PrepareBackup(p)
	util.CheckErr(err, "Failed to prepare backup for volume "+vol.Name+": %v", "fatal")
	err = p.BackupVolume(&vol)
	util.CheckErr(err, "Failed to backup volume "+vol.Name+": %v", "fatal")
	return
}
