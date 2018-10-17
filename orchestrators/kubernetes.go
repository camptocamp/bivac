package orchestrators

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"time"
	"unicode/utf8"

	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/handler"
	"github.com/camptocamp/bivac/volume"
	apiv1 "k8s.io/api/core/v1"
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

// GetPath returns the path of the backup
func (*KubernetesOrchestrator) GetPath(v *volume.Volume) string {
	return v.Namespace
}

// GetHandler returns the Orchestrator's handler
func (o *KubernetesOrchestrator) GetHandler() *handler.Bivac {
	return o.Handler
}

// GetVolumes returns the Kubernetes persistent volume claims, inspected and filtered
func (o *KubernetesOrchestrator) GetVolumes() (volumes []*volume.Volume, err error) {
	c := o.Handler

	// Get namespaces
	namespaces, err := o.getNamespaces()

	for _, namespace := range namespaces {
		o.setNamespace(namespace)
		pvcs, err := o.Client.CoreV1().PersistentVolumeClaims(o.Handler.Config.Kubernetes.Namespace).List(metav1.ListOptions{})
		if err != nil {
			log.Errorf("failed to retrieve the list of PVCs: %v", err)
		}

		for _, pvc := range pvcs.Items {
			nv := &volume.Volume{
				Config:     &volume.Config{},
				Mountpoint: "/data",
				Name:       pvc.Name,
				Namespace:  namespace,
			}

			containers, _ := o.GetContainersMountingVolume(nv)
			if len(containers) > 0 {
				nv.HostBind = containers[0].HostID
				nv.Hostname = containers[0].HostID
				nv.Mountpoint = containers[0].Path
			}

			v := volume.NewVolume(nv, c.Config, namespace)
			if b, r, s := o.blacklistedVolume(v); b {
				log.WithFields(log.Fields{
					"volume": pvc.Name,
					"reason": r,
					"source": s,
				}).Info("Ignoring volume")
				continue
			}
			volumes = append(volumes, v)
			log.Infof("Detected volume: %s.%s", namespace, v.Name)
		}
	}
	return
}

// LaunchContainer starts a container using the Kubernetes orchestrator
func (o *KubernetesOrchestrator) LaunchContainer(image string, cmd []string, volumes []*volume.Volume) (state int, stdout string, err error) {
	kvs := []apiv1.Volume{}
	kvms := []apiv1.VolumeMount{}
	var node string

	for _, v := range volumes {
		o.setNamespace(v.Namespace)
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

	managerHostname, err := os.Hostname()
	if err != nil {
		log.Errorf("failed to get hostname: %s", err)
		return
	}

	// get the namespace
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	namespace, _, err := kubeconfig.Namespace()
	if err != nil {
		log.Errorf("Failed to get namespace: %v", err)
		return
	}

	managerPod, err := o.Client.CoreV1().Pods(namespace).Get(managerHostname, metav1.GetOptions{})
	if err != nil {
		log.Errorf("failed to get current pod: %s", err)
		return
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
					Env:             managerPod.Spec.Containers[0].Env,
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
			if len(pod.Status.ContainerStatuses) == 0 {
				return 0, "", fmt.Errorf("no container statuses found")
			}
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

	log.Debug(stdout)
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

// GetContainersMountingVolume returns containers mounting a volume
func (o *KubernetesOrchestrator) GetContainersMountingVolume(v *volume.Volume) (containers []*volume.MountedVolume, err error) {
	o.setNamespace(v.Namespace)

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
			for _, volumeMount := range container.VolumeMounts {
				if c, ok := mapVolClaim[volumeMount.Name]; ok {
					if c == v.Name {
						mv := &volume.MountedVolume{
							PodID:       pod.Name,
							ContainerID: container.Name,
							HostID:      pod.Spec.NodeName,
							Volume:      v,
							Path:        volumeMount.MountPath,
						}
						containers = append(containers, mv)
					}
				}
			}
		}
	}

	return
}

// ContainerExec executes a command in a container
func (o *KubernetesOrchestrator) ContainerExec(mountedVolumes *volume.MountedVolume, command []string) (stdout string, err error) {
	var stdoutput, stderr bytes.Buffer

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
		Stdout: &stdoutput,
		Stderr: &stderr,
		Tty:    false,
	})

	stdout = stdoutput.String()
	return
}

func (o *KubernetesOrchestrator) setNamespace(namespace string) {
	o.Handler.Config.Kubernetes.Namespace = namespace
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

func (o *KubernetesOrchestrator) getNamespaces() (namespaces []string, err error) {
	if o.Handler.Config.Kubernetes.AllNamespaces == true {
		nms, err := o.Client.CoreV1().Namespaces().List(metav1.ListOptions{})
		if err != nil {
			err = fmt.Errorf("failed to retrieve the list of namespaces: %v", err)
			return []string{}, err
		}
		for _, namespace := range nms.Items {
			namespaces = append(namespaces, namespace.Name)
		}
	} else {
		namespaces = append(namespaces, o.Handler.Config.Kubernetes.Namespace)
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
