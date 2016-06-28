package volume

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

// Set up mocked Conplicity handler
type fakeHandler struct {
	Command []string
	Mounts  []string
}

func (h *fakeHandler) LaunchDuplicity(c, m []string) (docker.State, string, error) {
	fmt.Printf("Command: %s\n", strings.Join(c, " "))
	fmt.Printf("Mounts: %s\n", strings.Join(m, " "))

	return docker.State{}, "", nil
}

// Set up fake volume
var fakeVol = Volume{
	Name:            "Test",
	Target:          "/foo",
	BackupDir:       "/back",
	Mount:           "/mnt",
	FullIfOlderThan: "3W",
	RemoveOlderThan: "1Y",
	Client:          &fakeHandler{},
}

// TestNewVolume checks the creation of a new volume
func TestNewVolume(t *testing.T) {
	if fakeVol.Name != "Test" {
		t.Fatalf("Volume name is wrong. Expected Test, got %v", fakeVol.Name)
	}

	if fakeVol.Target != "/foo" {
		t.Fatalf("Volume target is wrong. Expected /foo, got %v", fakeVol.Target)
	}

	if fakeVol.BackupDir != "/back" {
		t.Fatalf("Volume backup dir is wrong. Expected /back, got %v", fakeVol.BackupDir)
	}

	if fakeVol.Mount != "/mnt" {
		t.Fatalf("Volume mount dir is wrong. Expected /mnt, got %v", fakeVol.Mount)
	}

	if fakeVol.FullIfOlderThan != "3W" {
		t.Fatalf("Volume FullIfOlderThan is wrong. Expected 3W, got %v", fakeVol.FullIfOlderThan)
	}

	if fakeVol.RemoveOlderThan != "1Y" {
		t.Fatalf("Volume RemoveOlderThan is wrong. Expected 1Y, got %v", fakeVol.RemoveOlderThan)
	}
}

// ExampleBackup checks the launching of a volume backup
func ExampleBackup() {
	fakeVol.Backup()
	// Output:
	// Command: --full-if-older-than 3W --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /back /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
}

// ExampleRemoveOld checks the removal of old backups
func ExampleRemoveOld() {
	fakeVol.RemoveOld()
	// Output:
	// Command: remove-older-than 1Y --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --force --name Test /foo
	// Mounts: duplicity_cache:/root/.cache/duplicity
}

// ExampleCleanup checks the cleanup of a backup
func ExampleCleanup() {
	fakeVol.Cleanup()
	// Output:
	// Command: cleanup --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --force --extra-clean --name Test /foo
	// Mounts: duplicity_cache:/root/.cache/duplicity
}

// ExampleVerify checks the verification of a backup
func ExampleVerify() {
	fakeVol.Verify()
	// Output:
	// Command: verify --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /foo /back
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
}

// ExampleStatus checks the status of a backup
func ExampleStatus() {
	fakeVol.Status()
	// Output:
	// Command: collection-status --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --name Test /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
}
