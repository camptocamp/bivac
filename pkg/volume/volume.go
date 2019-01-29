package volume

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Volume provides backup methods for a single volume
type Volume struct {
	ID         string
	Name       string
	BackupDir  string
	Mount      string
	Mountpoint string
	Driver     string
	Labels     map[string]string
	ReadOnly   bool
	HostBind   string
	Hostname   string
	Namespace  string

	LastBackupDate   string
	LastBackupStatus string
	Logs             map[string]string

	Metrics *Metrics `json:"-"`
}

// Filters hfcksdghfvd
type Filters struct {
	Whitelist []string
	Blacklist []string
}

// Metrics are used to fill the Prometheus endpoint
// TODO: Merge LastBackupDate and LastBackupStatus
type Metrics struct {
	LastBackupDate   prometheus.Gauge
	LastBackupStatus prometheus.Gauge
}

// MountedVolume stores mounted volumes inside a container
type MountedVolume struct {
	PodID       string
	ContainerID string
	HostID      string
	Volume      *Volume
	Path        string
}

func (v *Volume) SetupMetrics() {
	v.Metrics = &Metrics{}

	v.Metrics.LastBackupDate = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bivac_lastBackupDate",
		Help: "Date of the last backup",
		ConstLabels: map[string]string{
			"id":       v.ID,
			"volume":   v.Name,
			"hostbind": v.HostBind,
		},
	})
	v.Metrics.LastBackupStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "bivac_lastBackupStatus",
		Help: "Status of the last backup",
		ConstLabels: map[string]string{
			"id":       v.ID,
			"volume":   v.Name,
			"hostbind": v.HostBind,
		},
	})
	return
}
