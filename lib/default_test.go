package conplicity

import (
	"testing"

	"github.com/docker/engine-api/types"
)

func TestDefaultGetName(t *testing.T) {
	expected := "Default"
	got := (&DefaultProvider{}).GetName()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
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
