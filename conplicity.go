package main

import (
	"sort"
	"unicode/utf8"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/engines"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/providers"
	"github.com/camptocamp/conplicity/util"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
)

var version = "undefined"

func main() {
	var err error

	c, err := handler.NewConplicity(version)
	util.CheckErr(err, "Failed to setup Conplicity handler: %v", "fatal")

	log.Infof("Conplity v%s starting backup...", version)

	vols, err := c.VolumeList(context.Background(), filters.NewArgs())
	util.CheckErr(err, "Failed to list Docker volumes: %v", "panic")

	for _, vol := range vols.Volumes {
		voll, err := c.VolumeInspect(context.Background(), vol.Name)
		util.CheckErr(err, "Failed to inspect volume "+vol.Name+": %v", "fatal")

		metrics, err := backupVolume(c, &voll)
		util.CheckErr(err, "Failed to process volume "+vol.Name+": %v", "fatal")
		c.Metrics = append(c.Metrics, metrics...)
	}

	err = c.PushToPrometheus()
	util.CheckErr(err, "Failed post data to Prometheus Pushgateway: %v", "fatal")

	log.Infof("End backup...")
}

func backupVolume(c *handler.Conplicity, vol *types.Volume) (metrics []string, err error) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" || vol.Name == "lost+found" {
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

	if ignoreLbl, _ := util.GetVolumeLabel(vol, ".ignore"); ignoreLbl == "true" {
		log.WithFields(log.Fields{
			"volume": vol.Name,
			"reason": "blacklisted",
			"source": "volume label",
		}).Info("Ignoring volume")
		return
	}

	v := volume.NewVolume(vol)

	p := providers.GetProvider(c, v)
	log.WithFields(log.Fields{
		"volume":   vol.Name,
		"provider": p.GetName(),
	}).Info("Found data provider")
	err = providers.PrepareBackup(p)
	util.CheckErr(err, "Failed to prepare backup for volume "+vol.Name+": %v", "fatal")

	e := engines.GetEngine(c, v)
	log.WithFields(log.Fields{
		"volume": vol.Name,
		"engine": e.GetName(),
	}).Info("Found backup engine")

	metrics, err = e.Backup()
	util.CheckErr(err, "Failed to backup volume "+vol.Name+": %v", "fatal")
	return
}
