package manager

import (
	"testing"
	"time"

	"github.com/camptocamp/bivac/pkg/volume"

	"github.com/stretchr/testify/assert"
)

// isBackupNeeded
func TestIsBackupNeededBackupIntervalStatusSuccess(t *testing.T) {
	givenVolume := &volume.Volume{
		BackingUp:        false,
		LastBackupDate:   time.Now().UTC().Add(time.Hour * -2).Format("2006-01-02 15:04:05"),
		LastBackupStatus: "Success",
		Name:             "foo",
		Hostname:         "bar",
	}

	h, _ := time.ParseDuration("30m")
  d, _ := time.ParseDuration("0s") // default prefered time
	assert.Equal(t, isBackupNeeded(givenVolume, h,d), true)
	h, _ = time.ParseDuration("12h")
	assert.Equal(t, isBackupNeeded(givenVolume, h,d), false)
}

func TestIsBackupNeededBackupIntervalStatusFailed(t *testing.T) {
	givenVolume := &volume.Volume{
		BackingUp:        false,
		LastBackupDate:   time.Now().UTC().Add(time.Hour * -2).Format("2006-01-02 15:04:05"),
		LastBackupStatus: "Failed",
		Name:             "foo",
		Hostname:         "bar",
	}

	h, _ := time.ParseDuration("30m")
  d, _ := time.ParseDuration("0s") // default prefered time
	assert.Equal(t, isBackupNeeded(givenVolume, h,d), true)
	h, _ = time.ParseDuration("12h")
	assert.Equal(t, isBackupNeeded(givenVolume, h,d), true)
}
