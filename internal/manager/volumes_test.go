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
func TestRetrieveVolumesBasic(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "localhost",
		},
		&volume.Volume{
			ID:       "bar",
			Name:     "bar",
			HostBind: "localhost",
		},
	}
	givenFilters := volume.Filters{}
	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "localhost",
		},
		&volume.Volume{
			ID:       "bar",
			Name:     "bar",
			HostBind: "localhost",
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetPath(gomock.Any()).Return("localhost").Times(2)
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, nil).Times(1)

	m.Volumes = []*volume.Volume{}
	err := retrieveVolumes(m, givenFilters)

	// Do not manage Metrics field
	// Should be properly fixed
	for k, _ := range m.Volumes {
		m.Volumes[k].Metrics = nil
	}

	assert.Nil(t, err)
	assert.Equal(t, m.Volumes, expectedVolumes)
}

func TestRetrieveVolumesBlacklist(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:   "foo",
			Name: "foo",
		},
		&volume.Volume{
			ID:   "bar",
			Name: "bar",
		},
	}
	givenFilters := volume.Filters{
		Blacklist: []string{"foo"},
	}
	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:   "bar",
			Name: "bar",
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetPath(gomock.Any()).Return("localhost").Times(1)
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, nil).Times(1)

	m.Volumes = []*volume.Volume{}
	err := retrieveVolumes(m, givenFilters)

	// Do not manage Metrics field
	// Should be properly fixed
	for k, _ := range m.Volumes {
		m.Volumes[k].Metrics = nil
	}

	assert.Nil(t, err)
	assert.Equal(t, m.Volumes, expectedVolumes)
}

/*
func TestRetrieveVolumesWhitelist(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:   "foo",
			Name: "foo",
		},
		&volume.Volume{
			ID:   "bar",
			Name: "bar",
		},
	}
	givenFilters := volume.Filters{
		Whitelist: []string{"foo"},
	}
	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:   "foo",
			Name: "foo",
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, nil).Times(1)

	m.Volumes = []*volume.Volume{}
	err := retrieveVolumes(m, givenFilters)

	assert.Nil(t, err)
	assert.Equal(t, m.Volumes, expectedVolumes)
}
*/
func TestRetrieveVolumesOrchestratorError(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:   "foo",
			Name: "foo",
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, fmt.Errorf("error")).Times(1)

	m.Volumes = []*volume.Volume{}
	err := retrieveVolumes(m, volume.Filters{})

	assert.Equal(t, err.Error(), "error")
	assert.Equal(t, m.Volumes, []*volume.Volume{})
}

func TestRetrieveVolumesAppend(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "localhost",
		},
		//&volume.Volume{
		//	ID:   "bar",
		//	Name: "bar",
		//},
	}
	givenFilters := volume.Filters{}
	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "localhost",
		},
		//&volume.Volume{
		//	ID:   "bar",
		//	Name: "bar",
		//},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, nil).Times(1)

	m.Volumes = []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "localhost",
		},
	}
	err := retrieveVolumes(m, givenFilters)

	assert.Nil(t, err)
	assert.Equal(t, m.Volumes, expectedVolumes)
}

func TestRetrieveVolumesRemove(t *testing.T) {
	// Prepare test
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockOrchestrator := mocks.NewMockOrchestrator(mockCtrl)
	mockRegisterer := mocks.NewMockRegisterer(mockCtrl)

	givenVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "bar",
			Name:     "bar",
			HostBind: "bar",
			Hostname: "bar",
		},
	}
	givenFilters := volume.Filters{}
	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:       "bar",
			Name:     "bar",
			HostBind: "bar",
			Hostname: "bar",
		},
	}

	m := &Manager{
		Orchestrator: mockOrchestrator,
	}

	// Run test
	mockOrchestrator.EXPECT().GetVolumes(volume.Filters{}).Return(givenVolumes, nil).Times(1)
	mockRegisterer.EXPECT().Unregister(gomock.Any()).Return(true).AnyTimes()

	m.Volumes = []*volume.Volume{
		&volume.Volume{
			ID:       "foo",
			Name:     "foo",
			HostBind: "foo",
			Hostname: "foo",
		},
		&volume.Volume{
			ID:       "bar",
			Name:     "bar",
			HostBind: "bar",
			Hostname: "bar",
		},
		&volume.Volume{
			ID:       "fake",
			Name:     "fake",
			HostBind: "fake",
			Hostname: "fake",
		},
	}

	for _, v := range m.Volumes {
		v.SetupMetrics()
	}

	err := retrieveVolumes(m, givenFilters)

	// Do not manage Metrics field
	// Should be properly fixed
	for k, _ := range m.Volumes {
		m.Volumes[k].Metrics = nil
	}

	assert.Nil(t, err)
	assert.Equal(t, m.Volumes, expectedVolumes)
}

// backlistedVolume
func TestBlacklistedVolumeValid(t *testing.T) {
	givenVolume := &volume.Volume{
		ID:   "foo",
		Name: "foo",
	}
	givenFilters := volume.Filters{}

	// Run test
	result0, result1, result2 := blacklistedVolume(givenVolume, givenFilters)

	assert.Equal(t, result0, false)
	assert.Equal(t, result1, "")
	assert.Equal(t, result2, "")
}

func TestBlacklistedVolumeUnnamedVolume(t *testing.T) {
	givenVolume := &volume.Volume{
		ID:   "acf1e8ec1e87191518f29ff5ef4d983384fd3dc2228265c09bb64b9747e5af67",
		Name: "acf1e8ec1e87191518f29ff5ef4d983384fd3dc2228265c09bb64b9747e5af67",
	}
	givenFilters := volume.Filters{}

	// Run test
	result0, result1, result2 := blacklistedVolume(givenVolume, givenFilters)

	assert.Equal(t, result0, true)
	assert.Equal(t, result1, "unnamed")
	assert.Equal(t, result2, "")
}

func TestBlacklistedVolumeBlacklisted(t *testing.T) {
	givenVolume := &volume.Volume{
		ID:   "foo",
		Name: "foo",
	}
	givenFilters := volume.Filters{
		Blacklist: []string{"foo"},
	}

	// Run test
	result0, result1, result2 := blacklistedVolume(givenVolume, givenFilters)

	assert.Equal(t, result0, true)
	assert.Equal(t, result1, "blacklisted")
	assert.Equal(t, result2, "blacklist config")
}

func TestBlacklistedVolumeWhitelisted(t *testing.T) {
	givenVolume := &volume.Volume{
		ID:   "foo",
		Name: "foo",
	}
	givenFilters := volume.Filters{
		Whitelist: []string{"foo"},
	}

	// Run test
	result0, result1, result2 := blacklistedVolume(givenVolume, givenFilters)

	assert.Equal(t, result0, false)
	assert.Equal(t, result1, "")
	assert.Equal(t, result2, "")
}

func TestBlacklistedVolumeBlacklistedBecauseWhitelist(t *testing.T) {
	givenVolume := &volume.Volume{
		ID:   "foo",
		Name: "foo",
	}
	givenFilters := volume.Filters{
		Whitelist: []string{"bar"},
	}

	// Run test
	result0, result1, result2 := blacklistedVolume(givenVolume, givenFilters)

	assert.Equal(t, result0, true)
	assert.Equal(t, result1, "blacklisted")
	assert.Equal(t, result2, "whitelist config")
}
