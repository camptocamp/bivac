package agent

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func Start() (exitCode int) {
	exitCode = 0
	output, err := resticBackup()
	if err != nil {
		exitCode = handleExitCode(err)
		return
	}
	fmt.Printf("%s", output)
	return
}

func resticBackup() (output []byte, err error) {
	exitCode := 0
	output, err = exec.Command("restic", "--no-cache", "--json", "--host", os.Getenv("RESTIC_HOSTNAME"), "-r", os.Getenv("RESTIC_REPOSITORY"), "backup", os.Getenv("RESTIC_BACKUP_PATH")).CombinedOutput()
	if err != nil {
		exitCode = handleExitCode(err)
	}
	return
}

func handleExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}

//func resticCheckLock() (output []byte, err error) {
//	output, err = exec.Command("restic", "--no-cache", "--json", "--host", os.Getenv("RESTIC_HOSTNAME"), "-r", os.Getenv("RESTIC_REPOSITORY"), "locks").CombinedOutput()
//	return
//}
