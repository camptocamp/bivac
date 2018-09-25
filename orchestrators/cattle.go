package orchestrators

import (
	"encoding/base64"
	"encoding/json"
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

	log "github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher/v2"
	"golang.org/x/net/websocket"

	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"
)

// CattleOrchestrator implements a container orchestrator for Cattle
type CattleOrchestrator struct {
	Handler *handler.Bivac
	Client  *client.RancherClient
}

// NewCattleOrchestrator creates a Cattle client
func NewCattleOrchestrator(c *handler.Bivac) (o *CattleOrchestrator) {
	var err error
	o = &CattleOrchestrator{
		Handler: c,
	}

	o.Client, err = client.NewRancherClient(&client.ClientOpts{
		Url:       o.Handler.Config.Cattle.URL,
		AccessKey: o.Handler.Config.Cattle.AccessKey,
		SecretKey: o.Handler.Config.Cattle.SecretKey,
	})
	if err != nil {
		log.Errorf("failed to create a new Rancher client: %s", err)
	}

	return
}

// GetName returns the orchestrator name
func (*CattleOrchestrator) GetName() string {
	return "Cattle"
}

// GetPath returns the path of the backup
func (*CattleOrchestrator) GetPath(v *volume.Volume) string {
	return v.Hostname + "/" + v.Name
}

// GetHandler returns the Orchestrator's handler
func (o *CattleOrchestrator) GetHandler() *handler.Bivac {
	return o.Handler
}

// GetVolumes returns the Cattle volumes
func (o *CattleOrchestrator) GetVolumes() (volumes []*volume.Volume, err error) {
	c := o.Handler

	vs, err := o.Client.Volume.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -2,
			"all":   true,
		},
	})
	if err != nil {
		log.Errorf("failed to list volumes: %s", err)
	}

	var mountpoint string
	for _, v := range vs.Data {
		if len(v.Mounts) < 1 {
			mountpoint = "/data"
		} else {
			mountpoint = v.Mounts[0].Path
		}

		var hostID, hostname string
		var spc *client.StoragePoolCollection
		err := o.rawAPICall("GET", v.Links["storagePools"], &spc)
		if err != nil {
			log.Errorf("failed to retrieve storage pool from volume %s: %s", v.Name, err)
			continue
		}

		if len(spc.Data) == 0 {
			log.Errorf("no storage pool for the volume %s: %s", v.Name, err)
			continue
		}

		if len(spc.Data[0].HostIds) == 0 {
			log.Errorf("no host for the volume %s: %s", v.Name, err)
			continue
		}

		hostID = spc.Data[0].HostIds[0]

		h, err := o.Client.Host.ById(hostID)
		if err != nil {
			log.Errorf("failed to retrieve host from id %s: %s", hostID, err)
			hostname = ""
		} else {
			hostname = h.Hostname
		}

		nv := &volume.Volume{
			Config:     &volume.Config{},
			Mountpoint: mountpoint,
			Name:       v.Name,
			HostBind:   hostID,
			Hostname:   hostname,
		}

		v := volume.NewVolume(nv, c.Config, hostname)
		if b, r, s := o.blacklistedVolume(v); b {
			log.WithFields(log.Fields{
				"volume": v.Name,
				"reason": r,
				"source": s,
			}).Info("Ignoring volume")
			continue
		}
		volumes = append(volumes, v)
	}
	return
}

func createWorkerName() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, 10)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return "bivac-worker-" + string(b)
}

// LaunchContainer starts a containe using the Cattle orchestrator
func (o *CattleOrchestrator) LaunchContainer(image string, env map[string]string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	environment := make(map[string]interface{}, len(env))
	for envKey, envVal := range env {
		environment[envKey] = envVal
	}

	var hostbind string
	if len(volumes) > 0 {
		hostbind = volumes[0].HostBind
	} else {
		hostbind = ""
	}

	cvs := []string{}
	for _, v := range volumes {
		cvs = append(cvs, v.Name+":"+v.Mountpoint)
	}

	container, err := o.Client.Container.Create(&client.Container{
		Name:            createWorkerName(),
		RequestedHostId: hostbind,
		ImageUuid:       "docker:" + image,
		Command:         cmd,
		Environment:     environment,
		RestartPolicy: &client.RestartPolicy{
			MaximumRetryCount: 1,
			Name:              "on-failure",
		},
		DataVolumes: cvs,
	})
	if err != nil {
		log.Errorf("failed to create worker container: %s", err)
		return
	}

	defer o.DeleteWorker(container)

	stopped := false
	terminated := false
	for !terminated {
		container, err := o.Client.Container.ById(container.Id)
		if err != nil {
			log.Errorf("failed to inspect worker: %s", err)
		}

		// This workaround is awful but it's the only way to know if the container failed.
		if container.State == "stopped" {
			if container.StartCount == 1 {
				if stopped == false {
					stopped = true
					time.Sleep(5 * time.Second)
				} else {
					terminated = true
					state = 0
				}
			} else {
				state = 1
				terminated = true
			}
		}
	}

	var hostAccess *client.HostAccess
	err = o.rawAPICall("POST", container.Links["self"]+"/?action=logs", &hostAccess)
	if err != nil {
		log.Errorf("failed to read response from rancher: %s", err)
	}

	origin := o.Handler.Config.Cattle.URL

	u, err := url.Parse(hostAccess.Url)
	if err != nil {
		log.Errorf("failed to parse rancher server url: %s", err)
	}
	q := u.Query()
	q.Set("token", hostAccess.Token)
	u.RawQuery = q.Encode()

	ws, err := websocket.Dial(u.String(), "", origin)
	if err != nil {
		log.Errorf("failed to open websocket with rancher server: %s", err)
	}

	var data = make([]byte, 1024)
	var n int
	if n, err = ws.Read(data); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			log.Errorf("failed to retrieve logs: %s", err)
		}
	}

	re := regexp.MustCompile(`(?m)[0-9]{2,} [ZT\-\:\.0-9]+ (.*)`)
	for _, line := range re.FindAllStringSubmatch(string(data[:n]), -1) {
		stdout = strings.Join([]string{stdout, line[1]}, "\n")
	}

	log.WithFields(log.Fields{
		"container": container.Id,
		"volumes":   strings.Join(cvs[:], ","),
		"cmd":       strings.Join(cmd[:], " "),
	}).Debug(stdout)
	return
}

// DeleteWorker deletes a worker
func (o *CattleOrchestrator) DeleteWorker(container *client.Container) {
	err := o.Client.Container.Delete(container)
	if err != nil {
		log.Errorf("failed to delete worker: %s", err)
	}
	removed := false
	for !removed {
		container, err := o.Client.Container.ById(container.Id)
		if err != nil {
			log.Errorf("failed to inspect worker: %s", err)
		}
		if container.Removed != "" {
			removed = true
		}
	}
	return
}

// GetMountedVolumes returns mounted volumes
func (o *CattleOrchestrator) GetMountedVolumes(v *volume.Volume) (containers []*volume.MountedVolumes, err error) {
	c, err := o.Client.Container.List(&client.ListOpts{
		Filters: map[string]interface{}{
			"limit": -2,
			"all":   true,
		},
	})

	if err != nil {
		log.Errorf("failed to list containers: %s", err)
	}

	for _, container := range c.Data {
		mv := &volume.MountedVolumes{
			ContainerID: container.Id,
			Volumes:     make(map[string]string),
		}
		for _, mount := range container.Mounts {
			mv.Volumes[mount.VolumeName] = mount.Path
		}
		containers = append(containers, mv)
	}
	return
}

// ContainerExec executes a command in a container
func (o *CattleOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolumes, command []string) (stdout string, err error) {

	container, err := o.Client.Container.ById(mountedVolumes.ContainerID)
	if err != nil {
		log.Errorf("failed to retrieve container: %s", err)
		return
	}

	hostAccess, err := o.Client.Container.ActionExecute(container, &client.ContainerExec{
		AttachStdin:  false,
		AttachStdout: true,
		Command:      command,
		Tty:          false,
	})
	if err != nil {
		log.Errorf("failed to prepare command execution in container: %s", err)
		return
	}

	origin := o.Handler.Config.Cattle.URL

	u, err := url.Parse(hostAccess.Url)
	if err != nil {
		log.Errorf("failed to parse rancher server url: %s", err)
	}
	q := u.Query()
	q.Set("token", hostAccess.Token)
	u.RawQuery = q.Encode()

	ws, err := websocket.Dial(u.String(), "", origin)
	if err != nil {
		log.Errorf("failed to open websocket with rancher server: %s", err)
	}

	var data = make([]byte, 1024)
	var n int
	if n, err = ws.Read(data); err != nil {
		if err == io.EOF {
			err = nil
		} else {
			log.Errorf("failed to retrieve logs: %s", err)
		}
	}
	rawStdout, _ := base64.StdEncoding.DecodeString(string(data[:n]))
	stdout = string(rawStdout)

	log.WithFields(log.Fields{
		"container": mountedVolumes.ContainerID,
		"cmd":       strings.Join(command[:], " "),
	}).Debug(stdout)
	return
}

func (o *CattleOrchestrator) blacklistedVolume(vol *volume.Volume) (bool, string, string) {
	if utf8.RuneCountInString(vol.Name) == 64 || utf8.RuneCountInString(vol.Name) == 0 {
		return true, "unnamed", ""
	}

	if strings.Contains(vol.Name, "/") {
		return true, "blacklisted", "path"
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
	sort.Strings(list)
	i := sort.SearchStrings(list, vol.Name)
	if i < len(list) && list[i] == vol.Name {
		return true, "blacklisted", "blacklist config"
	}

	if vol.Config.Ignore {
		return true, "blacklisted", "volume config"
	}

	return false, "", ""
}

func (o *CattleOrchestrator) rawAPICall(method, endpoint string, object interface{}) (err error) {
	// TODO: Use go-rancher.
	// It was impossible to use it, maybe a problem in go-rancher or a lack of documentation.
	clientHTTP := &http.Client{}
	v := url.Values{}
	req, err := http.NewRequest(method, endpoint, strings.NewReader(v.Encode()))
	req.SetBasicAuth(o.Handler.Config.Cattle.AccessKey, o.Handler.Config.Cattle.SecretKey)
	resp, err := clientHTTP.Do(req)
	if err != nil {
		log.Errorf("failed to execute POST request: %s", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response from rancher: %s", err)
	}
	err = json.Unmarshal(body, object)
	if err != nil {
		log.Errorf("failed to unmarshal: %s", err)
	}
	return
}

func detectCattle() bool {
	_, err := net.LookupHost("rancher-metadata")
	if err != nil {
		return false
	}
	return true
}
