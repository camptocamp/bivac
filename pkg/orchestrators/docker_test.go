package orchestrators

import (
	"testing"

	"github.com/camptocamp/bivac/pkg/volume"
	"github.com/stretchr/testify/assert"
)

// backlistedVolume
func TestDockerBlacklistedVolume(t *testing.T) {
	testCases := []struct {
		name                   string
		configVolumesBlacklist []string
		givenVolume            *volume.Volume
		givenFilters           volume.Filters
		expected               []interface{}
	}{
		{
			name:                   "valid volume",
			configVolumesBlacklist: []string{},
			givenVolume: &volume.Volume{
				Name: "foo",
			},
			givenFilters: volume.Filters{},
			expected: []interface{}{
				false,
				"",
				"",
			},
		},
		{
			name:                   "unnamed volume",
			configVolumesBlacklist: []string{},
			givenVolume: &volume.Volume{
				Name: "acf1e8ec1e87191518f29ff5ef4d983384fd3dc2228265c09bb64b9747e5af67",
			},
			givenFilters: volume.Filters{},
			expected: []interface{}{
				true,
				"unnamed",
				"",
			},
		},
		{
			name:                   "blacklisted volume from global config",
			configVolumesBlacklist: []string{"foo", "bar"},
			givenVolume: &volume.Volume{
				Name: "foo",
			},
			givenFilters: volume.Filters{
				Blacklist: []string{"foo"},
			},
			expected: []interface{}{
				true,
				"blacklisted",
				"blacklist config",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		o := &DockerOrchestrator{}

		result0, result1, result2 := o.blacklistedVolume(tc.givenVolume, tc.givenFilters)

		assert.Equal(t, tc.expected[0], result0, tc.name)
		assert.Equal(t, tc.expected[1], result1, tc.name)
		assert.Equal(t, tc.expected[2], result2, tc.name)
	}
}
