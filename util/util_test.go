package util

import (
	"errors"
	"testing"

	"github.com/docker/engine-api/types"
)

var fakeVol = types.Volume{
	Name: "Test",
	Labels: map[string]string{
		"io.conplicity.test": "Fake label",
	},
}

func ExampleCheckErrNil() {
	CheckErr(nil, "test", "fatal")
	// Output:
}

func ExampleCheckErrNoExit() {
	fakeErr := errors.New("Fake error")
	CheckErr(fakeErr, "test: %v", "error")
	// Output:
}

func ExampleCheckErrExit() {
	// fakeErr := errors.New("Fake error")
	// How do we test the output and the os.Exit(1)?"
	// CheckErr(fakeErr, "test: %v", "fatal")
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
