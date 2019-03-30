package engine

import (
	"encoding/json"
	"io/ioutil"
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

	err = r.forget()
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	err = r.retrieveBackupsStats()
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}

	return utils.ReturnFormattedOutput(r.Output)
}

// Restore performs the restore of the passed volume
func (r *Engine) Restore(
	backupPath,
	hostname string,
	force bool,
	snapshotName string,
) string {
	var err error
	if force {
		err = r.unlockRepository()
		if err != nil {
			return utils.ReturnFormattedOutput(r.Output)
		}
	}
	err = r.restoreVolume(hostname, backupPath, snapshotName)
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
	}
	err = r.retrieveBackupsStats()
	if err != nil {
		return utils.ReturnFormattedOutput(r.Output)
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
	err = nil
	return
}

func (r *Engine) restoreVolume(
	hostname,
	backupPath string,
	snapshotName string,
) (err error) {
	rc := 0
	origionalBackupPath := r.getOrigionalBackupPath(
		hostname,
		backupPath,
		snapshotName,
	)
	workingPath, err := utils.GetRandomFilePath(backupPath)
	workingPath = strings.ReplaceAll(workingPath, "//", "/")
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	err = os.MkdirAll(workingPath, 0700)
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	output, err := exec.Command(
		"restic",
		append(
			r.DefaultArgs,
			[]string{
				"restore",
				snapshotName,
				"--target",
				workingPath,
			}...,
		)...,
	).CombinedOutput()
	restoreDumpPath := workingPath + origionalBackupPath
	files, err := ioutil.ReadDir(restoreDumpPath)
	if err != nil {
		rc = utils.HandleExitCode(err)
	}
	collisionName := ""
	for _, f := range files {
		fileName := f.Name()
		restoreSubPath := strings.ReplaceAll(backupPath+"/"+fileName, "//", "/")
		if restoreSubPath == workingPath {
			collisionName, err = utils.GetRandomFileName(workingPath)
			if err != nil {
				rc = utils.HandleExitCode(err)
			}
			restoreSubPath = strings.ReplaceAll(workingPath+"/"+collisionName, "//", "/")
		}
		err = utils.MergePaths(
			strings.ReplaceAll(restoreDumpPath+"/"+fileName, "//", "/"),
			restoreSubPath,
		)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
		err = os.RemoveAll(
			strings.ReplaceAll(restoreDumpPath+"/"+fileName, "//", "/"),
		)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
	}
	if len(collisionName) > 0 {
		tmpWorkingPath, err := utils.GetRandomFilePath(backupPath)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
		err = os.Rename(
			workingPath,
			tmpWorkingPath,
		)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
		err = os.Rename(
			strings.ReplaceAll(tmpWorkingPath+"/"+collisionName, "//", "/"),
			workingPath,
		)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
		err = os.RemoveAll(tmpWorkingPath)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
	} else {
		err = os.RemoveAll(workingPath)
		if err != nil {
			rc = utils.HandleExitCode(err)
		}
	}
	r.Output["restore"] = utils.OutputFormat{
		Stdout:   string(output),
		ExitCode: rc,
	}
	err = nil
	return
}

func (r *Engine) getOrigionalBackupPath(
	hostname,
	backupPath string,
	snapshotName string,
) string {
	output, err := exec.Command(
		"restic",
		append(
			r.DefaultArgs,
			[]string{"ls", snapshotName}...,
		)...,
	).CombinedOutput()
	if err != nil {
		return err.Error()
	}
	type Header struct {
		Paths []string `json:"paths"`
	}
	headerJSON := []byte("{\"paths\": [\"\"]")
	jsons := strings.Split(string(output), "\n")
	for i := 0; i < len(jsons); i++ {
		if strings.Index(jsons[i], "\",\"paths\":[\"") > -1 {
			headerJSON = []byte(jsons[i])
			break
		}
	}
	var header Header
	err = json.Unmarshal(headerJSON, &header)
	if err != nil {
		return ""
	}
	return header.Paths[0]
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
