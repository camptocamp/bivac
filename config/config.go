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
	LabelPrefix      string            `long:"label-prefix" description:"The volume prefix label." env:"BIVAC_LABEL_PREFIX"`
	ExtraEnv         map[string]string `long:"extra-env" description:"Extra environment variables to share with workers." env:"BIVAC_EXTRA_ENV"`
	ProvidersFile    string            `short:"p" long:"providers-file" description:"Path to providers configuration file." env:"BIVAC_PROVIDERS_FILE" default:"/providers-config.default.toml"`

	Restic struct {
		CommonArgs []string `long:"restic-args" description:"Arguments to pass to restic engine." env:"RESTIC_COMMON_ARGS" default:"-r %B/%P/%V"`
		BackupArgs []string `long:"restic-backup-args" description:"Arguments to pass to restic engine when backup." env:"RESTIC_BACKUP_ARGS" default:"%D --hostname %H"`
		ForgetArgs []string `long:"restic-forget-args" description:"Arguments to pass to restic engine when launching forget." env:"RESTIC_FORGET_ARGS" default:"--keep-daily 15 --prune"`
		Image      string   `long:"restic-image" description:"The restic docker image." env:"RESTIC_DOCKER_IMAGE" default:"restic/restic:latest"`
		Password   string   `long:"restic-password" description:"The restic backup password." env:"RESTIC_PASSWORD"`
	} `group:"Restic Options"`

	RClone struct {
		CommonArgs []string `long:"rclone-args" description:"Arguments to pass to rclone engine." env:"RCLONE_COMMON_ARGS"`
		BackupArgs []string `long:"rclone-backup-args" description:"Arguments to pass to rclone engine when backup." env:"RCLONE_BACKUP_ARGS" default:"%D %B/%P/%V"`
		Image      string   `long:"rclone-image" description:"The rclone docker image." env:"RCLONE_DOCKER_IMAGE" default:"camptocamp/rclone:1.42-1"`
	} `group:"RClone Options"`

	Duplicity struct {
		CommonArgs          []string `long:"duplicity-args" description:"Arguments to pass to duplicity engine." env:"DUPLICITY_COMMON_ARGS" default:"--s3-use-new-style --ssh-options -oStrictHostKeyChecking=no --no-encryption"`
		BackupArgs          []string `long:"duplicity-backup-args" description:"Arguments to pass to duplicity engine when backup." env:"DUPLICITY_BACKUP_ARGS" default:"--full-if-older-than 15D --allow-source-mismatch --name %V %D %B/%P/%V"`
		RemoveOlderThanArgs []string `long:"duplicity-remove-older-than-args" description:"Arguments to pass to duplicity engine when removing old backups." env:"DUPLICITY_REMOVE_OLDER_THAN_ARGS" default:"30D --force --name %V %B/%P/%V"`
		Image               string   `long:"duplicity-image" description:"The duplicity docker image." env:"DUPLICITY_DOCKER_IMAGE" default:"camptocamp/duplicity:latest"`
	} `group:"Duplicity Options"`

	Metrics struct {
		PushgatewayURL string `short:"g" long:"gateway-url" description:"The prometheus push gateway URL to use." env:"PUSHGATEWAY_URL"`
	} `group:"Metrics Options"`

	AWS struct {
		AccessKeyID     string `long:"aws-access-key-id" description:"The AWS access key ID." env:"AWS_ACCESS_KEY_ID"`
		SecretAccessKey string `long:"aws-secret-key-id" description:"The AWS secret access key." env:"AWS_SECRET_ACCESS_KEY"`
	} `group:"AWS Options"`

	Swift struct {
		Username          string `long:"swift-username" description:"The Swift user name." env:"SWIFT_USERNAME"`
		Password          string `long:"swift-password" description:"The Swift password." env:"SWIFT_PASSWORD"`
		AuthURL           string `long:"swift-auth_url" description:"The Swift auth URL." env:"SWIFT_AUTHURL"`
		TenantName        string `long:"swift-tenant-name" description:"The Swift tenant name." env:"SWIFT_TENANTNAME"`
		RegionName        string `long:"swift-region-name" description:"The Swift region name." env:"SWIFT_REGIONNAME"`
		UserDomainName    string `long:"swift-user-domain-name" description:"The Swift user domain name." env:"SWIFT_USER_DOMAIN_NAME"`
		ProjectName       string `long:"swift-project-name" description:"The Swift project name." env:"SWIFT_PROJECT_NAME"`
		ProjectDomainName string `long:"swift-project-domain-name" description:"The Swift project domain name." env:"SWIFT_PROJECT_DOMAIN_NAME"`
	} `group:"Swift Options"`

	Docker struct {
		Endpoint string `short:"e" long:"docker-endpoint" description:"The Docker endpoint." env:"DOCKER_ENDPOINT" default:"unix:///var/run/docker.sock"`
	} `group:"Docker Options"`

	Kubernetes struct {
		Namespace            string `long:"k8s-namespace" description:"Namespace where you want to run Bivac." env:"K8S_NAMESPACE"`
		AllNamespaces        bool   `long:"k8s-all-namespaces" description:"Backup volumes of all namespaces." env:"K8S_ALL_NAMESPACES"`
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
