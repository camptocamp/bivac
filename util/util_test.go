package util

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
)

var fakeDockerVol = types.Volume{
	Name: "Test",
	Labels: map[string]string{
		"io.conplicity.test": "Fake label",
	},
}

func ExampleCheckErr_nil() {
	CheckErr(nil, "test", "fatal")
	// Output:
}

func ExampleCheckErr_noExit() {
	fakeErr := errors.New("Fake error")
	CheckErr(fakeErr, "test: %v", "error")
	// Output:
}

func ExampleCheckErr_exit() {
	// fakeErr := errors.New("Fake error")
	// How do we test the output and the os.Exit(1)?"
	// CheckErr(fakeErr, "test: %v", "fatal")
}

func TestVolumeLabel(t *testing.T) {
	expectedStr := "Fake label"
	result, _ := GetVolumeLabel(&fakeDockerVol, "test")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}

func TestVolumeLabelNotFound(t *testing.T) {
	expectedStr := ""
	expectedErr := "Key .unknown not found in labels for volume Test"
	result, err := GetVolumeLabel(&fakeDockerVol, ".unknown")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}

	if err.Error() != expectedErr {
		t.Fatalf("Expected %v, got %v", expectedErr, err)
	}
}
