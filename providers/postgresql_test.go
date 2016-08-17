package providers

import (
	"testing"

	"github.com/docker/engine-api/types"
)

func TestPostgreSQLGetName(t *testing.T) {
	expected := "PostgreSQL"
	got := (&PostgreSQLProvider{}).GetName()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestPostgreSQLGetBackupDir(t *testing.T) {
	expected := "backups"
	got := (&PostgreSQLProvider{}).GetBackupDir()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestPostgreSQLGetPrepareCommand(t *testing.T) {
	mount := &types.MountPoint{
		Destination: "/mnt",
	}

	expected := []string{
		"sh",
		"-c",
		"mkdir -p /mnt/backups && pg_dumpall -Upostgres > /mnt/backups/all.sql",
	}
	got := (&PostgreSQLProvider{}).GetPrepareCommand(mount)
	if len(got) != 3 {
		t.Fatalf("Expected command to have 3 elements, got %v", len(got))
	} else {
		if expected[2] != got[2] {
			t.Fatalf("Expected %s, got %s", expected, got)
		}
	}
}
