package engines

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/orchestrators"
	"github.com/camptocamp/bivac/volume"
)

// RCloneEngine implements a backup engine with RClone
type RCloneEngine struct {
	Orchestrator orchestrators.Orchestrator
	Volume       *volume.Volume
}

// GetName returns the engine name
func (*RCloneEngine) GetName() string {
	return "RClone"
}

// replaceArgs replace arguments with their values
func (r *RCloneEngine) replaceArgs(args []string) (newArgs []string) {
	log.Debugf("Replacing args, Input: %v", args)
	for _, arg := range args {
		arg = strings.Replace(arg, "%B", r.Volume.Config.TargetURL, -1)
		arg = strings.Replace(arg, "%D", r.Volume.BackupDir, -1)
		arg = strings.Replace(arg, "%H", r.Volume.Hostname, -1)
		arg = strings.Replace(arg, "%N", r.Volume.Namespace, -1)
		arg = strings.Replace(arg, "%P", r.Orchestrator.GetPath(r.Volume), -1)
		arg = strings.Replace(arg, "%V", r.Volume.Name, -1)
		newArgs = append(newArgs, arg)
	}
	log.Debugf("Replacing args, Output: %v", newArgs)
	return
}

// Backup performs the backup of the passed volume
func (r *RCloneEngine) Backup() (err error) {
	config := r.Orchestrator.GetHandler().Config
	v := r.Volume

	v.BackupDir = v.Mountpoint + "/" + v.BackupDir

	state, _, err := r.launchRClone(
		append([]string{"sync"}, strings.Split(config.RClone.BackupArgs, " ")...),
		[]*volume.Volume{v},
	)
	if err != nil {
		err = fmt.Errorf("failed to launch RClone: %v", err)
	}
	if state != 0 {
		err = fmt.Errorf("RClone exited with state %v", state)
	}
	return
}

// launchRClone starts an rclone container with a given command
func (r *RCloneEngine) launchRClone(cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	config := r.Orchestrator.GetHandler().Config
	image := config.RClone.Image

	return r.Orchestrator.LaunchContainer(image, r.replaceArgs(append(cmd, strings.Split(config.RClone.CommonArgs, " ")...)), volumes)
}
