package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/jessevdk/go-flags"
)

// Config stores the handler's configuration and UI interface parameters
type Config struct {
	Version          bool              `short:"V" long:"version" description:"Display version."`
	Loglevel         string            `short:"l" long:"loglevel" description:"Set loglevel ('debug', 'info', 'warn', 'error', 'fatal', 'panic')." env:"BIVAC_LOG_LEVEL" default:"info"`
	VolumesBlacklist []string          `short:"b" long:"blacklist" description:"Volumes to blacklist in backups." env:"BIVAC_VOLUMES_BLACKLIST" env-delim:","`
	VolumesWhitelist []string          `short:"w" long:"whitelist" description:"Only backup whitelisted volumes." env:"BIVAC_VOLUMES_WHITELIST" env-delim:","`
	Manpage          bool              `short:"m" long:"manpage" description:"Output manpage."`
	NoVerify         bool              `long:"no-verify" description:"Do not verify backup." env:"BIVAC_NO_VERIFY"`
	JSON             bool              `short:"j" long:"json" description:"Log as JSON (to stderr)." env:"BIVAC_JSON_OUTPUT"`
	Engine           string            `short:"E" long:"engine" description:"Backup engine to use." env:"BIVAC_ENGINE" default:"restic"`
	Orchestrator     string            `short:"o" long:"orchestrator" description:"Container orchestrator to use." env:"BIVAC_ORCHESTRATOR"`
	TargetURL        string            `short:"u" long:"target-url" description:"The target URL to push to." env:"BIVAC_TARGET_URL"`
	CheckEvery       string            `long:"check-every" description:"Time between backup checks." env:"BIVAC_CHECK_EVERY" default:"24h"`
	RemoveOlderThan  string            `long:"remove-older-than" description:"Remove backups older than the specified interval." env:"BIVAC_REMOVE_OLDER_THAN" default:"30D"`
	LabelPrefix      string            `long:"label-prefix" description:"The volume prefix label." env:"BIVAC_LABEL_PREFIX"`
	ExtraEnv         map[string]string `long:"extra-env" description:"Extra environment variables to share with workers." env:"BIVAC_EXTRA_ENV"`

	Restic struct {
		Image    string `long:"restic-image" description:"The restic docker image." env:"RESTIC_DOCKER_IMAGE" default:"restic/restic:latest"`
		Password string `long:"restic-password" description:"The restic backup password." env:"RESTIC_PASSWORD"`
	} `group:"Restic Options"`

	RClone struct {
		Image string `long:"rclone-image" description:"The rclone docker image." env:"RCLONE_DOCKER_IMAGE" default:"camptocamp/rclone:1.33-1"`
	} `group:"RClone Options"`

	Duplicity struct {
		Image           string `long:"duplicity-image" description:"The duplicity docker image." env:"DUPLICITY_DOCKER_IMAGE" default:"camptocamp/duplicity:latest"`
		FullIfOlderThan string `long:"full-if-older-than" description:"The number of days after which a full backup must be performed." env:"BIVAC_FULL_IF_OLDER_THAN" default:"15D"`
	} `group:"Duplicity Options"`

	Metrics struct {
		PushgatewayURL string `short:"g" long:"gateway-url" description:"The prometheus push gateway URL to use." env:"PUSHGATEWAY_URL"`
	} `group:"Metrics Options"`

	AWS struct {
		AccessKeyID     string `long:"aws-access-key-id" description:"The AWS access key ID." env:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `long:"aws-secret-key-id" description:"The AWS secret access key." env:"AWS_SECRET_ACCESS_KEY"`
	} `group:"AWS Options"`

	Swift struct {
		Username   string `long:"swift-username" description:"The Swift user name." env:"SWIFT_USERNAME"`
		Password   string `long:"swift-password" description:"The Swift password." env:"SWIFT_PASSWORD"`
		AuthURL    string `long:"swift-auth_url" description:"The Swift auth URL." env:"SWIFT_AUTHURL"`
		TenantName string `long:"swift-tenant-name" description:"The Swift tenant name." env:"SWIFT_TENANTNAME"`
		RegionName string `long:"swift-region-name" description:"The Swift region name." env:"SWIFT_REGIONNAME"`
	} `group:"Swift Options"`

	Docker struct {
		Endpoint string `short:"e" long:"docker-endpoint" description:"The Docker endpoint." env:"DOCKER_ENDPOINT" default:"unix:///var/run/docker.sock"`
	} `group:"Docker Options"`

	Kubernetes struct {
		Namespace            string `long:"k8s-namespace" description:"Namespace where you want to run Bivac." env:"K8S_NAMESPACE"`
		KubeConfig           string `long:"k8s-kubeconfig" description:"Path to your kubeconfig file." env:"K8S_KUBECONFIG"`
		WorkerServiceAccount string `long:"k8s-worker-service-account" description:"Specify service account for workers." env:"K8S_WORKER_SERVICE_ACCOUNT"`
	} `group:"Kubernetes Options"`

	Cattle struct {
		Environment string `long:"cattle-env" description:"The Cattle environment." env:"CATTLE_ENV"`
		AccessKey   string `long:"cattle-accesskey" description:"The Cattle access key." env:"CATTLE_ACCESS_KEY"`
		SecretKey   string `long:"cattle-secretkey" description:"The Cattle secretkey." env:"CATTLE_SECRET_KEY"`
		URL         string `long:"cattle-url" description:"The Cattle url." env:"CATTLE_URL"`
	} `group:"Cattle Options"`
}

// LoadConfig loads the config from flags & environment
func LoadConfig(version string) *Config {
	var c Config
	parser := flags.NewParser(&c, flags.Default)
	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

	if c.Version {
		fmt.Printf("Bivac %v\n", version)
		os.Exit(0)
	}

	if c.Manpage {
		var buf bytes.Buffer
		parser.ShortDescription = "Docker volumes backup"
		parser.LongDescription = `Bivac lets you backup all your names Docker volumes using Duplicity or RClone.

Bivac supports multiple engines for performing the backup:

* Restic (default engine)

* RClone: use for heavy data that Restic/Duplicity cannot manage efficiently

* Duplicity
`
		parser.WriteManPage(&buf)
		fmt.Printf(buf.String())
		os.Exit(0)
	}

	sort.Strings(c.VolumesBlacklist)
	return &c
}
