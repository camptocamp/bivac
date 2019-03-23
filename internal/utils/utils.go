package utils

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

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

// Generate a random string
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	stringByte := make([]byte, length)
	for i := range stringByte {
		stringByte[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(stringByte)
}

// Get a random file name unique from the files found in the parentPath
func GetRandomFileName(parentPath string) (string, error) {
	randomFileName := GenerateRandomString(16)
	randomFilePath := strings.ReplaceAll(parentPath+"/"+randomFileName, "//", "/")
	_, err := os.Stat(randomFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return randomFileName, nil
		}
		return "", err
	}
	return GetRandomFileName(parentPath)
}

// Get a random file name unique from the file paths found in the parentPath
func GetRandomFilePath(parentPath string) (string, error) {
	randomFileName, err := GetRandomFileName(parentPath)
	if err != nil {
		return "", err
	}
	randomFilePath := strings.ReplaceAll(parentPath+"/"+randomFileName, "//", "/")
	return randomFilePath, nil
}

// Merge a source path into a target path
func MergePaths(rootSourcePath string, rootTargetDir string) error {
	rootSourceFInfo, err := os.Stat(rootSourcePath)
	if err != nil {
		return err
	}
	if !rootSourceFInfo.IsDir() {
		err = CopyFile(rootSourcePath, rootTargetDir)
		if err != nil {
			return err
		}
		return nil
	}
	rootTargetFInfo, err := os.Stat(rootTargetDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !rootTargetFInfo.IsDir() {
			err = os.Remove(rootTargetDir)
			if err != nil {
				return err
			}
		}
	}
	err = filepath.Walk(
		rootSourcePath,
		func(
			sourcePath string,
			sourceFInfo os.FileInfo,
			err error,
		) error {
			sharedPath := sourcePath[len(rootSourcePath):]
			if err != nil {
				return err
			}
			targetPath := strings.ReplaceAll(rootTargetDir+"/"+sharedPath, "//", "/")
			if sourceFInfo.IsDir() {
				targetFInfo, err := os.Stat(targetPath)
				if err != nil {
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					if !targetFInfo.IsDir() {
						err = os.Remove(targetPath)
						if err != nil {
							return err
						}
					}
				}
				os.MkdirAll(targetPath, sourceFInfo.Mode())
			} else {
				err = CopyFile(sourcePath, targetPath)
				if err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
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
