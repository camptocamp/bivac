package engines

import (
	"os/exec"
	"syscall"
)

// Engine implements a backup engine interface
type Engine interface {
	Backup() error
	GetName() string
}

func handleExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}
