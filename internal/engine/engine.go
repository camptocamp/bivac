package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/camptocamp/bivac/internal/utils"
)

// Engine stores informations to use Restic backup engine
type Engine struct {
	DefaultArgs []string
	Output      map[string]utils.OutputFormat
}

// Snapshot is a struct returned by the function snapshots()
type Snapshot struct {
	Time     time.Time `json:"time"`
	Parent   string    `json:"parent"`
	Tree     string    `json:"tree"`
	Path     []string  `json:"path"`
	Hostname string    `json:"hostname"`
	ID       string    `json:"id"`
	ShortID  string    `json:"short_id"`
}

// GetName returns the engine name
func (*Engine) GetName() string {
	return "restic"
}

// Backup performs the backup of the passed volume
func (r *Engine) Backup(backupPath, hostname string, force bool) string {
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

	// A backup lock may remains. A retry loop with sleeps is probably the best solution to avoid lock errors.
	for i := 0; i < 3; i++ {
		err = r.forget()
		if err == nil {
			break
		}
		time.Sleep(60 * time.Second)
	}
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	for i := 0; i < 3; i++ {
		err = r.retrieveBackupsStats()
		if err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}

	return utils.ReturnFormattedOutput(r.Output)
}

func (r *Engine) initializeRepository() (err error) {
	rc := 0

	// Check if the remote repository exists
	output, err := exec.Command("restic", append(r.DefaultArgs, "snapshots")...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
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
		rc = utils.HandleExitCode(err)
	}
	r.Output["init"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	fmt.Printf("init: %s\n", output)
	err = nil
	return
}

func (r *Engine) backupVolume(hostname, backupPath string) (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, []string{"--host", hostname, "backup", backupPath}...)...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	r.Output["backup"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	fmt.Printf("backup: %s\n", output)
	err = nil
	return
}

func (r *Engine) forget() (err error) {
	rc := 0
	cmd := append(r.DefaultArgs, "forget")
	cmd = append(cmd, strings.Split(os.Getenv("RESTIC_FORGET_ARGS"), " ")...)

	output, err := exec.Command("restic", cmd...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	r.Output["forget"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	fmt.Printf("forget: %s\n", output)
	err = nil
	return
}

func (r *Engine) retrieveBackupsStats() (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, []string{"snapshots"}...)...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	r.Output["snapshots"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	fmt.Printf("snapshots: %s\n", output)

	return
}

func (r *Engine) unlockRepository() (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, []string{"unlock", "--remove-all"}...)...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	r.Output["unlock"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	fmt.Printf("unlock: %s\n", output)
	err = nil
	return
}

// GetBackupDates runs a Restic command locally to retrieve latest snapshot date
func (r *Engine) GetBackupDates() (latestSnapshotDate, oldestSnapshotDate time.Time, err error) {
	output, _ := exec.Command("restic", append(r.DefaultArgs, []string{"snapshots"}...)...).CombinedOutput()

	var data []Snapshot
	err = json.Unmarshal(output, &data)
	if err != nil {
		return
	}

	if len(data) == 0 {
		return
	}

	latestSnapshot := data[len(data)-1]

	latestSnapshotDate = latestSnapshot.Time
	if err != nil {
		return
	}

	oldestSnapshot := data[0]

	oldestSnapshotDate = oldestSnapshot.Time
	if err != nil {
		return
	}
	return
}

// RawCommand runs a custom Restic command locally
func (r *Engine) RawCommand(cmd []string) (err error) {
	rc := 0
	output, err := exec.Command("restic", append(r.DefaultArgs, cmd...)...).CombinedOutput()
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	r.Output["raw"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	return
}
