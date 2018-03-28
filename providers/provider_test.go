package providers

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/orchestrators"
	"github.com/camptocamp/conplicity/volume"
)

func TestPrepareBackupWithDocker(t *testing.T) {

	// Use Default provider
	dir, _ := ioutil.TempDir("", "test_backup_volume")
	defer os.RemoveAll(dir)

	c := &handler.Conplicity{}
	c.Config = &config.Config{}
	c.Config.Orchestrator = "docker"
	c.Config.Duplicity.Image = "camptocamp/duplicity:latest"
	c.Config.Docker.Endpoint = "unix:///var/run/docker.sock"

	o := orchestrators.GetOrchestrator(c)

	p := &DefaultProvider{
		BaseProvider: &BaseProvider{
			orchestrator: o,
			vol: &volume.Volume{
				Name:       dir,
				Driver:     "local",
				Mountpoint: "/mnt",
			},
		},
	}

	log.SetLevel(log.DebugLevel)

	err := PrepareBackup(p)

	if err != nil {
		t.Fatalf("Expected no error, got error: %v", err)
	}
}

func TestBaseGetHandler(t *testing.T) {
	expected := ""

	p := &BaseProvider{
		orchestrator: &orchestrators.DockerOrchestrator{
			Handler: &handler.Conplicity{},
		},
	}
	got := p.orchestrator.GetHandler().Hostname
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestBaseGetVolume(t *testing.T) {
	got := (&BaseProvider{}).GetVolume()
	if got != nil {
		t.Fatalf("Expected to get nil, got %v", got)
	}
}

func TestBaseGetBackupDir(t *testing.T) {
	got := (&BaseProvider{}).GetBackupDir()
	if got != "" {
		t.Fatalf("Expected to get nil, got %s", got)
	}
}
