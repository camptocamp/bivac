package orchestrators

import (
	"testing"

	"github.com/camptocamp/bivac/config"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"
)

var fakeVolume = volume.Volume{
	Name: "foo_bar",
	Config: &volume.Config{
		Ignore: false,
	},
}

func TestKubernetesGetName(t *testing.T) {
	expectedResult := "Kubernetes"

	result := (&KubernetesOrchestrator{}).GetName()
	if result != expectedResult {
		t.Fatalf("Expected %s, got %s", expectedResult, result)
	}
}

func TestBlacklistedVolume(t *testing.T) {
	fakeVolume.Name = "duplicity_cache"
	expectedStatus := true
	expectedReason := "blacklisted"
	expectedSource := "blacklist config"

	// Prepare
	c := handler.Bivac{
		Config: &config.Config{
			VolumesBlacklist: []string{"foo", "bar"},
		},
	}
	o := KubernetesOrchestrator{
		Handler: &c,
	}

	// Run
	status, reason, source := o.blacklistedVolume(&fakeVolume)

	// Check
	if status != expectedStatus {
		t.Fatalf("Expected status: true, got false")
	}
	if reason != expectedReason {
		t.Fatalf("Expected reason: %s, got %s", expectedReason, reason)
	}
	if source != expectedSource {
		t.Fatalf("Expected source: %s, got %s", expectedSource, source)
	}
}

func TestBlacklistedVolumeIgnore(t *testing.T) {
	fakeVolume.Name = "foo_bar"
	fakeVolume.Config.Ignore = true
	expectedStatus := true
	expectedReason := "blacklisted"
	expectedSource := "volume config"

	// Prepare
	c := handler.Bivac{
		Config: &config.Config{
			VolumesBlacklist: []string{"foo", "bar"},
		},
	}
	o := KubernetesOrchestrator{
		Handler: &c,
	}

	// Run
	status, reason, source := o.blacklistedVolume(&fakeVolume)

	// Check
	if status != expectedStatus {
		t.Fatalf("Expected true, got false")
	}
	if reason != expectedReason {
		t.Fatalf("Expected reason: %s, got %s", expectedReason, reason)
	}
	if source != expectedSource {
		t.Fatalf("Expected source: %s, got %s", expectedSource, source)
	}
}

func TestBlacklistedVolumeUnnamed(t *testing.T) {
	fakeVolume.Name = "b9a8145fa9e0b581bbfa90bbe4bd2e49105eba59a9181f77d394e0e6f482333b"
	fakeVolume.Config.Ignore = false
	expectedStatus := true
	expectedReason := "unnamed"
	expectedSource := ""

	// Prepare
	c := handler.Bivac{
		Config: &config.Config{
			VolumesBlacklist: []string{"foo", "bar"},
		},
	}
	o := KubernetesOrchestrator{
		Handler: &c,
	}

	// Run
	status, reason, source := o.blacklistedVolume(&fakeVolume)

	// Check
	if status != expectedStatus {
		t.Fatalf("Expected true, got false")
	}
	if reason != expectedReason {
		t.Fatalf("Expected reason: %s, got %s", expectedReason, reason)
	}
	if source != expectedSource {
		t.Fatalf("Expected source: %s, got %s", expectedSource, source)
	}
}

func TestNotBlacklistedVolume(t *testing.T) {
	fakeVolume.Name = "foo_bar"
	fakeVolume.Config.Ignore = false
	expectedStatus := false
	expectedReason := ""
	expectedSource := ""

	// Prepare
	c := handler.Bivac{
		Config: &config.Config{
			VolumesBlacklist: []string{"foo", "bar"},
		},
	}
	o := KubernetesOrchestrator{
		Handler: &c,
	}

	// Run
	status, reason, source := o.blacklistedVolume(&fakeVolume)

	// Check
	if status != expectedStatus {
		t.Fatalf("Expected false, got true")
	}
	if reason != expectedReason {
		t.Fatalf("Expected reason: %s, got %s", expectedReason, reason)
	}
	if source != expectedSource {
		t.Fatalf("Expected source: %s, got %s", expectedSource, source)
	}
}
