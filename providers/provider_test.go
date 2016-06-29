package providers

import (
	"testing"

	"github.com/camptocamp/conplicity/lib"
)

func TestBaseGetHandler(t *testing.T) {
	expected := ""

	p := &BaseProvider{
		handler: &conplicity.Conplicity{},
	}
	got := p.GetHandler().Hostname
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestBaseGetVolume(t *testing.T) {
	got := (&BaseProvider{}).GetVolume()
	if got != nil {
		t.Fatalf("Expected to get nil, got %s", got)
	}
}

func TestBaseGetBackupDir(t *testing.T) {
	got := (&BaseProvider{}).GetBackupDir()
	if got != "" {
		t.Fatalf("Expected to get nil, got %s", got)
	}
}
