package volume

import "testing"

// Set up fake volume
var fakeVol = Volume{
	Target:    "/foo",
	BackupDir: "/back",
	Mount:     "/mnt",
	Config:    &Config{},
}

// TestNewVolume checks the creation of a new volume
func TestNewVolume(t *testing.T) {
	fakeVol.Config.Duplicity.FullIfOlderThan = "3W"
	fakeVol.Config.Duplicity.RemoveOlderThan = "1Y"

	if fakeVol.Target != "/foo" {
		t.Fatalf("Volume target is wrong. Expected /foo, got %v", fakeVol.Target)
	}

	if fakeVol.BackupDir != "/back" {
		t.Fatalf("Volume backup dir is wrong. Expected /back, got %v", fakeVol.BackupDir)
	}

	if fakeVol.Mount != "/mnt" {
		t.Fatalf("Volume mount dir is wrong. Expected /mnt, got %v", fakeVol.Mount)
	}

	if fakeVol.Config.Duplicity.FullIfOlderThan != "3W" {
		t.Fatalf("Volume FullIfOlderThan is wrong. Expected 3W, got %v", fakeVol.Config.Duplicity.FullIfOlderThan)
	}

	if fakeVol.Config.Duplicity.RemoveOlderThan != "1Y" {
		t.Fatalf("Volume RemoveOlderThan is wrong. Expected 1Y, got %v", fakeVol.Config.Duplicity.RemoveOlderThan)
	}
}
