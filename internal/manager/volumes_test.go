package manager

import (
	"fmt"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/camptocamp/bivac/mocks"
	"github.com/camptocamp/bivac/pkg/volume"
)

// retrieveVolumes
func TestRetrieveVolumes(t *testing.T) {
	// Prepare tests
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	testCases := []struct {
		name            string
		givenVolumes    []*volume.Volume
		givenError      error
		givenManager    *Manager
		givenFilters    volume.Filters
		expectedError   error
		expectedVolumes []*volume.Volume
	}{
		{
			name:       "all valid volumes",
			givenError: nil,
			givenVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
				&volume.Volume{
					Name: "bar",
				},
			},
			givenFilters:  volume.Filters{},
			expectedError: nil,
			expectedVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
				&volume.Volume{
					Name: "bar",
				},
			},
		},
		{
			name:       "one valid, one blacklisted",
			givenError: nil,
			givenVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
				&volume.Volume{
					Name: "bar",
				},
			},
			givenFilters: volume.Filters{
				Blacklist: []string{"bar"},
			},
			expectedError: nil,
			expectedVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
			},
		},
		{
			name:       "one valid, one blacklisted, one whitelisted",
			givenError: nil,
			givenVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
				&volume.Volume{
					Name: "bar",
				},
				&volume.Volume{
					Name: "fake",
				},
			},
			givenFilters: volume.Filters{
				Blacklist: []string{"bar"},
				Whitelist: []string{"fake"},
			},
			expectedError: nil,
			expectedVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "fake",
				},
			},
		},
		{
			name:       "one valid but orchestrator error",
			givenError: fmt.Errorf("error"),
			givenVolumes: []*volume.Volume{
				&volume.Volume{
					Name: "foo",
				},
			},
			givenFilters:    volume.Filters{},
			expectedError:   fmt.Errorf("error"),
			expectedVolumes: []*volume.Volume{},
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test cases
	for _, tc := range testCases {
		mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(tc.givenVolumes, tc.givenError).Times(1)

		err := retrieveVolumes(m, tc.givenFilters)

		assert.Equal(t, tc.expectedError, err, tc.name)
		assert.Equal(t, tc.expectedVolumes, m.Volumes, tc.name)
	}
}

// backlistedVolume
func TestBlacklistedVolume(t *testing.T) {
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
		result0, result1, result2 := blacklistedVolume(tc.givenVolume, tc.givenFilters)

		assert.Equal(t, tc.expected[0], result0, tc.name)
		assert.Equal(t, tc.expected[1], result1, tc.name)
		assert.Equal(t, tc.expected[2], result2, tc.name)
	}
}
