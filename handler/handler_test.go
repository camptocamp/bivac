package handler

import "testing"

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
}
