package orchestrators

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/context"

	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/util"
	"github.com/camptocamp/bivac/volume"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"

	log "github.com/Sirupsen/logrus"
	docker "github.com/docker/docker/client"
)

// DockerOrchestrator implements a container orchestrator for Docker
type DockerOrchestrator struct {
	Handler *handler.Bivac
	Client  *docker.Client
}

// NewDockerOrchestrator creates a Docker client
func NewDockerOrchestrator(c *handler.Bivac) (o *DockerOrchestrator) {
	var err error
	o = &DockerOrchestrator{
		Handler: c,
	}
	o.Client, err = docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	util.CheckErr(err, "failed to create a Docker client: %v", "fatal")
	return
}

// GetName returns the orchestrator name
func (*DockerOrchestrator) GetName() string {
	return "Docker"
}

// GetPath returns the path of the backup
func (*DockerOrchestrator) GetPath(v *volume.Volume) string {
	return v.Hostname + "/" + v.Name
}

// GetHandler returns the Orchestrator's handler
func (o *DockerOrchestrator) GetHandler() *handler.Bivac {
	return o.Handler
}

// GetVolumes returns the Docker volumes, inspected and filtered
func (o *DockerOrchestrator) GetVolumes() (volumes []*volume.Volume, err error) {
	c := o.Handler
	vols, err := o.Client.VolumeList(context.Background(), filters.NewArgs())
	if err != nil {
		err = fmt.Errorf("Failed to list Docker volumes: %v", err)
		return
	}
	for _, vol := range vols.Volumes {
		var voll types.Volume
		voll, err = o.Client.VolumeInspect(context.Background(), vol.Name)
		if err != nil {
			err = fmt.Errorf("Failed to inspect volume %s: %v", vol.Name, err)
			return
		}

		nv := &volume.Volume{
			Config:      &volume.Config{},
			Mountpoint:  voll.Mountpoint,
			Name:        voll.Name,
			Labels:      voll.Labels,
			LabelPrefix: c.Config.LabelPrefix,
			Hostname:    c.Hostname,
		}

		v := volume.NewVolume(nv, c.Config, c.Hostname)
		if b, r, s := o.blacklistedVolume(v); b {
			log.WithFields(log.Fields{
				"volume": vol.Name,
				"reason": r,
				"source": s,
			}).Info("Ignoring volume")
			continue
		}
		volumes = append(volumes, v)
	}
	return
}

// LaunchContainer starts a container using the Docker orchestrator
func (o *DockerOrchestrator) LaunchContainer(image string, env map[string]string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	err = pullImage(o.Client, image)
	if err != nil {
		err = fmt.Errorf("failed to pull image: %v", err)
		return
	}

	var envVars []string
	for envName, envValue := range env {
		envVars = append(envVars, envName+"="+envValue)
	}

	log.WithFields(log.Fields{
		"image":       image,
		"command":     strings.Join(cmd, " "),
		"environment": strings.Join(envVars, ", "),
	}).Debug("Creating container")

	var mounts []mount.Mount

	for _, v := range volumes {
		m := mount.Mount{
			Type:     "volume",
			Target:   v.Mountpoint,
			Source:   v.Name,
			ReadOnly: v.ReadOnly,
		}
		mounts = append(mounts, m)
	}

	container, err := o.Client.ContainerCreate(
		context.Background(),
		&container.Config{
			Cmd:          cmd,
			Env:          envVars,
			Image:        image,
			OpenStdin:    true,
			StdinOnce:    true,
			AttachStdin:  true,
			AttachStdout: true,
			AttachStderr: true,
			Tty:          false,
		},
		&container.HostConfig{
			Mounts: mounts,
		}, nil, "",
	)
	if err != nil {
		err = fmt.Errorf("failed to create container: %v", err)
		return
	}
	defer removeContainer(o.Client, container.ID)

	log.Debugf("Launching with '%v'...", strings.Join(cmd, " "))
	err = o.Client.ContainerStart(context.Background(), container.ID, types.ContainerStartOptions{})
	if err != nil {
		err = fmt.Errorf("failed to start container: %v", err)
	}

	var exited bool

	for !exited {
		var cont types.ContainerJSON
		cont, err = o.Client.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %v", err)
			return
		}

		if cont.State.Status == "exited" {
			exited = true
			state = cont.State.ExitCode
		}
	}

	body, err := o.Client.ContainerLogs(context.Background(), container.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Details:    true,
		Follow:     true,
	})
	if err != nil {
		err = fmt.Errorf("failed to retrieve logs: %v", err)
		return
	}

	stdoutput := new(bytes.Buffer)
	defer body.Close()
	_, err = stdcopy.StdCopy(stdoutput, ioutil.Discard, body)
	if err != nil {
		err = fmt.Errorf("failed to read logs from response: %v", err)
		return
	}
	stdout = stdoutput.String()
	log.Debug(stdout)

	return
}

// GetMountedVolumes returns mounted volumes
func (o *DockerOrchestrator) GetMountedVolumes(v *volume.Volume) (containers []*volume.MountedVolumes, err error) {
	c, err := o.Client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		err = fmt.Errorf("failed to list containers: %v", err)
		return
	}

	for _, container := range c {
		mv := &volume.MountedVolumes{
			ContainerID: container.ID,
			Volumes:     make(map[string]string),
		}
		for _, mount := range container.Mounts {
			if mount.Type == "volume" {
				mv.Volumes[mount.Name] = mount.Destination
			}
		}
		containers = append(containers, mv)
	}
	return
}

// ContainerExec executes a command in a container
func (o *DockerOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolumes, command []string) (stdout string, err error) {
	exec, err := o.Client.ContainerExecCreate(context.Background(), mountedVolumes.ContainerID, types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          command,
	})
	if err != nil {
		err = fmt.Errorf("failed to create exec: %v", err)
	}

	conn, err := o.Client.ContainerExecAttach(context.Background(), exec.ID, types.ExecConfig{})
	if err != nil {
		err = fmt.Errorf("failed to attach: %s", err)
	}
	defer conn.Close()

	err = o.Client.ContainerExecStart(context.Background(), exec.ID, types.ExecStartCheck{})
	if err != nil {
		err = fmt.Errorf("failed to start exec: %v", err)
	}

	stdoutput := new(bytes.Buffer)
	stdcopy.StdCopy(stdoutput, ioutil.Discard, conn.Reader)

	stdout = stdoutput.String()
	return
}

func (o *DockerOrchestrator) blacklistedVolume(vol *volume.Volume) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || vol.Name == "duplicity_cache" || vol.Name == "lost+found" {
		return true, "unnamed", ""
	}

	// Use whitelist if defined
	if l := o.Handler.Config.VolumesWhitelist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return false, "", ""
		}
		return true, "blacklisted", "whitelist config"
	}

	list := o.Handler.Config.VolumesBlacklist
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}

	if vol.Config.Ignore {
		return true, "blacklisted", "volume config"
	}

	return false, "", ""
}

func pullImage(c *docker.Client, image string) (err error) {
	if _, _, err = c.ImageInspectWithRaw(context.Background(), image); err != nil {
		// TODO: output pull to logs
		log.WithFields(log.Fields{
			"image": image,
		}).Info("Pulling image")
		resp, err := c.ImagePull(context.Background(), image, types.ImagePullOptions{})
		if err != nil {
			log.Errorf("ImagePull returned an error: %v", err)
			return err
		}
		defer resp.Close()
		body, err := ioutil.ReadAll(resp)
		if err != nil {
			log.Errorf("Failed to read from ImagePull response: %v", err)
			return err
		}
		log.Debugf("Pull image response body: %v", string(body))
	} else {
		log.WithFields(log.Fields{
			"image": image,
		}).Debug("Image already pulled, not pulling")
	}

	return nil
}

func removeContainer(c *docker.Client, id string) {
	log.WithFields(log.Fields{
		"container": id,
	}).Debug("Removing container")
	err := c.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	})
	util.CheckErr(err, "Failed to remove container "+id+": %v", "error")
}

func detectDocker(c *handler.Bivac) bool {
	client, err := docker.NewClient(c.Config.Docker.Endpoint, "", nil, nil)
	if err != nil {
		return false
	}
	_, err = client.Ping(context.Background())
	if err != nil {
		return false
	}
	return true
}
