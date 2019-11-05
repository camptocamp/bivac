package manager

import (
	"testing"
	"time"

	"github.com/camptocamp/bivac/pkg/volume"

	"github.com/stretchr/testify/assert"
)

// isBackupNeeded
func TestIsBackupNeededBackupInterval(t *testing.T) {
	givenVolume := &volume.Volume{
		BackingUp:      false,
		LastBackupDate: time.Now().Add(time.Hour * -2).Format("2006-01-02 15:04:05"),
		Name:           "foo",
		Hostname:       "bar",
	}

	h, _ := time.ParseDuration("30m")
	assert.Equal(t, isBackupNeeded(givenVolume, h), true)
	h, _ = time.ParseDuration("12h")
	assert.Equal(t, isBackupNeeded(givenVolume, h), false)
}
