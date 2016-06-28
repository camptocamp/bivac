package util

import (
	"fmt"
	"testing"

	"github.com/camptocamp/conplicity/util"
	"github.com/fsouza/go-dockerclient"
)

func TestVolumeLabel(t *testing.T) {
	vol := docker.Volume{
		Labels: map[string]string{
			"io.conplicity.test": "Fake label",
		},
	}
	expectedStr := "Fake label"
	result := util.GetVolumeLabel(&vol, ".test")
	fmt.Println(result)
	if result != expectedStr {
		t.Fatalf("Expected %s, got %s", expectedStr, result)
	}
}
