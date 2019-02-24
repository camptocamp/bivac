package orchestrators

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	volumetypes "github.com/docker/docker/api/types/volume"
	gomock "github.com/golang/mock/gomock"
	"golang.org/x/net/context"

	"github.com/camptocamp/bivac/mocks"
	"github.com/camptocamp/bivac/pkg/volume"
	"github.com/stretchr/testify/assert"
)

var fakeHostname, _ = os.Hostname()

// NewDockerOrchestrator
func TestDockerNewDockerOrchestrator(t *testing.T) {
}

// GetName
func TestDockerGetNameSuccess(t *testing.T) {
	o := &DockerOrchestrator{}

	name := o.GetName()

	assert.Equal(t, name, "docker")
}

// GetVolumes
func TestDockerGetVolumesSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	dockerVolumes := volumetypes.VolumeListOKBody{
		Volumes: []*types.Volume{
			&types.Volume{
				Name:       "foo",
				Mountpoint: "/foo",
			},
			&types.Volume{
				Name:       "bar",
				Mountpoint: "/bar",
			},
		},
	}

	mockDocker.EXPECT().Info(context.Background()).Return(types.Info{
		Name: fakeHostname,
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeList(context.Background(), filters.NewArgs()).Return(dockerVolumes, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "foo").Return(types.Volume{
		Name:       "foo",
		Mountpoint: "/foo",
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "bar").Return(types.Volume{
		Name:       "bar",
		Mountpoint: "/bar",
	}, nil).Times(1)

	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:         "bar",
			Name:       "bar",
			Mountpoint: "/bar",
			HostBind:   fakeHostname,
			Hostname:   fakeHostname,
			Logs:       make(map[string]string),
			BackingUp:  false,
		},
	}

	o := &DockerOrchestrator{
		client: mockDocker,
	}
	volumes, err := o.GetVolumes(volume.Filters{
		Blacklist: []string{"foo"},
	})

	assert.Nil(t, err)
	assert.Equal(t, volumes, expectedVolumes)
}

func TestDockerGetVolumesBlacklisted(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	dockerVolumes := volumetypes.VolumeListOKBody{
		Volumes: []*types.Volume{
			&types.Volume{
				Name:       "foo",
				Mountpoint: "/foo",
			},
			&types.Volume{
				Name:       "bar",
				Mountpoint: "/bar",
			},
			&types.Volume{
				Name:       "toto",
				Mountpoint: "/toto",
				Labels: map[string]string{
					"bivac.ignore": "true",
				},
			},
		},
	}

	mockDocker.EXPECT().Info(context.Background()).Return(types.Info{
		Name: fakeHostname,
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeList(context.Background(), filters.NewArgs()).Return(dockerVolumes, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "foo").Return(types.Volume{
		Name:       "foo",
		Mountpoint: "/foo",
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "bar").Return(types.Volume{
		Name:       "bar",
		Mountpoint: "/bar",
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "toto").Return(types.Volume{
		Name:       "toto",
		Mountpoint: "/toto",
		Labels: map[string]string{
			"bivac.ignore": "true",
		},
	}, nil).Times(1)

	expectedVolumes := []*volume.Volume{
		&volume.Volume{
			ID:         "foo",
			Name:       "foo",
			Mountpoint: "/foo",
			HostBind:   fakeHostname,
			Hostname:   fakeHostname,
			Logs:       make(map[string]string),
		},
		&volume.Volume{
			ID:         "bar",
			Name:       "bar",
			Mountpoint: "/bar",
			HostBind:   fakeHostname,
			Hostname:   fakeHostname,
			Logs:       make(map[string]string),
		},
	}

	o := &DockerOrchestrator{
		client: mockDocker,
	}
	volumes, err := o.GetVolumes(volume.Filters{})

	assert.Nil(t, err)
	assert.Equal(t, volumes, expectedVolumes)
}

/*************************/
/*                       */
/*      DeployAgent      */
/*                       */
/*************************/
func TestDockerDeployAgentSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	fakeCmd := []string{"agent"}
	fakeEnv := []string{}
	fakeImage := "camptocamp/bivac:fake"
	fakeVolume := &volume.Volume{
		ID:         "foo",
		Name:       "foo",
		Mountpoint: "/foo",
		ReadOnly:   false,
	}

	/*
		containerConfig := &containertypes.Config{
			Cmd:          fakeCmd,
			Env:          gomock.Any(),
			Image:        fakeImage,
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
		}
	*/
	containerHostConfig := &containertypes.HostConfig{
		Mounts: []mount.Mount{
			mount.Mount{
				Type:     "volume",
				Target:   fakeVolume.Mountpoint,
				Source:   fakeVolume.Name,
				ReadOnly: fakeVolume.ReadOnly,
			},
		},
	}

	// PullImage passthrough
	mockDocker.EXPECT().ImageInspectWithRaw(context.Background(), fakeImage).Return(types.ImageInspect{}, make([]byte, 0), nil).Times(1)
	mockDocker.EXPECT().ContainerInspect(context.Background(), gomock.Any()).Return(types.ContainerJSON{
		Mounts: []types.MountPoint{},
	}, nil).Times(1)
	mockDocker.EXPECT().ContainerCreate(context.Background(), gomock.Any(), containerHostConfig, nil, "").Return(containertypes.ContainerCreateCreatedBody{ID: "alpha"}, nil).Times(1)
	mockDocker.EXPECT().ContainerRemove(context.Background(), "alpha", types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}).Return(nil).Times(1)
	mockDocker.EXPECT().ContainerStart(context.Background(), "alpha", types.ContainerStartOptions{}).Return(nil).Times(1)
	mockDocker.EXPECT().ContainerInspect(context.Background(), "alpha").Return(
		types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{
					Status: "exited",
				},
			},
		},
		nil).Times(1)
	mockDocker.EXPECT().ContainerLogs(context.Background(), "alpha", types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Follow:     true,
	}).Return(ioutil.NopCloser(bytes.NewReader([]byte("foo"))), nil).Times(1)

	// Run test
	o := &DockerOrchestrator{
		client: mockDocker,
	}
	success, _, err := o.DeployAgent(fakeImage, fakeCmd, fakeEnv, fakeVolume)

	assert.Nil(t, err)
	assert.True(t, success)
	// TODO: fix assert stdout
	//assert.Equal(t, "foo", stdout)
}

// PullImage
func TestDockerPullImageSuccessNoPull(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	fakeImage := "camptocamp/bivac:fake"

	mockDocker.EXPECT().ImageInspectWithRaw(context.Background(), fakeImage).Return(types.ImageInspect{}, make([]byte, 0), nil).Times(1)

	// Run test
	o := &DockerOrchestrator{
		client: mockDocker,
	}
	err := o.PullImage(fakeImage)

	assert.Nil(t, err)
}

func TestDockerPullImageSuccessPull(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	fakeImage := "camptocamp/bivac:fake"

	mockDocker.EXPECT().ImageInspectWithRaw(context.Background(), fakeImage).Return(types.ImageInspect{}, make([]byte, 0), fmt.Errorf("random error")).Times(1)
	mockDocker.EXPECT().ImagePull(context.Background(), fakeImage, types.ImagePullOptions{}).Return(ioutil.NopCloser(strings.NewReader("foo")), nil).Times(1)

	// Run test
	o := &DockerOrchestrator{
		client: mockDocker,
	}
	err := o.PullImage(fakeImage)

	assert.Nil(t, err)
}

func TestDockerPullImageFailToPull(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewMockCommonAPIClient(mockCtrl)

	fakeImage := "camptocamp/bivac:fake"

	mockDocker.EXPECT().ImageInspectWithRaw(context.Background(), fakeImage).Return(types.ImageInspect{}, make([]byte, 0), fmt.Errorf("random error")).Times(1)
	mockDocker.EXPECT().ImagePull(context.Background(), fakeImage, types.ImagePullOptions{}).Return(ioutil.NopCloser(strings.NewReader("foo")), fmt.Errorf("error")).Times(1)

	// Run test
	o := &DockerOrchestrator{
		client: mockDocker,
	}
	err := o.PullImage(fakeImage)

	assert.Equal(t, err, fmt.Errorf("error"))
}

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
		{
			name:                   "volume config ignore",
			configVolumesBlacklist: []string{},
			givenVolume: &volume.Volume{
				Name: "toto",
				Labels: map[string]string{
					"bivac.ignore": "true",
				},
			},
			givenFilters: volume.Filters{},
			expected: []interface{}{
				true,
				"ignored",
				"volume config",
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
