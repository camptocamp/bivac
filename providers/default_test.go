package providers

import (
	"testing"

	"github.com/docker/docker/api/types"
)

func TestDefaultGetName(t *testing.T) {
	expected := "Default"
	got := (&DefaultProvider{}).GetName()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestDefaultPrepareBackup(t *testing.T) {
	got := (&DefaultProvider{}).PrepareBackup()
	if got != nil {
		t.Fatalf("Expected to get nil, got %s", got)
	}
}

func TestDefaultGetPrepareCommand(t *testing.T) {
	mount := &types.MountPoint{
		Destination: "/mnt",
	}

	got := (&DefaultProvider{}).GetPrepareCommand(mount)
	if got != nil {
		t.Fatalf("Expected to get nil, got %s", got)
	}
}
