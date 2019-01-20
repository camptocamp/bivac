package manager

import (
	//"fmt"
	"sort"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	//"github.com/camptocamp/bivac/pkg/orchestrators"
	"github.com/camptocamp/bivac/pkg/volume"
)

func retrieveVolumes(m *Manager, volumeFilters volume.Filters) (err error) {
	volumes, err := m.Orchestrator.GetVolumes(volume.Filters{})
	if err != nil {
		// Do we really want to cleanup volume list if an error occurs?
		m.Volumes = []*volume.Volume{}
		return
	}

	var newVolumes []*volume.Volume
	for _, volume := range volumes {
		b, r, s := blacklistedVolume(volume, volumeFilters)
		if b {
			log.WithFields(log.Fields{
				"volume": volume.Name,
				"reason": r,
				"source": s,
			}).Debugf("Ignoring volume")
			continue
		}
		newVolumes = append(newVolumes, volume)
	}

	m.Volumes = newVolumes
	return
}

func blacklistedVolume(vol *volume.Volume, volumeFilters volume.Filters) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	// Use whitelist if defined
	if l := volumeFilters.Whitelist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return false, "", ""
		}
		return true, "blacklisted", "whitelist config"
	}

	i := sort.SearchStrings(volumeFilters.Blacklist, vol.Name)
	if i < len(volumeFilters.Blacklist) && volumeFilters.Blacklist[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}
	return false, "", ""
}
