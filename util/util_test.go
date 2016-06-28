package util

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var vol = docker.Volume{
	Name: "Test",
	Labels: map[string]string{
		"io.conplicity.test": "Fake label",
	},
}

func TestVolumeLabel(t *testing.T) {
	expectedStr := "Fake label"
	result, _ := GetVolumeLabel(&vol, ".test")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}

func TestVolumeLabelNotFound(t *testing.T) {
	expectedStr := ""
	expectedErr := "Key .unknown not found in labels for volume Test"
	result, err := GetVolumeLabel(&vol, ".unknown")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}

	if err.Error() != expectedErr {
		t.Fatalf("Expected %v, got %v", expectedErr, err)
	}
}
