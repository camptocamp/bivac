package utils

import (
	"encoding/json"
	"os/exec"
	"syscall"
)

type OutputFormat struct {
	Stdout   string `json:"stdout"`
	ExitCode int    `json:"rc"`
}

type MsgFormat struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

func ReturnFormattedOutput(output interface{}) string {
	m := MsgFormat{
		Type:    "success",
		Content: output,
	}
	b, err := json.Marshal(m)
	if err != nil {
		return ReturnError(err)
	}
	return string(b)
}

func ReturnError(e error) string {
	msg := MsgFormat{
		Type:    "error",
		Content: e.Error(),
	}
	data, _ := json.Marshal(msg)
	return string(data)
}

// HandleExitCode retrieve a command exit code from an error
func HandleExitCode(err error) int {
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 0
}
