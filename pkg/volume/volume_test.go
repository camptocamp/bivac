package volume

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fakeHostname, _ = os.Hostname()

// SetupMetrics
func TestSetupMetrics(t *testing.T) {
	v := Volume{
		ID:         "bar",
		Name:       "bar",
		Mountpoint: "/bar",
		HostBind:   fakeHostname,
		Hostname:   fakeHostname,
		Logs:       make(map[string]string),
		BackingUp:  false,
		RepoName:   "bar",
		SubPath:    "",
	}
	v.SetupMetrics()
	assert.Equal(t, v.ID, "bar")
}
