package engines

import (
	"fmt"
	"strings"
	"testing"

	"github.com/camptocamp/conplicity/config"
	"github.com/camptocamp/conplicity/lib"
	"github.com/camptocamp/conplicity/volume"
	docker "github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
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
var fakeVol = &volume.Volume{
	Volume: &types.Volume{
		Name: "Foo",
	},
	Target:          "/foo",
	BackupDir:       "/back",
	Mount:           "/mnt",
	FullIfOlderThan: "3W",
	RemoveOlderThan: "1Y",
}

var client, _ = docker.NewClient("unix:///var/run/docker.sock", "", nil, nil)
var fakeEngine = &DuplicityEngine{
	Handler: &conplicity.Conplicity{
		Config: &config.Config{},
		Client: client,
	},
}

// ExampleBackup checks the launching of a volume backup
func ExampleBackup() {
	m, _ := fakeEngine.Backup()
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: --full-if-older-than 3W --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /back /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="backupExitCode"} 42
}

// ExampleRemoveOld checks the removal of old backups
func ExampleRemoveOld() {
	fakeEngine.removeOld(fakeVol)
	// Output:
	// Command: remove-older-than 1Y --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --force --name Test /foo
	// Mounts: duplicity_cache:/root/.cache/duplicity
}

// ExampleCleanup checks the cleanup of a backup
func ExampleCleanup() {
	fakeEngine.cleanup(fakeVol)
	// Output:
	// Command: cleanup --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --force --extra-clean --name Test /foo
	// Mounts: duplicity_cache:/root/.cache/duplicity
}

// ExampleVerify checks the verification of a backup
func ExampleVerify() {
	m, _ := fakeEngine.verify(fakeVol)
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: verify --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --allow-source-mismatch --name Test /foo /back
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="verifyExitCode"} 42
}

// ExampleStatus checks the status of a backup
func ExampleStatus() {
	fakeStdout = "Last full backup date: Mon Jan 2 15:04:05 2006  \nChain end time: Mon Jan 2 15:04:05 2006  \n"
	m, _ := fakeEngine.status(fakeVol)
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
	m, _ := fakeEngine.status(fakeVol)
	fmt.Printf("Metrics: %s\n", strings.Join(m, "\n"))
	// Output:
	// Command: collection-status --s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption --name Test /foo
	// Mounts: /mnt duplicity_cache:/root/.cache/duplicity
	// Metrics: conplicity{volume="Test",what="lastBackup"} 0
	// conplicity{volume="Test",what="lastFullBackup"} 0
}

// TestStatusNoFullBackupDate checks the status of a backup
func TestStatusNoFullBackupDate(t *testing.T) {
	fakeEngine.Handler.Config.Duplicity.Image = "camptocamp/duplicity:latest"
	fakeStdout = "Wrong stdout"
	_, err := fakeEngine.status(fakeVol)

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
	_, err := fakeEngine.status(fakeVol)

	if err == nil {
		t.Fatal("Expected an error, got no error")
	}

	expected := "Failed to parse Duplicity output for chain end time of Test"
	got := err.Error()
	if got != expected {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}
