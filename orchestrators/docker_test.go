package orchestrators

import (
	"errors"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	volumetypes "github.com/docker/docker/api/types/volume"

	"github.com/camptocamp/bivac/config"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/metrics"
	"github.com/camptocamp/bivac/mocks"
	"github.com/camptocamp/bivac/volume"
)

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}

// GetName
func TestDockerGetName(t *testing.T) {
	expectedResult := "Docker"

	o := &DockerOrchestrator{}

	result := o.GetName()

	assert.Equal(t, expectedResult, result, "should be equal")
}

// GetHandler
func TestDockerGetHandler(t *testing.T) {
	fakeHandler := &handler.Bivac{
		Config: &config.Config{},
	}

	expectedResult := fakeHandler

	o := &DockerOrchestrator{
		Handler: fakeHandler,
	}

	result := o.GetHandler()

	assert.Equal(t, expectedResult, result, "should be equal")
}

// GetVolumes
func TestDockerGetVolumes(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	dockerVolumes := volumetypes.VolumesListOKBody{
		Volumes: []*types.Volume{
			&types.Volume{
				Name:       "foo",
				Mountpoint: "/foo",
				Labels: map[string]string{
					"Lorem": "ipsum",
					"dolor": "sit",
					"amet":  ".",
				},
			},
			&types.Volume{
				Name:       "bar",
				Mountpoint: "/var/lib/bar",
			},
		},
	}
	mockDocker.EXPECT().VolumeList(context.Background(), filters.NewArgs()).Return(dockerVolumes, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "foo").Return(types.Volume{
		Name:       "foo",
		Mountpoint: "/foo",
		Labels: map[string]string{
			"Lorem": "ipsum",
			"dolor": "sit",
			"amet":  ".",
		},
	}, nil).Times(1)
	mockDocker.EXPECT().VolumeInspect(context.Background(), "bar").Return(types.Volume{
		Name:       "bar",
		Mountpoint: "/var/lib/bar",
	}, nil).Times(1)

	expectedResult := []*volume.Volume{
		&volume.Volume{
			Config: &volume.Config{},
			MetricsHandler: &metrics.PrometheusMetrics{
				Instance: "fakeNode",
				Volume:   "foo",
				Metrics:  make(map[string]*metrics.Metric),
			},
			Mountpoint: "/foo",
			Name:       "foo",
			Labels: map[string]string{
				"Lorem": "ipsum",
				"dolor": "sit",
				"amet":  ".",
			},
			LabelPrefix: "pref",
		},
		&volume.Volume{
			Config: &volume.Config{},
			MetricsHandler: &metrics.PrometheusMetrics{
				Instance: "fakeNode",
				Volume:   "bar",
				Metrics:  make(map[string]*metrics.Metric),
			},
			Mountpoint:  "/var/lib/bar",
			Name:        "bar",
			Labels:      nil,
			LabelPrefix: "pref",
		},
	}

	o := &DockerOrchestrator{
		Handler: &handler.Bivac{
			Config: &config.Config{
				LabelPrefix: "pref",
			},
			Hostname: "fakeNode",
		},
		Client: mockDocker,
	}
	v, err := o.GetVolumes()

	assert.Nil(t, err)

	assert.Equal(t, expectedResult, v, "should be equal")
}

// GetMountedVolumes
func TestDockerGetMountedVolumes(t *testing.T) {
	// Prepare tests
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	testCases := []struct {
		name               string
		givenContainerList []interface{}
		expected           []interface{}
	}{
		{
			name: "basic context",
			givenContainerList: []interface{}{
				[]types.Container{
					types.Container{
						ID: "foo",
					},
					types.Container{
						ID: "bar",
						Mounts: []types.MountPoint{
							types.MountPoint{
								Type:        "volume",
								Name:        "fakeVolume",
								Destination: "/fakeMountpoint",
							},
						},
					},
				},
				nil,
			},
			expected: []interface{}{
				[]*volume.MountedVolumes{
					&volume.MountedVolumes{
						ContainerID: "foo",
						Volumes:     map[string]string{},
					},
					&volume.MountedVolumes{
						ContainerID: "bar",
						Volumes: map[string]string{
							"fakeVolume": "/fakeMountpoint",
						},
					},
				},
				nil,
			},
		},
		{
			name: "no container",
			givenContainerList: []interface{}{
				[]types.Container{},
				nil,
			},
			expected: []interface{}{
				[]*volume.MountedVolumes(nil),
				nil,
			},
		},
	}

	// Run test cases
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	for _, tc := range testCases {
		mockDocker.EXPECT().ContainerList(context.Background(), types.ContainerListOptions{}).Return(tc.givenContainerList[0], tc.givenContainerList[1]).Times(1)
		result, err := o.GetMountedVolumes()

		assert.Nil(t, err)
		assert.Equal(t, tc.expected[0], result, tc.name)
	}
}

// ContainerExec
func TestDockerContainerExecFailToCreate(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	mockDocker.EXPECT().ContainerExecCreate(
		context.Background(),
		"foo",
		types.ExecConfig{
			Cmd: []string{"lorem", "ipsum"},
		},
	).Return(
		types.IDResponse{},
		errors.New("exec create error"),
	).Times(1)

	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	result := o.ContainerExec(
		&volume.MountedVolumes{
			ContainerID: "foo",
		},
		[]string{"lorem", "ipsum"},
	)

	assert.Contains(t, result.Error(), "exec create error")
}

func TestDockerContainerExecFailToStart(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	mockDocker.EXPECT().ContainerExecCreate(
		context.Background(),
		"foo",
		types.ExecConfig{
			Cmd: []string{"lorem", "ipsum"},
		},
	).Return(
		types.IDResponse{
			ID: "9b996311d5",
		},
		nil,
	).Times(1)
	mockDocker.EXPECT().ContainerExecStart(
		context.Background(),
		"9b996311d5",
		types.ExecStartCheck{},
	).Return(
		errors.New("exec start error"),
	).Times(1)

	// Run test
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	result := o.ContainerExec(
		&volume.MountedVolumes{
			ContainerID: "foo",
		},
		[]string{"lorem", "ipsum"},
	)

	assert.Contains(t, result.Error(), "exec start error")
}

func TestDockerContainerExecFailToInspect(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	mockDocker.EXPECT().ContainerExecCreate(
		context.Background(),
		"foo",
		types.ExecConfig{
			Cmd: []string{"lorem", "ipsum"},
		},
	).Return(
		types.IDResponse{
			ID: "9b996311d5",
		},
		nil,
	).Times(1)
	mockDocker.EXPECT().ContainerExecStart(
		context.Background(),
		"9b996311d5",
		types.ExecStartCheck{},
	).Return(
		nil,
	).Times(1)
	mockDocker.EXPECT().ContainerExecInspect(
		context.Background(),
		"9b996311d5",
	).Return(
		types.ContainerExecInspect{},
		errors.New("exec inspect error"),
	).Times(1)

	// Run test
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	result := o.ContainerExec(
		&volume.MountedVolumes{
			ContainerID: "foo",
		},
		[]string{"lorem", "ipsum"},
	)

	assert.Contains(t, result.Error(), "exec inspect error")
}

func TestDockerContainerExecFailToInspect(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	mockDocker.EXPECT().ContainerExecCreate(
		context.Background(),
		"foo",
		types.ExecConfig{
			Cmd: []string{"lorem", "ipsum"},
		},
	).Return(
		types.IDResponse{
			ID: "9b996311d5",
		},
		nil,
	).Times(1)
	mockDocker.EXPECT().ContainerExecStart(
		context.Background(),
		"9b996311d5",
		types.ExecStartCheck{},
	).Return(
		nil,
	).Times(1)
	mockDocker.EXPECT().ContainerExecInspect(
		context.Background(),
		"9b996311d5",
	).Return(
		types.ContainerExecInspect{},
		errors.New("exec inspect error"),
	).Times(1)

	// Run test
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	result := o.ContainerExec(
		&volume.MountedVolumes{
			ContainerID: "foo",
		},
		[]string{"lorem", "ipsum"},
	)

	assert.Contains(t, result.Error(), "exec inspect error")
}

// backlistedVolume
func TestDockerBlacklistedVolume(t *testing.T) {
	// Prepare tests
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	testCases := []struct {
		name                   string
		configVolumesBlacklist []string
		given                  *volume.Volume
		expected               []interface{}
	}{
		{
			name: "valid volume",
			configVolumesBlacklist: []string{},
			given: &volume.Volume{
				Name: "foo",
				Config: &volume.Config{
					Ignore: false,
				},
			},
			expected: []interface{}{
				false,
				"",
				"",
			},
		},
		{
			name: "unnamed volume",
			configVolumesBlacklist: []string{},
			given: &volume.Volume{
				Name: "acf1e8ec1e87191518f29ff5ef4d983384fd3dc2228265c09bb64b9747e5af67",
				Config: &volume.Config{
					Ignore: false,
				},
			},
			expected: []interface{}{
				true,
				"unnamed",
				"",
			},
		},
		{
			name: "blacklisted volume from global config",
			configVolumesBlacklist: []string{"foo", "bar"},
			given: &volume.Volume{
				Name: "foo",
				Config: &volume.Config{
					Ignore: false,
				},
			},
			expected: []interface{}{
				true,
				"blacklisted",
				"blacklist config",
			},
		},
		{
			name: "blacklisted volume from volume config",
			configVolumesBlacklist: []string{},
			given: &volume.Volume{
				Name: "foo",
				Config: &volume.Config{
					Ignore: true,
				},
			},
			expected: []interface{}{
				true,
				"blacklisted",
				"volume config",
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		o := &DockerOrchestrator{
			Client: mockDocker,
			Handler: &handler.Bivac{
				Config: &config.Config{
					VolumesBlacklist: tc.configVolumesBlacklist,
				},
			},
		}
		result0, result1, result2 := o.blacklistedVolume(tc.given)

		assert.Equal(t, tc.expected[0], result0, tc.name)
		assert.Equal(t, tc.expected[1], result1, tc.name)
		assert.Equal(t, tc.expected[2], result2, tc.name)
	}
}

// pullImage
func TestDockerPullImage(t *testing.T) {
	// Prepare tests
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	testCases := []struct {
		name                        string
		givenImageInspectWithRaw    []interface{}
		returnedImageInspectWithRaw []interface{}
		timesImageInspectWithRaw    int
		givenImagePull              []interface{}
		returnedImagePull           []interface{}
		timesImagePull              int
		given                       string
		expected                    error
	}{
		{
			name: "already pulled",
			givenImageInspectWithRaw: []interface{}{
				context.Background(),
				"foo",
			},
			returnedImageInspectWithRaw: []interface{}{
				types.ImageInspect{},
				[]byte(``),
				nil,
			},
			timesImageInspectWithRaw: 1,
			givenImagePull: []interface{}{
				context.Background(),
				"foo",
				types.ImagePullOptions{},
			},
			returnedImagePull: []interface{}{
				nil,
				errors.New("toto"),
			},
			timesImagePull: 0,
			given:          "foo",
		},
		{
			name: "success pull",
			givenImageInspectWithRaw: []interface{}{
				context.Background(),
				"foo",
			},
			returnedImageInspectWithRaw: []interface{}{
				types.ImageInspect{},
				[]byte(``),
				errors.New("fake error"),
			},
			timesImageInspectWithRaw: 1,
			givenImagePull: []interface{}{
				context.Background(),
				"foo",
				types.ImagePullOptions{},
			},
			returnedImagePull: []interface{}{
				ioutil.NopCloser(strings.NewReader("foobar")),
				nil,
			},
			timesImagePull: 1,
			given:          "foo",
		},
	}

	// Run test cases
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	for _, tc := range testCases {
		mockDocker.EXPECT().ImageInspectWithRaw(
			tc.givenImageInspectWithRaw[0],
			tc.givenImageInspectWithRaw[1],
		).Return(
			tc.returnedImageInspectWithRaw[0],
			tc.returnedImageInspectWithRaw[1],
			tc.returnedImageInspectWithRaw[2],
		).Times(tc.timesImageInspectWithRaw)
		mockDocker.EXPECT().ImagePull(
			tc.givenImagePull[0],
			tc.givenImagePull[1],
			tc.givenImagePull[2],
		).Return(
			tc.returnedImagePull[0],
			tc.returnedImagePull[1],
		).Times(tc.timesImagePull)

		result := o.pullImage(tc.given)

		assert.Equal(t, tc.expected, result, tc.name)
	}
}

// removeContainer
func TestDockerRemoveContainer(t *testing.T) {
	// Prepare tests
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDocker := mocks.NewDocker(mockCtrl)

	testCases := []struct {
		name                    string
		givenContainerRemove    []interface{}
		returnedContainerRemove error
		given                   string
		expected                error
	}{
		{
			name: "basic",
			givenContainerRemove: []interface{}{
				context.Background(),
				"foo",
				types.ContainerRemoveOptions{
					Force:         true,
					RemoveVolumes: true,
				},
			},
			returnedContainerRemove: nil,
			given:    "foo",
			expected: nil,
		},
	}

	// Run test cases
	o := &DockerOrchestrator{
		Client: mockDocker,
	}

	for _, tc := range testCases {
		mockDocker.EXPECT().ContainerRemove(
			tc.givenContainerRemove[0],
			tc.givenContainerRemove[1],
			tc.givenContainerRemove[2],
		).Return(
			tc.returnedContainerRemove,
		).Times(1)
		result := o.removeContainer(tc.given)

		assert.Equal(t, tc.expected, result, tc.name)
	}
}
