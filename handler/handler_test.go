package handler

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/volume"

	log "github.com/Sirupsen/logrus"
)

func TestSetup(t *testing.T) {
	var fakeHandler = Conplicity{}

	t.Skip("Fails with t flag")

	fakeHandler.Setup("noversion")

	// Check Hostname
	if fakeHandler.Hostname == "" {
		t.Fatal("Hostname should not be nil")
	}

	// Check default Loglevel
	if l := log.GetLevel(); l != log.InfoLevel {
		t.Fatalf("Expected %v loglevel by default, got %v", log.InfoLevel, l)
	}

	// Check setting Loglevel
	fakeHandler.Config.Loglevel = "debug"
	fakeHandler.setupLoglevel()
	if l := log.GetLevel(); l != log.DebugLevel {
		t.Fatalf("Expected %v loglevel, got %v", log.DebugLevel, l)
	}

	// Check setting Loglevel to wrong value
	fakeHandler.Config.Loglevel = "wrong"
	err := fakeHandler.setupLoglevel()
	if err == nil {
		t.Fatal("Expected setupLoglevel to fail")
	}
}

func TestSchedulerVolumeNoVerify(t *testing.T) {
	fakeMountpoint, err := ioutil.TempDir("", "testConplicity")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %v", err)
	}

	defer os.RemoveAll(fakeMountpoint)

	vol := volume.Volume{
		Mountpoint: fakeMountpoint,
		Config: &volume.Config{
			NoVerify: true,
		},
	}

	c := Conplicity{}

	result, err := c.IsCheckScheduled(&vol)

	if result != false {
		t.Fatal("Expected vol.Config.NoVerify equals to false")
	}
}

func TestSchedulerVolumePermissionDenied(t *testing.T) {
	fakeMountpoint, err := ioutil.TempDir("", "testConplicity")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %v", err)
	}

	defer os.RemoveAll(fakeMountpoint)

	os.OpenFile(fakeMountpoint+"/.conplicity_last_check", os.O_RDONLY|os.O_CREATE, 0644)
	os.Chmod(fakeMountpoint, 0644)

	vol := volume.Volume{
		Mountpoint: fakeMountpoint,
		Config: &volume.Config{
			NoVerify: false,
		},
	}

	c := Conplicity{
		Config: &config.Config{
			CheckEvery: "1h",
		},
	}

	result, err := c.IsCheckScheduled(&vol)

	if result != false {
		t.Fatal("Expected false, got true.")
	}
}

func TestSchedulerVolumeInvalidCheckEvery(t *testing.T) {
	fakeMountpoint, err := ioutil.TempDir("", "testConplicity")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %v", err)
	}

	defer os.RemoveAll(fakeMountpoint)

	os.OpenFile(fakeMountpoint+"/.conplicity_last_check", os.O_RDONLY|os.O_CREATE, 0644)

	vol := volume.Volume{
		Mountpoint: fakeMountpoint,
		Config: &volume.Config{
			NoVerify: false,
		},
	}

	c := Conplicity{
		Config: &config.Config{
			CheckEvery: "fake",
		},
	}

	result, err := c.IsCheckScheduled(&vol)

	if result != false {
		t.Fatal("Expected false, got true.")
	}
}

func TestSchedulerVolumeVerifyNotRequired(t *testing.T) {
	fakeMountpoint, err := ioutil.TempDir("", "testConplicity")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %v", err)
	}

	defer os.RemoveAll(fakeMountpoint)

	vol := volume.Volume{
		Mountpoint: fakeMountpoint,
		Config: &volume.Config{
			NoVerify: false,
		},
	}

	c := Conplicity{
		Config: &config.Config{
			CheckEvery: "1h",
		},
	}

	result, err := c.IsCheckScheduled(&vol)

	if result != false {
		t.Fatal("Expected false, got true.")
	}
}

func TestSchedulerVolumeVerifyRequired(t *testing.T) {
	fakeMountpoint, err := ioutil.TempDir("", "testConplicity")
	if err != nil {
		t.Fatalf("Cannot create temporary directory: %v", err)
	}

	defer os.RemoveAll(fakeMountpoint)

	os.OpenFile(fakeMountpoint+"/.conplicity_last_check", os.O_RDONLY|os.O_CREATE, 0644)
	h := time.Now().Local().AddDate(0, 0, -1)
	os.Chtimes(fakeMountpoint+"/.conplicity_last_check", h, h)

	vol := volume.Volume{
		Mountpoint: fakeMountpoint,
		Config: &volume.Config{
			NoVerify: false,
		},
	}

	c := Conplicity{
		Config: &config.Config{
			CheckEvery: "1h",
		},
	}

	result, err := c.IsCheckScheduled(&vol)

	if result != true {
		t.Fatal("Expected true, got false.")
	}
}
