package manager

import (
	"sort"
	"unicode/utf8"

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
	for _, v := range volumes {
		b, _, _ := blacklistedVolume(v, volumeFilters)
		if !b {
			newVolumes = append(newVolumes, v)
		}
	}

	// Append new volumes
	volumeManaged := false
	for _, nv := range newVolumes {
		volumeManaged = false
		for _, mv := range m.Volumes {
			if mv.ID == nv.ID {
				volumeManaged = true
				break
			}
		}
		if !volumeManaged {
			m.Volumes = append(m.Volumes, nv)
		}
	}

	// Remove deleted volumes
	for mk, mv := range m.Volumes {
		volumeExists := false
		for _, nv := range newVolumes {
			if mv.ID == nv.ID {
				volumeExists = true
				break
			}
		}
		if !volumeExists {
			m.Volumes = append(m.Volumes[:mk], m.Volumes[mk+1:]...)
		}
	}

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
