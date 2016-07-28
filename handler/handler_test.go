package handler

import (
	"os"
	"testing"

	"golang.org/x/net/context"

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

	// Check Client
	expectedInfo, _ := os.Hostname()
	gotInfo, _ := fakeHandler.Client.Info(context.Background())
	if gotInfo.Name != expectedInfo {
		t.Fatalf("Expected %s, got %s", expectedInfo, gotInfo.Name)
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
