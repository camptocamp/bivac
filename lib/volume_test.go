package conplicity

import (
	"fmt"
	"strings"
	"testing"
)

// Set up mocked Conplicity handler
type fakeHandler struct{}

var fakeStdout string

func (h *fakeHandler) LaunchDuplicity(c, m []string) (state int, stdout string, err error) {
	fmt.Printf("Command: %s\n", strings.Join(c, " "))
	fmt.Printf("Mounts: %s\n", strings.Join(m, " "))

	stdout = fakeStdout
	state = 42

	return state, stdout, nil
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
	m, _ := fakeVol.Backup()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: --full-if-older-than 3W --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /back /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="backupExitCode"} 42
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
	m, _ := fakeVol.Verify()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: verify --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /foo /back
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="verifyExitCode"} 42
}

// ExampleStatus checks the status of a backup
func ExampleStatus() {
	fakeStdout = "Last full backup date: Mon Jan 2 15:04:05 2006  \nChain end time: Mon Jan 2 15:04:05 2006  \n"
	m, _ := fakeVol.Status()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: collection-status --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --name Test /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="lastBackup"} 1136214245
	// conplicity{volume="Test",what="lastFullBackup"} 1136214245
}

// ExampleStatusNoFullBackup checks the status of a backup
func ExampleStatusNoFullBackup() {
	fakeStdout = "Last full backup date: none  \nNo backup chains with active signatures found  \n"
	m, _ := fakeVol.Status()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: collection-status --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --name Test /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="lastBackup"} 0
	// conplicity{volume="Test",what="lastFullBackup"} 0
}

// ExampleStatusMetadataSync checks the status of a backup
func ExampleStatusMetadataSync() {
	fakeStdout = "Local and Remote metadata are synchronized, no sync needed.\r\n"
	m, _ := fakeVol.Status()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: collection-status --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --name Test /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="lastBackup"} 0
	// conplicity{volume="Test",what="lastFullBackup"} 0
}

// TestStatusNoFullBackupDate checks the status of a backup
func TestStatusNoFullBackupDate(t *testing.T) {
	fakeStdout = "Wrong stdout"
	_, err := fakeVol.Status()

	if err == nil {
		t.Fatal("Expected an error, got no error")
	}

	expected := "Failed to parse Duplicity output for last full backup date of Test"
	got := err.Error()
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

// TestStatusNoChainEndTime checks the status of a backup
func TestStatusNoChainEndTime(t *testing.T) {
	fakeStdout = "Last full backup date: Mon Jan 2 15:04:05 2006  \nWhatever else  \n"
	_, err := fakeVol.Status()

	if err == nil {
		t.Fatal("Expected an error, got no error")
	}

	expected := "Failed to parse Duplicity output for chain end time of Test"
	got := err.Error()
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}
