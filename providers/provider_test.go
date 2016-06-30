package providers

import (
	"io/ioutil"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/lib"
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

	p = GetProvider(&conplicity.Conplicity{}, &types.Volume{
		Mountpoint: dir,
	})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test MySQL detection
	expected = "MySQL"
	dir, _ = ioutil.TempDir("", "test_get_provider_mysql")
	defer os.RemoveAll(dir)
	os.Mkdir(dir+"/mysql", 0755)

	p = GetProvider(&conplicity.Conplicity{}, &types.Volume{
		Mountpoint: dir,
	})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test OpenLDAP detection
	expected = "OpenLDAP"
	dir, _ = ioutil.TempDir("", "test_get_provider_openldap")
	defer os.RemoveAll(dir)
	ioutil.WriteFile(dir+"/DB_CONFIG", []byte{}, 0644)

	p = GetProvider(&conplicity.Conplicity{}, &types.Volume{
		Mountpoint: dir,
	})
	got = p.GetName()
	if got != expected {
		t.Fatalf("Expected provider %s, got %s", expected, got)
	}

	// Test Default detection
	expected = "Default"
	dir, _ = ioutil.TempDir("", "test_get_provider_default")
	defer os.RemoveAll(dir)

	p = GetProvider(&conplicity.Conplicity{}, &types.Volume{
		Mountpoint: dir,
	})
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
			handler: &conplicity.Conplicity{},
			vol: &types.Volume{
				Name:       dir,
				Driver:     "local",
				Mountpoint: "/mnt",
			},
		},
	}

	// Fill Config manually
	p.handler.Config = &conplicity.Config{} // Init Config
	p.handler.Config.Image = "camptocamp/duplicity:latest"
	p.handler.Config.Docker.Endpoint = "unix:///var/run/docker.sock"
	p.handler.Hostname, _ = os.Hostname()
	p.handler.SetupDocker()

	log.SetLevel(log.DebugLevel)

	err := PrepareBackup(p)

	if err != nil {
		t.Fatalf("Expected no error, got error: %v", err)
	}
}

func TestBackupVolume(t *testing.T) {
	// Use Base provider
	dir, _ := ioutil.TempDir("", "test_backup_volume")
	defer os.RemoveAll(dir)

	p := &BaseProvider{
		handler: &conplicity.Conplicity{},
	}

	// Fill Config manually
	p.handler.Config = &conplicity.Config{} // Init Config
	p.handler.Config.Duplicity.FullIfOlderThan = "314D"
	p.handler.Config.Duplicity.RemoveOlderThan = "1Y"
	p.handler.Config.Duplicity.TargetURL = "file:///tmp/backup"
	p.handler.Config.Image = "camptocamp/duplicity:latest"
	p.handler.Config.Docker.Endpoint = "unix:///var/run/docker.sock"
	p.handler.Hostname, _ = os.Hostname()
	p.handler.SetupDocker()

	log.SetLevel(log.DebugLevel)

	err := p.BackupVolume(&types.Volume{
		Name:       dir,
		Driver:     "local",
		Mountpoint: "/mnt",
	})

	if err != nil {
		t.Fatalf("Expected no error, got error: %v", err)
	}
}

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
