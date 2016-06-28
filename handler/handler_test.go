package handler

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

var fakeHandler = Conplicity{}

func TestSetup(t *testing.T) {
	fakeHandler.Setup("noversion")

	// Check Hostname
	if fakeHandler.Hostname == "" {
		t.Fatal("Hostname should not be nil")
	}

	// Check Client
	expectedEndpoint := "unix:///var/run/docker.sock"
	gotEndpoint := fakeHandler.Client.Endpoint()
	if gotEndpoint != expectedEndpoint {
		t.Fatalf("Expected %s, got %s", expectedEndpoint, gotEndpoint)
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
