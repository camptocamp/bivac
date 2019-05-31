package manager

import (
	"sort"
	"unicode/utf8"

	"github.com/camptocamp/bivac/internal/engine"
	"github.com/camptocamp/bivac/pkg/volume"
)

func retrieveVolumes(m *Manager, volumeFilters volume.Filters) (err error) {
	volumes, err := m.Orchestrator.GetVolumes(volume.Filters{})
	if err != nil {
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
	var volumeManaged bool
	for _, nv := range newVolumes {
		volumeManaged = false
		for _, mv := range m.Volumes {
			if mv.ID == nv.ID {
				volumeManaged = true
				break
			}
		}
		if !volumeManaged {
			nv.SetupMetrics()
			getLastBackupDate(m, nv)
			m.Volumes = append(m.Volumes, nv)
		}
	}

	// Remove deleted volumes
	var vols []*volume.Volume
	for _, mv := range m.Volumes {
		volumeExists := false
		for _, nv := range newVolumes {
			if mv.ID == nv.ID {
				volumeExists = true
				break
			}
		}
		if volumeExists {
			vols = append(vols, mv)
		} else {
			mv.CleanupMetrics()
			mv = nil
		}
	}

	m.Volumes = vols
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

	if l := volumeFilters.Blacklist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return true, "blacklisted", "blacklist config"
		}
	}
	return false, "", ""
}

func getLastBackupDate(m *Manager, v *volume.Volume) (err error) {
	e := &engine.Engine{
		DefaultArgs: []string{
			"--no-cache",
			"--json",
			"-r",
			m.TargetURL + "/" + m.Orchestrator.GetPath(v) + "/" + v.Name,
		},
	}

	latestBackup, oldestBackup, err := e.GetBackupDates()
	if err != nil {
		return
	}

	v.LastBackupDate = latestBackup.Format("2006-01-02 15:04:05")
	v.LastBackupStatus = "Unknown"

	// Leads to several flaws, should be improved
	v.Metrics.LastBackupDate.Set(float64(latestBackup.Unix()))

	v.Metrics.OldestBackupDate.Set(float64(oldestBackup.Unix()))

	// Unknown status
	v.Metrics.LastBackupStatus.Set(-1)
	return
}
