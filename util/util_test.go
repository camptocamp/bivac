package util

import (
	"errors"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var fakeVol = docker.Volume{
	Name: "Test",
	Labels: map[string]string{
		"io.conplicity.test": "Fake label",
	},
}

func ExampleCheckErrNil() {
	CheckErr(nil, "test", 0)
	// Output:
}

func ExampleCheckErrNoExit() {
	fakeErr := errors.New("Fake error %v")
	CheckErr(fakeErr, "test", 0)
	// Output:
}

func ExampleCheckErrExit() {
	fakeErr := errors.New("Fake error %v")
	CheckErr(fakeErr, "test", -1)
	// Output:
	// How do we test the exit?
}

func TestVolumeLabel(t *testing.T) {
	expectedStr := "Fake label"
	result, _ := GetVolumeLabel(&fakeVol, ".test")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}

func TestVolumeLabelNotFound(t *testing.T) {
	expectedStr := ""
	expectedErr := "Key .unknown not found in labels for volume Test"
	result, err := GetVolumeLabel(&fakeVol, ".unknown")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}

	if err.Error() != expectedErr {
		t.Fatalf("Expected %v, got %v", expectedErr, err)
	}
}
