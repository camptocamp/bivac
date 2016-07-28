package providers

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/handler"
	"github.com/camptocamp/conplicity/volume"
	"github.com/docker/engine-api/types"
)

func TestGetProvider(t *testing.T) {
	var dir, expected, got string
	var p Provider

	// Test PostgreSQL detection
	expected = "PostgreSQL"
	dir, _ = ioutil.TempDir("", "test_get_provider_postgresql")
	defer os.RemoveAll(dir)
	ioutil.WriteFile(dir+"/PG_VERSION", []byte{}, 0644)

	p = GetProvider(&handler.Conplicity{}, &volume.Volume{
		Volume: &types.Volume{
			Mountpoint: dir,
		}})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test MySQL detection
	expected = "MySQL"
	dir, _ = ioutil.TempDir("", "test_get_provider_mysql")
	defer os.RemoveAll(dir)
	os.Mkdir(dir+"/mysql", 0755)

	p = GetProvider(&handler.Conplicity{}, &volume.Volume{
		Volume: &types.Volume{
			Mountpoint: dir,
		}})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test OpenLDAP detection
	expected = "OpenLDAP"
	dir, _ = ioutil.TempDir("", "test_get_provider_openldap")
	defer os.RemoveAll(dir)
	ioutil.WriteFile(dir+"/DB_CONFIG", []byte{}, 0644)

	p = GetProvider(&handler.Conplicity{}, &volume.Volume{
		Volume: &types.Volume{
			Mountpoint: dir,
		}})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test Default detection
	expected = "Default"
	dir, _ = ioutil.TempDir("", "test_get_provider_default")
	defer os.RemoveAll(dir)

	p = GetProvider(&handler.Conplicity{}, &volume.Volume{
		Volume: &types.Volume{
			Mountpoint: dir,
		}})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}
}

func TestPrepareBackup(t *testing.T) {
	// Use Default provider
	dir, _ := ioutil.TempDir("", "test_backup_volume")
	defer os.RemoveAll(dir)

	p := &DefaultProvider{
		BaseProvider: &BaseProvider{
			handler: &handler.Conplicity{},
			vol: &volume.Volume{
				Volume: &types.Volume{
					Name:       dir,
					Driver:     "local",
					Mountpoint: "/mnt",
				},
			},
		},
	}

	// Fill Config manually
	p.handler.Config = &config.Config{} // Init Config
	p.handler.Config.Duplicity.Image = "camptocamp/duplicity:latest"
	p.handler.Config.Docker.Endpoint = "unix:///var/run/docker.sock"
	p.handler.Hostname, _ = os.Hostname()
	p.handler.SetupDocker()

	log.SetLevel(log.DebugLevel)

	err := PrepareBackup(p)

	if err != nil {
		t.Fatalf("Expected no error, got error: %v", err)
	}
}

func TestBaseGetHandler(t *testing.T) {
	expected := ""

	p := &BaseProvider{
		handler: &handler.Conplicity{},
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
