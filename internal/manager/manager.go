package manager

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/internal/utils"
	"github.com/camptocamp/bivac/pkg/orchestrators"
	"github.com/camptocamp/bivac/pkg/volume"
)

// Orchestrators groups the parameters of all supported orchestrators in one structure
type Orchestrators struct {
	Docker     orchestrators.DockerConfig
	Cattle     orchestrators.CattleConfig
	Kubernetes orchestrators.KubernetesConfig
}

// Manager contains all informations used by the Bivac manager
type Manager struct {
	Orchestrator orchestrators.Orchestrator
	Volumes      []*volume.Volume
	Server       *Server
	Providers    *Providers
	TargetURL    string
	RetryCount   int
	LogServer    string
	BuildInfo    utils.BuildInfo
	AgentImage   string

	backupSlots chan *volume.Volume
}

// Start starts a Bivac manager which handle backups management
func Start(buildInfo utils.BuildInfo, o orchestrators.Orchestrator, s Server, volumeFilters volume.Filters, providersFile, targetURL, logServer, agentImage string, retryCount, parallelCount int, refreshRate, backupInterval string, backupTimeSpec string) (err error) {
	p, err := LoadProviders(providersFile)
	if err != nil {
		err = fmt.Errorf("failed to read providers file: %s", err)
		return
	}

	refreshInterval, err := time.ParseDuration(refreshRate)
	if err != nil {
		err = fmt.Errorf("failed to parse refresh time: %s", err)
		return
	}

  //check backupTimeSpec format and constraint 00h < xxh < 23h e.g: 23h59m does not respect constraint
  backupTime, err := time.ParseDuration(backupTimeSpec)
	if err != nil {
		err = fmt.Errorf("failed to parse backup prefered time: %s", err)
		return
	}

  if backupTime.Hours() >= 23 && backupTime.Hours() >= 0 {
		err = fmt.Errorf("backup prefered time does not respect constraint (00h < x < 23h): %s", err)
		return
	}

	backupInt, err := time.ParseDuration(backupInterval)
	if err != nil {
		err = fmt.Errorf("failed to parse backup interval: %s", err)
		return
	}

	m := &Manager{
		Orchestrator: o,
		Server:       &s,
		Providers:    &p,
		TargetURL:    targetURL,
		RetryCount:   retryCount,
		LogServer:    logServer,
		BuildInfo:    buildInfo,
		AgentImage:   agentImage,

		backupSlots: make(chan *volume.Volume, 100),
	}

	// Catch orphan agents
	orphanAgents, err := m.Orchestrator.RetrieveOrphanAgents()
	if err != nil {
		log.Errorf("failed to retrieve orphan agents: %s", err)
	}

	// Manage volumes
	go func(m *Manager, volumeFilters volume.Filters) {

		log.Debugf("Starting volume manager...")

		for {
			err = retrieveVolumes(m, volumeFilters)
			if err != nil {
				log.Errorf("failed to retrieve volumes: %s", err)
			}

			for _, v := range m.Volumes {
				if val, ok := orphanAgents[v.ID]; ok {
					v.BackingUp = true
					go m.attachOrphanAgent(val, v)
					delete(orphanAgents, val)
				}

				if !isBackupNeeded(v, backupInt, backupTime) {
					continue
				}

				m.backupSlots <- v
			}

			time.Sleep(refreshInterval)
		}
	}(m, volumeFilters)

	// Manage backups
	go func(m *Manager, parallelCount int) {
		slots := make(map[string](chan bool))

		log.Infof("Starting backup manager...")

		for {
			v := <-m.backupSlots
			if _, ok := slots[v.HostBind]; !ok {
				slots[v.HostBind] = make(chan bool, parallelCount)
			}
			select {
			case slots[v.HostBind] <- true:
			default:
				continue
			}
			if ok, _ := m.Orchestrator.IsNodeAvailable(v.HostBind); !ok && v.HostBind != "unbound" && m.Orchestrator.GetName() == "cattle" {
				log.WithFields(log.Fields{
					"node": v.HostBind,
				}).Warning("Node unavailable.")
				<-slots[v.HostBind]
				continue
			}

			go func(v *volume.Volume) {
				var timedout bool
				tearDown := make(chan bool)

				log.WithFields(log.Fields{
					"volume":   v.Name,
					"hostname": v.Hostname,
				}).Debugf("Backing up volume.")
				defer func() {
					if !timedout {
						tearDown <- true
						<-slots[v.HostBind]
					}
				}()

				// Workaround which avoid a stucked backup to block the whole backup process
				// If the backup process takes more than one hour,
				// the backup slot is released.
				go func() {
					timeout := time.After(1 * time.Hour)
					select {
					case <-tearDown:
						return
					case <-timeout:
						timedout = true
						<-slots[v.HostBind]
					}
				}()

				err = nil
				for i := 0; i <= m.RetryCount; i++ {
					err = backupVolume(m, v, false)
					if err != nil {
						log.WithFields(log.Fields{
							"volume":   v.Name,
							"hostname": v.Hostname,
							"try":      i + 1,
						}).Errorf("failed to backup volume: %s", err)

						time.Sleep(2 * time.Second)
					} else {
						break
					}
				}
			}(v)
		}
	}(m, parallelCount)

	// Manage API server
	m.StartServer()

	return
}

func isBackupNeeded(v *volume.Volume, backupInt time.Duration,backupTime time.Duration) bool {
	if v.BackingUp {
		return false
	}

	if v.LastBackupDate == "" {
		return true
	}

	var dateRef string
	if v.LastBackupStartDate == "" {
		dateRef = v.LastBackupDate
	} else {
		dateRef = v.LastBackupStartDate
	}

	lbd, err := time.Parse("2006-01-02 15:04:05", dateRef)
	if err != nil {
		log.WithFields(log.Fields{
			"volume":   v.Name,
			"hostname": v.Hostname,
		}).Errorf("failed to parse backup date of volume `%s': %s", v.Name, err)
		return false
	}

  // retry if last backup status is Failed
	if lbd.Add(time.Hour).Before(time.Now().UTC()) && v.LastBackupStatus == "Failed" {
		return true
	}

  // convert targetBackupTime in date with today time
  now := time.Now().UTC()
  targetBackupTime := time.Date(now.Year(), now.Month(), now.Day(),0,0,0,0,now.Location()).Add(backupTime)

  // set backup interval to 23h if backupTime is not set to default value
  if backupTime.Seconds() > 0 && targetBackupTime.Before(now) {
    fixInterval, _ := time.ParseDuration("23h")
	  if lbd.Add(fixInterval).Before(now) {
      return true
    }
  }

  // if backupTime is set to default value
	if lbd.Add(backupInt).Before(now) && backupTime.Seconds() == 0 {
		return true
	}
	return false
}

// GetOrchestrator returns an orchestrator interface based on the name you specified or on the orchestrator Bivac is running on
func GetOrchestrator(name string, orchs Orchestrators) (o orchestrators.Orchestrator, err error) {
	if name != "" {
		log.Debugf("Choosing orchestrator based on configuration...")
		switch name {
		case "docker":
			o, err = orchestrators.NewDockerOrchestrator(&orchs.Docker)
		case "cattle":
			o, err = orchestrators.NewCattleOrchestrator(&orchs.Cattle)
		case "kubernetes":
			o, err = orchestrators.NewKubernetesOrchestrator(&orchs.Kubernetes)
		default:
			err = fmt.Errorf("'%s' is not a valid orchestrator", name)
			return
		}
	} else {
		log.Debugf("Trying to detect orchestrator based on environment...")

		if orchestrators.DetectCattle() {
			o, err = orchestrators.NewCattleOrchestrator(&orchs.Cattle)
		} else if orchestrators.DetectKubernetes() {
			o, err = orchestrators.NewKubernetesOrchestrator(&orchs.Kubernetes)
		} else if orchestrators.DetectDocker(&orchs.Docker) {
			o, err = orchestrators.NewDockerOrchestrator(&orchs.Docker)
		} else {
			err = fmt.Errorf("no orchestrator detected")
			return
		}
	}
	if err != nil {
		log.Infof("Using orchestrator: %s", o.GetName())
	}
	return
}

// BackupVolume does a backup of a volume
func (m *Manager) BackupVolume(volumeID string, force bool) (err error) {
	for _, v := range m.Volumes {
		if v.ID == volumeID {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Debug("Backup manually requested.")
			err = backupVolume(m, v, force)
			if err != nil {
				err = fmt.Errorf("failed to backup volume: %s", err)
				return
			}
		}
	}
	return
}

// RestoreVolume does a restore of a volume
func (m *Manager) RestoreVolume(
	volumeID string,
	force bool,
	snapshotName string,
) (err error) {
	for _, v := range m.Volumes {
		if v.ID == volumeID {
			log.WithFields(log.Fields{
				"volume":   v.Name,
				"hostname": v.Hostname,
			}).Debug("Restore manually requested.")
			err = restoreVolume(m, v, force, snapshotName)
			if err != nil {
				err = fmt.Errorf(
					"failed to restore volume: %s",
					err,
				)
				return
			}
		}
	}
	return
}

// GetInformations returns informations regarding the Bivac manager
func (m *Manager) GetInformations() (informations map[string]string) {
	informations = map[string]string{
		"version":        m.BuildInfo.Version,
		"build_date":     m.BuildInfo.Date,
		"build_commit":   m.BuildInfo.CommitSha1,
		"golang_version": m.BuildInfo.Runtime,
		"orchestrator":   m.Orchestrator.GetName(),
		"address":        m.Server.Address,
		"volumes_count":  fmt.Sprintf("%d", len(m.Volumes)),
	}
	return
}
