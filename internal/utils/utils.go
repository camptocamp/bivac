package utils

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

// OutputFormat stores output of Restic commands
type OutputFormat struct {
	Stdout   string `json:"stdout"`
	ExitCode int    `json:"rc"`
}

// MsgFormat is a format used to communicate with the Bivac API
type MsgFormat struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// ReturnFormattedOutput returns a formatted message
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

// ReturnError returns a formatted error
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

// Copy a file's binary contents to another file
func CopyFile(sourcePath string, targetPath string) error {
	sourceFInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !sourceFInfo.Mode().IsRegular() {
		return nil
	}
	targetFInfo, err := os.Stat(targetPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if !targetFInfo.Mode().IsRegular() {
		if targetFInfo.IsDir() {
			err := os.RemoveAll(targetPath)
			if err != nil {
				return err
			}
		} else {
			return nil
		}
	}
	if os.SameFile(sourceFInfo, targetFInfo) {
		return nil
	}
	err = os.Link(sourcePath, targetPath)
	if err != nil {
		err = copyFileContents(sourcePath, targetPath)
		if err != nil {
			return err
		}
	}
	return nil
}

// slower but safer than creating a hardlink when a target file exists
func copyFileContents(sourcePath string, targetPath string) error {
	sourceFInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	sourceData, err := ioutil.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(targetPath, sourceData, sourceFInfo.Mode())
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
