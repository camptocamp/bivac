package orchestrators

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/camptocamp/bivac/pkg/volume"
	//"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/go-rancher/v2"
	"golang.org/x/net/websocket"
)

// CattleConfig stores Cattle configuration
type CattleConfig struct {
	URL       string
	AccessKey string
	SecretKey string
}

// CattleOrchestrator implements a container orchestrator for Cattle
type CattleOrchestrator struct {
	config *CattleConfig
	client *client.RancherClient
}

// NewCattleOrchestrator creates a Cattle client
func NewCattleOrchestrator(config *CattleConfig) (o *CattleOrchestrator, err error) {
	o = &CattleOrchestrator{
		config: config,
	}
	o.client, err = client.NewRancherClient(&client.ClientOpts{
		Url:       config.URL,
		AccessKey: config.AccessKey,
		SecretKey: config.SecretKey,
		Timeout:   30 * time.Second,
	})
	if err != nil {
		err = fmt.Errorf("failed to create new Rancher client: %s", err)
	}
	return
}

// GetName returns the orchestrator name
func (*CattleOrchestrator) GetName() string {
	return "cattle"
}

// GetPath returns the backup path
func (*CattleOrchestrator) GetPath(v *volume.Volume) string {
	return v.Hostname
}

// GetVolumes returns the Cattle volumes, inspected and filtered
func (o *CattleOrchestrator) GetVolumes(volumeFilters volume.Filters) (volumes []*volume.Volume, err error) {
	vs, err := o.client.Volume.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -2,
			"all":   true,
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to list volumes: %s", err)
		return
	}

	var mountpoint string
	var h *client.Host
	for _, v := range vs.Data {
		if len(v.Mounts) < 1 {
			mountpoint = "/data"
		} else {
			mountpoint = v.Mounts[0].Path
		}

		var hostID, hostname string
		var spc *client.StoragePoolCollection
		err = o.rawAPICall("GET", v.Links["storagePools"], "", &spc)
		if err != nil {
			continue
		}

		if len(spc.Data) == 0 {
			continue
		}

		if len(spc.Data[0].HostIds) == 0 {
			continue
		}

		hostID = spc.Data[0].HostIds[0]

		h, err = o.client.Host.ById(hostID)
		if err != nil {
			hostname = hostID
		} else {
			hostname = h.Hostname
		}

		v := &volume.Volume{
			ID:         v.Id,
			Name:       v.Name,
			Mountpoint: mountpoint,
			HostBind:   hostID,
			Hostname:   hostname,
		}

		if b, _, _ := o.blacklistedVolume(v, volumeFilters); b {
			continue
		}
		volumes = append(volumes, v)
	}
	return
}

func createAgentName() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return "bivac-agent-" + string(b)
}

// DeployAgent creates a `bivac agent` container
func (o *CattleOrchestrator) DeployAgent(image string, cmd []string, envs []string, v *volume.Volume) (success bool, output string, err error) {
	success = false

	environment := make(map[string]interface{})
	for _, env := range envs {
		splitted := strings.Split(env, "=")
		environment[splitted[0]] = splitted[1]
	}
	container, err := o.client.Container.Create(&client.Container{
		Name:            createAgentName(),
		RequestedHostId: v.HostBind,
		ImageUuid:       "docker:" + image,
		Command:         cmd,
		Environment:     environment,
		RestartPolicy: &client.RestartPolicy{
			MaximumRetryCount: 1,
			Name:              "on-failure",
		},
		Labels: map[string]interface{}{
			"io.rancher.container.pull_image": "always",
		},
		DataVolumes: []string{
			v.Name + ":" + v.Mountpoint,
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to create an agent container: %s", err)
		return
	}
	defer o.RemoveContainer(container)

	stopped := false
	terminated := false
	timeout := time.After(60 * time.Second)
	for !terminated {
		container, err := o.client.Container.ById(container.Id)
		if err != nil {
			err = fmt.Errorf("failed to inspect agent: %s", err)
			return false, "", err
		}

		// This workaround is awful but it's the only way to know if the container failed.
		if container.State == "stopped" {
			if container.StartCount == 1 {
				if stopped == false {
					stopped = true
					time.Sleep(5 * time.Second)
				} else {
					terminated = true
					success = true
				}
			} else {
				success = false
				terminated = true
			}
		} else if container.State == "starting" {
			select {
			case <-timeout:
				err = fmt.Errorf("failed to start agent: timeout")
				return false, "", err
			default:
				continue
			}
		} else if container.State == "error" {
			terminated = true
			success = false
		}
		time.Sleep(1 * time.Second)
	}

	container, err = o.client.Container.ById(container.Id)
	if err != nil {
		err = fmt.Errorf("failed to inspect the agent before retrieving the logs: %s", err)
		return false, "", err
	}

	var hostAccess *client.HostAccess
	logsParams := `{"follow":false,"lines":9999,"since":"","timestamps":true}`
	err = o.rawAPICall("POST", container.Links["self"]+"/?action=logs", logsParams, &hostAccess)
	if err != nil {
		err = fmt.Errorf("failed to read response from rancher: %s", err)
		return
	}

	origin := o.config.URL

	u, err := url.Parse(hostAccess.Url)
	if err != nil {
		err = fmt.Errorf("failed to parse rancher server url: %s", err)
		return
	}

	q := u.Query()
	q.Set("token", hostAccess.Token)
	u.RawQuery = q.Encode()

	ws, err := websocket.Dial(u.String(), "", origin)
	if err != nil {
		err = fmt.Errorf("failed to open websocket with rancher server: %s", err)
		return
	}
	defer ws.Close()

	var data bytes.Buffer
	io.Copy(&data, ws)

	re := regexp.MustCompile(`(?m)[0-9]{2,} [ZT\-\:\.0-9]+ (.*)`)
	for _, line := range re.FindAllStringSubmatch(data.String(), -1) {
		output = strings.Join([]string{output, line[1]}, "\n")
	}
	return
}

// RemoveContainer remove container based on its ID
func (o *CattleOrchestrator) RemoveContainer(container *client.Container) {
	err := o.client.Container.Delete(container)
	if err != nil {
		err = fmt.Errorf("failed to remove container: %s", err)
		return
	}
	removed := false
	for !removed {
		container, err := o.client.Container.ById(container.Id)
		if err != nil {
			err = fmt.Errorf("failed to inspect container: %s", err)
			return
		}
		if container.Removed != "" {
			removed = true
		}
	}
	return
}

// GetContainersMountingVolume returns containers mounting a volume
func (o *CattleOrchestrator) GetContainersMountingVolume(v *volume.Volume) (mountedVolumes []*volume.MountedVolume, err error) {
	vol, err := o.client.Volume.ById(v.ID)
	if err != nil {
		err = fmt.Errorf("failed to get volume: %s", err)
		return
	}

	for _, mount := range vol.Mounts {
		instance, err := o.client.Container.ById(mount.InstanceId)
		if err != nil {
			continue
		}

		if instance.State != "running" {
			continue
		}

		mv := &volume.MountedVolume{
			ContainerID: mount.InstanceId,
			Volume:      v,
			Path:        mount.Path,
		}
		mountedVolumes = append(mountedVolumes, mv)
	}
	return
}

// ContainerExec executes a command in a container
func (o *CattleOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolume, command []string) (stdout string, err error) {
	container, err := o.client.Container.ById(mountedVolumes.ContainerID)
	if err != nil {
		err = fmt.Errorf("failed to retrieve container: %s", err)
		return
	}

	hostAccess, err := o.client.Container.ActionExecute(container, &client.ContainerExec{
		AttachStdin:  false,
		AttachStdout: true,
		Command:      command,
		Tty:          false,
	})
	if err != nil {
		err = fmt.Errorf("failed to prepare command execution in containers: %s", err)
		return
	}

	origin := o.config.URL

	u, err := url.Parse(hostAccess.Url)
	if err != nil {
		err = fmt.Errorf("failed to parse rancher server url: %s", err)
		return
	}
	q := u.Query()
	q.Set("token", hostAccess.Token)
	u.RawQuery = q.Encode()

	ws, err := websocket.Dial(u.String(), "", origin)
	if err != nil {
		err = fmt.Errorf("failed to open websocket with rancher server: %s", err)
		return
	}

	var data bytes.Buffer
	io.Copy(&data, ws)

	rawStdout, _ := base64.StdEncoding.DecodeString(data.String())
	stdout = string(rawStdout)

	return
}

func (o *CattleOrchestrator) blacklistedVolume(vol *volume.Volume, volumeFilters volume.Filters) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || utf8.RuneCountInString(vol.Name) == 0 {
		return true, "unnamed", ""
	}

	if strings.Contains(vol.Name, "/") {
		return true, "unnamed", "path"
	}
	// Use whitelist if defined
	if l := volumeFilters.Whitelist; len(l) > 0 && l[0] != "" {
		sort.Strings(l)
		i := sort.SearchStrings(l, vol.Name)
		if i < len(l) && l[i] == vol.Name {
			return false, "", ""
		}
		return true, "blacklisted", "whitelist config"
	}

	i := sort.SearchStrings(volumeFilters.Blacklist, vol.Name)
	if i < len(volumeFilters.Blacklist) && volumeFilters.Blacklist[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}
	return false, "", ""
}

func (o *CattleOrchestrator) rawAPICall(method, endpoint string, data string, object interface{}) (err error) {
	// TODO: Use go-rancher.
	// It was impossible to use it, maybe a problem in go-rancher or a lack of documentation.
	clientHTTP := &http.Client{}
	//v := url.Values{}
	req, err := http.NewRequest(method, endpoint, strings.NewReader(data))
	req.SetBasicAuth(o.config.AccessKey, o.config.SecretKey)
	resp, err := clientHTTP.Do(req)
	defer resp.Body.Close()
	if err != nil {
		err = fmt.Errorf("failed to execute POST request: %s", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("failed to read response from rancher: %s", err)
		return
	}
	err = json.Unmarshal(body, object)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal: %s", err)
		return
	}
	return
}

func DetectCattle() bool {
	_, err := net.LookupHost("rancher-metadata")
	if err != nil {
		return false
	}
	return true
}
