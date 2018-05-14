package orchestrators

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"

	log "github.com/Sirupsen/logrus"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

// KubernetesOrchestrator implements a container orchestrator for Kubernetes
type KubernetesOrchestrator struct {
	Handler *handler.Bivac
	Client  *kubernetes.Clientset
}

// NewKubernetesOrchestrator creates a Kubernetes client
func NewKubernetesOrchestrator(c *handler.Bivac) (o *KubernetesOrchestrator) {
	var err error
	o = &KubernetesOrchestrator{
		Handler: c,
	}

	config, err := o.getConfig()
	if err != nil {
		log.Fatalf("failed to retrieve Kubernetes config: %s", err)
	}

	o.Client, err = kubernetes.NewForConfig(config)
	if err != nil {
		log.Errorf("failed to create a Kubernetes client: %v", err)
	}
	return
}

// GetName returns the orchestrator name
func (*KubernetesOrchestrator) GetName() string {
	return "Kubernetes"
}

// GetHandler returns the Orchestrator's handler
func (o *KubernetesOrchestrator) GetHandler() *handler.Bivac {
	return o.Handler
}

// GetVolumes returns the Kubernetes persistent volume claims, inspected and filtered
func (o *KubernetesOrchestrator) GetVolumes() (volumes []*volume.Volume, err error) {
	c := o.Handler

	pvcs, err := o.Client.CoreV1().PersistentVolumeClaims(o.Handler.Config.Kubernetes.Namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("failed to retrieve the list of PVCs: %v", err)
	}

	containers, err := o.GetMountedVolumes()
	mountedVolumes := make(map[string]string)
	bindHostVolume := make(map[string]string)
	for _, container := range containers {
		for volName, volMountpath := range container.Volumes {
			mountedVolumes[volName] = volMountpath
			bindHostVolume[volName] = container.HostID
		}
	}
	var mountpoint string
	for _, pvc := range pvcs.Items {
		if value, ok := mountedVolumes[pvc.Name]; ok {
			mountpoint = value
		} else {
			mountpoint = "/data"
		}
		nv := &volume.Volume{
			Config:     &volume.Config{},
			Mountpoint: mountpoint,
			Name:       pvc.Name,
			HostBind:   bindHostVolume[pvc.Name],
			Hostname:   bindHostVolume[pvc.Name],
		}

		v := volume.NewVolume(nv, c.Config, c.Hostname)
		if b, r, s := o.blacklistedVolume(v); b {
			log.WithFields(log.Fields{
				"volume": pvc.Name,
				"reason": r,
				"source": s,
			}).Info("Ignoring volume")
			continue
		}
		volumes = append(volumes, v)
		log.Infof("%+v", v)
	}
	return
}

// LaunchContainer starts a container using the Kubernetes orchestrator
func (o *KubernetesOrchestrator) LaunchContainer(image string, env map[string]string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {

	var envVars []apiv1.EnvVar
	for envName, envValue := range env {
		ev := apiv1.EnvVar{
			Name:  envName,
			Value: envValue,
		}
		envVars = append(envVars, ev)
	}

	kvs := []apiv1.Volume{}
	kvms := []apiv1.VolumeMount{}
	var node string

	for _, v := range volumes {
		pvc, err := o.Client.CoreV1().PersistentVolumeClaims(o.Handler.Config.Kubernetes.Namespace).Get(v.Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("failed to retrieve PersistentVolumeClaim \""+v.Name+"\": %s", err)
			continue
		}

		for _, am := range pvc.Spec.AccessModes {
			if am == apiv1.ReadWriteOnce {
				node = v.HostBind
			}
		}

		kv := apiv1.Volume{
			Name: v.Name,
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: v.Name,
					ReadOnly:  false,
				},
			},
		}
		kvs = append(kvs, kv)

		kvm := apiv1.VolumeMount{
			Name:      v.Name,
			ReadOnly:  v.ReadOnly,
			MountPath: v.Mountpoint,
		}
		kvms = append(kvms, kvm)
	}

	pod, err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "bivac-worker-",
		},
		Spec: apiv1.PodSpec{
			NodeName:           node,
			RestartPolicy:      "Never",
			Volumes:            kvs,
			ServiceAccountName: o.Handler.Config.Kubernetes.WorkerServiceAccount,
			Containers: []apiv1.Container{
				{
					Name:            "bivac-worker",
					Image:           image,
					Args:            cmd,
					Env:             envVars,
					VolumeMounts:    kvms,
					ImagePullPolicy: apiv1.PullAlways,
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to create worker: %s", err)
	}

	workerName := pod.ObjectMeta.Name

	defer o.DeleteWorker(workerName)

	timeout := time.After(60 * time.Second)
	terminated := false
	for !terminated {
		pod, err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).Get(workerName, metav1.GetOptions{})
		if err != nil {
			log.Errorf("failed to get pod: %s", err)
		}

		if pod.Status.Phase == apiv1.PodSucceeded || pod.Status.Phase == apiv1.PodFailed {
			state = int(pod.Status.ContainerStatuses[0].State.Terminated.ExitCode)
			terminated = true
		} else if pod.Status.Phase != apiv1.PodRunning {
			select {
			case <-timeout:
				err = fmt.Errorf("failed to start worker: timeout")
				return -1, "", err
			default:
				continue
			}
		}
	}

	req := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).GetLogs(workerName, &apiv1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		log.Errorf("failed to read logs: %s", err)
	}

	defer readCloser.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(readCloser)
	stdout = buf.String()

	return
}

// DeleteWorker deletes a worker
func (o *KubernetesOrchestrator) DeleteWorker(name string) {
	err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		log.Errorf("failed to delete worker: %s", err)
	}
	return
}

// GetMountedVolumes returns mounted volumes
func (o *KubernetesOrchestrator) GetMountedVolumes() (containers []*volume.MountedVolumes, err error) {

	pods, err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Errorf("failed to get pods: %s", err)
	}

	mapVolClaim := make(map[string]string)

	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				mapVolClaim[volume.Name] = volume.PersistentVolumeClaim.ClaimName
			}
		}

		for _, container := range pod.Spec.Containers {
			mv := &volume.MountedVolumes{
				PodID:       pod.Name,
				ContainerID: container.Name,
				HostID:      pod.Spec.NodeName,
				Volumes:     make(map[string]string),
			}
			for _, volumeMount := range container.VolumeMounts {
				if c, ok := mapVolClaim[volumeMount.Name]; ok {
					mv.Volumes[c] = volumeMount.MountPath
				}
			}
			containers = append(containers, mv)
		}
	}

	return
}

// ContainerExec executes a command in a container
func (o *KubernetesOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolumes, command []string) (err error) {
	var stdout, stderr bytes.Buffer

	config, err := o.getConfig()
	if err != nil {
		log.Fatalf("failed to retrieve Kubernetes config: %s", err)
	}

	req := o.Client.Core().RESTClient().Post().
		Resource("pods").
		Name(mountedVolumes.PodID).
		Namespace(o.Handler.Config.Kubernetes.Namespace).
		SubResource("exec").
		Param("container", mountedVolumes.ContainerID)
	req.VersionedParams(&apiv1.PodExecOptions{
		Container: mountedVolumes.ContainerID,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Errorf("failed to call the API: %s", err)
		return err
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
		Tty:    false,
	})

	return
}

// ContainerPrepareBackup executes a command in a container
func (o *KubernetesOrchestrator) ContainerPrepareBackup(mountedVolumes *volume.MountedVolumes, command []string) (backupVolume *volume.Volume, err error) {
	pr, pw := io.Pipe()
	go func() {
		var stderr bytes.Buffer

		config, err := o.getConfig()
		if err != nil {
			log.Fatalf("failed to retrieve Kubernetes config: %s", err)
		}

		req := o.Client.Core().RESTClient().Post().
			Resource("pods").
			Name(mountedVolumes.PodID).
			Namespace(o.Handler.Config.Kubernetes.Namespace).
			SubResource("exec").
			Param("container", mountedVolumes.ContainerID)
		req.VersionedParams(&apiv1.PodExecOptions{
			Container: mountedVolumes.ContainerID,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

		exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
		if err != nil {
			log.Errorf("failed to call the API: %s", err)
			return
		}
		err = exec.Stream(remotecommand.StreamOptions{
			Stdin:  nil,
			Stdout: pw,
			Stderr: &stderr,
			Tty:    false,
		})
		defer pw.Close()
		if stderr.Len() > 0 {
			log.Warningf("STDERR of the prepare backup command: %s", stderr.String())
		}
		return
	}()

	_, err = o.Client.CoreV1().PersistentVolumeClaims(o.Handler.Config.Kubernetes.Namespace).Create(&apiv1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bivac-tmp",
		},
		Spec: apiv1.PersistentVolumeClaimSpec{
			AccessModes: []apiv1.PersistentVolumeAccessMode{
				apiv1.ReadWriteOnce,
				apiv1.ReadWriteMany,
			},
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceStorage: resource.MustParse("100Gi"),
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to create temporary persistent volume: %s", err)
	}
	tmpVol := []apiv1.Volume{
		{
			Name: "bivac-tmp",
			VolumeSource: apiv1.VolumeSource{
				PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
					ClaimName: "bivac-tmp",
					ReadOnly:  false,
				},
			},
		},
	}

	tmpMntVol := []apiv1.VolumeMount{
		{
			Name:      "bivac-tmp",
			ReadOnly:  false,
			MountPath: "/data",
		},
	}
	pod, err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).Create(&apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "bivac-worker-",
		},
		Spec: apiv1.PodSpec{
			RestartPolicy: "Never",
			Volumes:       tmpVol,
			Containers: []apiv1.Container{
				{
					Name:  "bivac-worker",
					Image: "busybox",
					Args: []string{
						"sleep",
						"100000",
					},
					VolumeMounts: tmpMntVol,
				},
			},
		},
	})
	if err != nil {
		log.Errorf("failed to create worker: %s", err)
	}
	workerName := pod.ObjectMeta.Name
	defer o.DeleteWorker(workerName)

	running := false
	for !running {
		pod, err := o.Client.CoreV1().Pods(o.Handler.Config.Kubernetes.Namespace).Get(workerName, metav1.GetOptions{})
		if err != nil {
			log.Errorf("failed to get pod: %s", err)
		}

		if pod.Status.Phase == apiv1.PodRunning {
			running = true
		}
	}

	config, err := o.getConfig()
	if err != nil {
		log.Fatalf("failed to retrieve Kubernetes config: %s", err)
	}
	req := o.Client.Core().RESTClient().Post().
		Resource("pods").
		Name(workerName).
		Namespace(o.Handler.Config.Kubernetes.Namespace).
		SubResource("exec").
		Param("container", "bivac-worker")
	req.VersionedParams(&apiv1.PodExecOptions{
		Container: "bivac-worker",
		Command: []string{
			"/bin/sh",
			"-c",
			"cat > /data/backup",
		},
		Stdin:  true,
		Stdout: false,
		Stderr: false,
		TTY:    false,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		log.Errorf("failed to call the API: %s", err)
		return
	}
	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: ioutil.Discard,
		Stderr: ioutil.Discard,
		Stdin:  pr,
		Tty:    false,
	})
	defer pr.Close()
	return
}

func (o *KubernetesOrchestrator) blacklistedVolume(vol *volume.Volume) (bool, string, string) {

	defaultBlacklistedVolumes := []string{
		"duplicity_cache",
		"restic_cache",
		"duplicity-cache",
		"restic-cache",
		"lost+found",
	}

	if utf8.RuneCountInString(vol.Name) == 64 {
		return true, "unnamed", ""
	}

	list := o.Handler.Config.VolumesBlacklist
	list = append(list, defaultBlacklistedVolumes...)
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

func (o *KubernetesOrchestrator) getConfig() (config *rest.Config, err error) {
	if o.Handler.Config.Kubernetes.KubeConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", o.Handler.Config.Kubernetes.KubeConfig)
	} else {
		kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(),
			&clientcmd.ConfigOverrides{},
		)

		if o.Handler.Config.Kubernetes.Namespace != "" {
			log.Warningf("Using provided Kubernetes namespace.")
		} else {
			o.Handler.Config.Kubernetes.Namespace, _, err = kubeconfig.Namespace()
		}

		if err != nil {
			log.Errorf("Failed to retrieve the namespace from the cluster config: %v", err)
		}
		config, err = rest.InClusterConfig()
	}
	return
}

func detectKubernetes() bool {
	_, err := rest.InClusterConfig()
	if err != nil {
		return false
	}
	return true
}
