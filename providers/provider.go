package providers

import (
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"

	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/volume"
)

// Providers stores the list of available providers
type Providers struct {
	Providers map[string]Provider
}

// Provider stores data for one provider
type Provider struct {
	Name         string `toml:"-"`
	PreCmd       string `toml:"pre_cmd"`
	PostCmd      string `toml:"post_cmd"`
	DetectionCmd string `toml:"detect_cmd"`
	BackupDir    string `toml:"backup_dir"`
}

type configToml struct {
	Providers map[string]Provider `toml:"providers"`
}

// LoadProviders returns the list of providers from the provider config file
func LoadProviders(path string) (providers Providers, err error) {
	c := &configToml{}
	providers.Providers = make(map[string]Provider)
	_, err = toml.DecodeFile(path, &c)
	if err != nil {
		err = fmt.Errorf("failed to load providers from config file: %s", err)
		return
	}

	for key, value := range c.Providers {
		provider := Provider{
			Name:         key,
			PreCmd:       value.PreCmd,
			PostCmd:      value.PostCmd,
			DetectionCmd: value.DetectionCmd,
			BackupDir:    value.BackupDir,
		}
		providers.Providers[key] = provider
	}
	return
}

// GetProvider returns a provider based on detection commands
func (providers *Providers) GetProvider(o orchestrators.Orchestrator, v *volume.Volume) (prov Provider, err error) {
	detectionCmds := []string{}
	for _, p := range providers.Providers {
		detectionCmds = append(detectionCmds, fmt.Sprintf("(%s && echo '%s')", p.DetectionCmd, p.Name))
	}
	detectionCmds = append(detectionCmds, "true")
	fullDetectionCmd := strings.Join(detectionCmds, " || ")

	containers, err := o.GetContainersMountingVolume(v)
	if err != nil {
		return
	}
	if len(containers) < 1 {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Info("No running container found using the volume.")
		return
	}

	var stdout string
	for _, container := range containers {
		fullDetectionCmd = strings.Replace(fullDetectionCmd, "$volume", container.Path, -1)
		log.WithFields(log.Fields{
			"volume": v.Name,
			"cmd":    fullDetectionCmd,
		}).Debugf("Running detection command in container %s...", container.ContainerID)

		stdout, err = o.ContainerExec(container, []string{"bash", "-c", fullDetectionCmd})
		if err != nil {
			log.Errorf("failed to run provider detection: %s", err)
		}

		stdout = strings.TrimSpace(stdout)

		for _, p := range providers.Providers {
			if p.Name == stdout {
				log.WithFields(log.Fields{
					"volume": v.Name,
				}).Infof("This volume should be a %s datadir", p.Name)
				prov = p
				v.BackupDir = p.BackupDir
				return
			}
		}
	}
	return
}

// RunCmd runs a command into a container
func RunCmd(p Provider, o orchestrators.Orchestrator, v *volume.Volume, cmd string) (err error) {
	containers, err := o.GetContainersMountingVolume(v)
	if err != nil {
		return err
	}

	cmdSuccess := false
	var stdout string
	for _, container := range containers {
		cmd = strings.Replace(cmd, "$volume", container.Path, -1)

		log.WithFields(log.Fields{
			"volume": v.Name,
			"cmd":    cmd,
		}).Debugf("Running command in container %s...", container.ContainerID)

		stdout, err = o.ContainerExec(container, []string{"bash", "-c", cmd})
		if err != nil {
			log.WithFields(log.Fields{
				"volume":    v.Name,
				"cmd":       cmd,
				"container": container.ContainerID,
			}).Errorf("failed to run command in container: %s", err)
		} else {
			cmdSuccess = true
			break
		}
	}

	if cmdSuccess {
		log.WithFields(log.Fields{
			"volume": v.Name,
			"cmd":    cmd,
		}).Debugf("stdout: %s", stdout)
	} else {
		return fmt.Errorf("failed to run command \"%s\" in containers mounting volume %s", cmd, v.Name)
	}
	return
}
