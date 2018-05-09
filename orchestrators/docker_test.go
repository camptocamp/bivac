package orchestrators

import (
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"reflect"
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

	if result != expectedResult {
		t.Fatalf("Expected %s, got %s", expectedResult, result)
	}
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

	if !reflect.DeepEqual(expectedResult, result) {
		t.Fatalf("Expected %+v, got %+v", expectedResult, result)
	}
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

/*
func TestDockerGetVolumesFailToParseVolumeList(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Volumes": "foo",
			"Warnings": []
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, err := o.GetVolumes()

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerGetVolumesFailToInspectVolume(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Volumes": [
				{
					"Name": "tardis",
					"Driver": "local",
					"Mountpoint": "/var/lib/docker/volumes/tardis",
					"Labels": null,
					"Scope": "local"
				}
			],
			"Warnings": []
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/volumes/tardis", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Name": [],
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, err := o.GetVolumes()

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerGetVolumesSuccess(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Volumes": [
				{
					"Name": "tardis",
					"Driver": "local",
					"Mountpoint": "/var/lib/docker/volumes/tardis",
					"Labels": null,
					"Scope": "local"
				}
			],
			"Warnings": []
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/volumes/tardis", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Name": "tardis",
			"Driver": "custom",
			"Mountpoint": "/var/lib/docker/volumes/tardis/_data",
			"Status": {
				"hello": "world"
			},
			"Scope": "local"
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	volumes, _ := o.GetVolumes()

	srv.Shutdown(nil)

	if len(volumes) != 1 {
		t.Fatalf("Expected 1 volume, got %d", len(volumes))
	}
}

func TestDockerGetVolumesBlacklistedVolume(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Volumes": [
				{
					"Name": "tardis",
					"Driver": "local",
					"Mountpoint": "/var/lib/docker/volumes/tardis",
					"Labels": null,
					"Scope": "local"
				}
			],
			"Warnings": []
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/volumes/tardis", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Name": "99da9cf4d9cc6532404260cc05283ce14429015983e921c3f40f8ba423752001",
			"Driver": "custom",
			"Mountpoint": "/var/lib/docker/volumes/tardis/_data"
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	volumes, _ := o.GetVolumes()

	srv.Shutdown(nil)

	if len(volumes) != 0 {
		t.Fatalf("Expected 0 volume, got %d", len(volumes))
	}
}

// LaunchContainer
func TestDockerLaunchContainerPullImageFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, _, err := o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerLaunchContainerContainerCreateFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, _, err := o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerLaunchContainerEnvMapping(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		expectedResult := []string{
			"foo=bar",
			"fake=ekaf",
		}
		body, _ := ioutil.ReadAll(r.Body)
		var container *container.Config
		json.Unmarshal(body, &container)

		if !sameStringSlice(expectedResult, container.Env) {
			t.Fatalf("Expected %+v, got %+v", expectedResult, container.Env)
		}
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	env := map[string]string{
		"foo":  "bar",
		"fake": "ekaf",
	}
	o.LaunchContainer("foo", env, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)
}

func TestDockerLaunchContainerMountedVolumesMapping(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		expectedResult := []string{
			"foo:/foo:ro",
			"bar:/bar",
		}
		body, _ := ioutil.ReadAll(r.Body)
		var container *dockerConfigWrapper
		json.Unmarshal(body, &container)

		if !sameStringSlice(expectedResult, container.HostConfig.Binds) {
			t.Fatalf("Expected %+v, got %+v", expectedResult, container.HostConfig.Binds)
		}
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	v := []*volume.Volume{
		&volume.Volume{
			Mountpoint: "/foo",
			Name:       "foo",
			ReadOnly:   true,
		},
		&volume.Volume{
			Mountpoint: "/bar",
			Name:       "bar",
			ReadOnly:   false,
		},
	}
	o.LaunchContainer("foo", map[string]string{}, []string{}, v)

	srv.Shutdown(nil)
}

func TestDockerLaunchContainerContainerStartFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)
}

func TestDockerLaunchContainerContainerInspectFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/foo/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, _, err := o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerLaunchContainerContainerLogsFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/foo/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Config": {
				"Hostname": "ba033ac44011",
				"Image": "ubuntu",
				"StopSignal": "SIGTERM"
			},
			"Id": "foo",
			"Image": "bar",
			"Name": "foo",
			"Path": "/bin/sh",
			"State": {
				"ExitCode": 9,
				"Status": "exited"
			}
		}
		`)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/logs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, _, err := o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerLaunchContainerSuccess(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/create", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/containers/foo/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Config": {
				"Hostname": "ba033ac44011",
				"Image": "ubuntu",
				"StopSignal": "SIGTERM"
			},
			"Id": "foo",
			"Image": "bar",
			"Name": "foo",
			"Path": "/bin/sh",
			"State": {
				"ExitCode": 0,
				"Status": "exited"
			}
		}
		`)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/containers/foo/logs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
		w.Header().Set("Connection", "Upgrade")
		w.Header().Set("Upgrade", "tcp")
		w.Write([]byte(""))
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, _, err := o.LaunchContainer("foo", map[string]string{}, []string{}, []*volume.Volume{})

	srv.Shutdown(nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
}

// GetMountedVolumes
func TestDockerGetMountedVolumesFailToListContainers(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		[
			{
				"Image": "ubuntu:latest",
				"ImageID": "d74508fb6632491cea586a1fd7d748dfc5274cd6fdfedee309ecdcbc2bf5cb82"
			}
		]
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = dockerEndpoint
	o := NewDockerOrchestrator(c)
	_, err := o.GetMountedVolumes()

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerGetMountedVolumesListContainers(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		[
			{
				"Id": "8dfafdbc3a40",
				"Names":["/boring_feynman"],
				"Image": "ubuntu:latest",
				"Mounts": [
					{
						"Type": "volume",
						"Name": "foo",
						"Source": "/data",
						"Destination": "/data",
						"Driver": "local",
						"Mode": "ro,Z",
						"RW": false,
						"Propagation": ""
					}
				]
			}
		]
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	expectedResult := []*volume.MountedVolumes{
		&volume.MountedVolumes{
			ContainerID: "8dfafdbc3a40",
			Volumes: map[string]string{
				"foo": "/data",
			},
		},
	}

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	result, _ := o.GetMountedVolumes()

	srv.Shutdown(nil)

	if !reflect.DeepEqual(expectedResult, result) {
		t.Fatalf("Expected %+v, got %+v", expectedResult, result)
	}
}

// ContainerExec
func TestDockerContainerExecCreateFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo/exec", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"

	mountedVolumes := &volume.MountedVolumes{
		ContainerID: "foo",
	}

	o := NewDockerOrchestrator(c)
	err := o.ContainerExec(mountedVolumes, []string{"sh"})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerContainerExecStartFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo/exec", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/exec/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"

	mountedVolumes := &volume.MountedVolumes{
		ContainerID: "foo",
	}

	o := NewDockerOrchestrator(c)
	err := o.ContainerExec(mountedVolumes, []string{"sh"})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerContainerExecInspectFailed(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo/exec", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/exec/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
		w.Write([]byte(""))
	})
	r.HandleFunc("/exec/foo/json", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"

	mountedVolumes := &volume.MountedVolumes{
		ContainerID: "foo",
	}

	o := NewDockerOrchestrator(c)
	err := o.ContainerExec(mountedVolumes, []string{"sh"})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerContainerExecInvalidExitCode(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo/exec", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/exec/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
		w.Write([]byte(""))
	})
	r.HandleFunc("/exec/foo/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"ContainerID": "foo",
			"ExitCode": 2,
			"Running": false
		}
		`)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"

	mountedVolumes := &volume.MountedVolumes{
		ContainerID: "foo",
	}

	o := NewDockerOrchestrator(c)
	err := o.ContainerExec(mountedVolumes, []string{"sh"})

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}

func TestDockerContainerExecValidExitCode(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo/exec", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id": "foo",
			"Warnings":[]
		}
		`)
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	r.HandleFunc("/exec/foo/start", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/vnd.docker.raw-stream")
		w.Write([]byte(""))
	})
	r.HandleFunc("/exec/foo/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"ContainerID": "foo",
			"ExitCode": 0,
			"Running": false
		}
		`)
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"

	mountedVolumes := &volume.MountedVolumes{
		ContainerID: "foo",
	}

	o := NewDockerOrchestrator(c)
	err := o.ContainerExec(mountedVolumes, []string{"sh"})

	srv.Shutdown(nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
}

// blacklistedVolume
func TestDockerBlacklistedVolumeUnnamed(t *testing.T) {
	expectedA := true
	expectedB := "unnamed"
	expectedC := ""
	v := &volume.Volume{
		Name: "duplicity_cache",
	}

	o := &DockerOrchestrator{}
	a, b, c := o.blacklistedVolume(v)

	if expectedA != a || expectedB != b || expectedC != c {
		t.Fatalf("Expected (%v, %v, %s), got (%v, %v, %v)", expectedA, expectedB, expectedC, a, b, c)
	}
}

func TestDockerBlacklistedVolumeBlacklisted(t *testing.T) {
	expectedA := true
	expectedB := "blacklisted"
	expectedC := "blacklist config"
	v := &volume.Volume{
		Name: "foo",
	}

	o := &DockerOrchestrator{
		Handler: &handler.Bivac{
			Config: &config.Config{
				VolumesBlacklist: []string{"foo", "bar"},
			},
		},
	}
	a, b, c := o.blacklistedVolume(v)

	if expectedA != a || expectedB != b || expectedC != c {
		t.Fatalf("Expected (%v, %v, %s), got (%v, %v, %v)", expectedA, expectedB, expectedC, a, b, c)
	}
}

func TestDockerBlacklistedVolumeIgnored(t *testing.T) {
	expectedA := true
	expectedB := "blacklisted"
	expectedC := "volume config"
	v := &volume.Volume{
		Name: "foo",
		Config: &volume.Config{
			Ignore: true,
		},
	}

	o := &DockerOrchestrator{
		Handler: &handler.Bivac{
			Config: &config.Config{
				VolumesBlacklist: []string{"bar"},
			},
		},
	}
	a, b, c := o.blacklistedVolume(v)

	if expectedA != a || expectedB != b || expectedC != c {
		t.Fatalf("Expected (%v, %v, %s), got (%v, %v, %v)", expectedA, expectedB, expectedC, a, b, c)
	}
}

// pullImage
func TestDockerPullImageSuccessPull(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	err := pullImage(o.Client, "foo")

	srv.Shutdown(nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

}

func TestDockerPullImageAlreadyPull(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/foo/json", func(w http.ResponseWriter, r *http.Request) {
		content := []byte(`
		{
			"Id" : "sha256:85f05633ddc1c50679be2b16a0479ab6f7637f8884e0cfe0f4d20e1ebb3d6e7c",
			"Container" : "cb91e48a60d01f1e27028b4fc6819f4f290b3cf12496c8176ec714d0d390984a"
		}
		`)
		w.Header().Set("Content-Type", "application/json")
		w.Write(content)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	err := pullImage(o.Client, "foo")

	srv.Shutdown(nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}

}

func TestDockerPullImageFailToPull(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/images/create", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	err := pullImage(o.Client, "foo")

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}

}

// removeContainer
func TestDockerRemoveContainerSuccess(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	err := removeContainer(o.Client, "foo")

	srv.Shutdown(nil)

	if err != nil {
		t.Fatalf("Expected no error, got %s", err)
	}
}

func TestDockerRemoveContainerFail(t *testing.T) {
	r := mux.NewRouter()
	r.HandleFunc("/containers/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
	})
	srv := startFakeHTTPServer(r)

	c := &handler.Bivac{
		Config: &config.Config{},
	}
	c.Config.Docker.Endpoint = "http://127.0.0.1:9878"
	o := NewDockerOrchestrator(c)
	err := removeContainer(o.Client, "foo")

	srv.Shutdown(nil)

	if err == nil {
		t.Fatalf("Invalid content provided but no error raised.")
	}
}
*/
