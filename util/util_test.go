package util

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var vol = docker.Volume{
	Labels: map[string]string{
		"io.conplicity.test": "Fake label",
	},
}

func TestVolumeLabel(t *testing.T) {
	expectedStr := "Fake label"
	result := GetVolumeLabel(&vol, ".test")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}

func TestVolumeLabelNotFound(t *testing.T) {
	expectedStr := ""
	result := GetVolumeLabel(&vol, ".unknown")
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}
