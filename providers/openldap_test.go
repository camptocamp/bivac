package providers

import (
	"testing"
)

func TestOpenLDAPGetName(t *testing.T) {
	expected := "OpenLDAP"
	got := (&OpenLDAPProvider{}).GetName()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestOpenLDAPGetBackupDir(t *testing.T) {
	expected := "backups"
	got := (&OpenLDAPProvider{}).GetBackupDir()
	if expected != got {
		t.Fatalf("Expected %s, got %s", expected, got)
	}
}

func TestOpenLDAPGetPrepareCommand(t *testing.T) {
	mount := "/mnt"

	expected := []string{
		"sh",
		"-c",
		"mkdir -p /mnt/backups && slapcat > /mnt/backups/all.ldif",
	}
	got := (&OpenLDAPProvider{}).GetPrepareCommand(mount)
	if len(got) != 3 {
		t.Fatalf("Expected command to have 3 elements, got %v", len(got))
	} else {
		if expected[2] != got[2] {
			t.Fatalf("Expected %s, got %s", expected, got)
		}
	}
}
