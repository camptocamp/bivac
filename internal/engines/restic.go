package engines

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/camptocamp/bivac/internal/utils"
)

// ResticEngine implements a backup engine with Restic
type ResticEngine struct {
	DefaultArgs []string
	Output      map[string]utils.OutputFormat
}

// Snapshot is a struct returned by the function snapshots()
type Snapshot struct {
	Time time.Time `json:"time"`
}

// GetName returns the engine name
func (*ResticEngine) GetName() string {
	return "Restic"
}

// Backup performs the backup of the passed volume
func (r *ResticEngine) Backup(backupPath, hostname string, force bool) string {
	var err error

	err = r.initializeRepository()
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	if force {
		err = r.unlockRepository()
		if err != nil {
			return utils.ReturnFormattedOutput(r.Output)
		}
	}

	err = r.backupVolume(hostname, backupPath)
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	err = r.forget()
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	return utils.ReturnFormattedOutput(r.Output)
}

func (r *ResticEngine) initializeRepository() (err error) {
	rc := 0

	// Check if the remote repository exists
	output, err := exec.Command("restic", append(r.DefaultArgs, "snapshots")...).CombinedOutput()
	if err != nil {
		rc = handleExitCode(err)
	}
	if rc == 0 {
		return
	}
	r.Output["testInit"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil

	rc = 0
	// Create remote repository
	output, err = exec.Command("restic", append(r.DefaultArgs, "init")...).CombinedOutput()
	if err != nil {
		rc = handleExitCode(err)
	}
	r.Output["init"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil
	return
}

func (r *ResticEngine) backupVolume(hostname, backupPath string) (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, []string{"--host", hostname, "backup", backupPath}...)...).CombinedOutput()
	if err != nil {
		rc = handleExitCode(err)
	}
	r.Output["backup"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil
	return
}

func (r *ResticEngine) forget() (err error) {
	rc := 0
	cmd := append(r.DefaultArgs, "forget")
	cmd = append(cmd, strings.Split(os.Getenv("RESTIC_FORGET_ARGS"), " ")...)

	output, err := exec.Command("restic", cmd...).CombinedOutput()
	if err != nil {
		rc = handleExitCode(err)
	}
	r.Output["forget"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil
	return
}

func (r *ResticEngine) unlockRepository() (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, []string{"unlock", "--remove-all"}...)...).CombinedOutput()
	if err != nil {
		rc = handleExitCode(err)
	}
	r.Output["unlock"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil
	return
}

func (r *ResticEngine) GetLastBackupDate() (t time.Time, err error) {
	output, _ := exec.Command("restic", append(r.DefaultArgs, []string{"snapshots", "--last"}...)...).CombinedOutput()

	var data []map[string]interface{}
	err = json.Unmarshal(output, &data)
	if err != nil {
		return
	}
	snapshot := data[len(data)-1]

	t, err = time.Parse("2006-01-02T15:04:05Z", snapshot["time"].(string))
	if err != nil {
		return
	}
	return
}
