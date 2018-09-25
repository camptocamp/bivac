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
	BackupCmd    string `toml:"backup_cmd"`
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
			BackupCmd:    value.BackupCmd,
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

	container, err := getContainer(o, v)
	if err != nil {
		return
	}
	if container == nil {
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Info("No running container found using the volume.")
		return
	}

	stdout, err := o.ContainerExec(container, []string{"bash", "-c", fullDetectionCmd})
	if err != nil {
		log.Errorf("failed to run provider detection: %s", err)
	}

	switch strings.TrimSpace(stdout) {
	case "mysql":
		log.WithFields(log.Fields{
			"volume": v.Name,
		}).Debug("mysql directory found, this should be MySQL datadir")
		prov = providers.Providers["mysql"]
		v.BackupDir = prov.BackupDir
	}
	return
}

// RunCmd runs a command into a container
func RunCmd(p Provider, o orchestrators.Orchestrator, v *volume.Volume, cmd string) (err error) {
	container, err := getContainer(o, v)
	if err != nil {
		return err
	}

	cmd = strings.Replace(cmd, "$volume", container.Volumes[v.Name], -1)
	stdout, err := o.ContainerExec(container, []string{"bash", "-c", cmd})
	if err != nil {
		return fmt.Errorf("failed to execute command in container: %v", err)
	}
	log.WithFields(log.Fields{
		"volume": v.Name,
		"cmd":    cmd,
	}).Debugf("stdout: %s", stdout)
	return
}

func getContainer(o orchestrators.Orchestrator, v *volume.Volume) (mountedVolumes *volume.MountedVolumes, err error) {
	containers, err := o.GetMountedVolumes(v)
	if err != nil {
		err = fmt.Errorf("failed to list containers: %v", err)
		return
	}

	for _, c := range containers {
		for volName := range c.Volumes {
			if volName == v.Name {
				mountedVolumes = c
				return
			}
		}
	}
	return
}
