package manager

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/manager"
	"github.com/camptocamp/bivac/pkg/volume"
)

var (
	// Server stores informations relative the Bivac server
	server manager.Server

	// Orchestrator is the name of the orchestrator on which Bivac should connect to
	Orchestrator string

	Orchestrators manager.Orchestrators

	dbPath           string
	resticForgetArgs string

	providersFile string
	targetURL     string
	retryCount    int
)
var envs = make(map[string]string)

// TODO: Rename this command to something more explicit
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start Bivac backup manager",
	Run: func(cmd *cobra.Command, args []string) {
		// Global variables
		whitelistVolumes, _ := cmd.Flags().GetString("whitelist")
		blacklistVolumes, _ := cmd.Flags().GetString("blacklist")

		volumesFilters := volume.Filters{
			Whitelist: strings.Split(whitelistVolumes, ","),
			Blacklist: strings.Split(blacklistVolumes, ","),
		}

		o, err := manager.GetOrchestrator(Orchestrator, Orchestrators)
		if err != nil {
			log.Errorf("failed to retrieve orchestrator: %s", err)
			return
		}

		err = manager.Start(o, server, volumesFilters, providersFile, targetURL, retryCount)
		if err != nil {
			log.Errorf("failed to start manager: %s", err)
			return
		}
	},
}

func init() {
	managerCmd.Flags().StringVarP(&server.Address, "server.address", "", "0.0.0.0:8182", "Address to bind on.")
	envs["BIVAC_SERVER_ADDRESS"] = "server.address"
	managerCmd.Flags().StringVarP(&server.PSK, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	managerCmd.Flags().StringVarP(&Orchestrator, "orchestrator", "o", "", "Orchestrator on which Bivac should connect to.")
	envs["BIVAC_ORCHESTRATOR"] = "orchestrator"

	managerCmd.Flags().StringVarP(&Orchestrators.Docker.Endpoint, "docker.endpoint", "", "unix:///var/run/docker.sock", "Docker endpoint.")
	envs["BIVAC_DOCKER_ENDPOINT"] = "docker.endpoint"

	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.URL, "cattle.url", "", "", "The Cattle URL.")
	envs["CATTLE_URL"] = "cattle.url"
	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.AccessKey, "cattle.accesskey", "", "", "The Cattle access key.")
	envs["CATTLE_ACCESS_KEY"] = "cattle.accesskey"
	managerCmd.Flags().StringVarP(&Orchestrators.Cattle.SecretKey, "cattle.secretkey", "", "", "The Cattle secret key.")
	envs["CATTLE_SECRET_KEY"] = "cattle.secretkey"

	managerCmd.Flags().StringVarP(&resticForgetArgs, "restic.forget.args", "", "--keep-daily 15 --prune", "Restic forget arguments.")
	envs["RESTIC_FORGET_ARGS"] = "restic.forget.args"

	managerCmd.Flags().StringVarP(&providersFile, "providers.config", "", "/providers-config.default.toml", "Configuration file for providers.")
	envs["BIVAC_PROVIDERS_CONFIG"] = "providers.config"

	managerCmd.Flags().StringVarP(&targetURL, "target.url", "r", "", "The target URL to push the backups to.")
	envs["BIVAC_TARGET_URL"] = "target.url"

	managerCmd.Flags().IntVarP(&retryCount, "retry.count", "", 0, "Retry to backup the volume if something goes wrong with Bivac.")
	envs["BIVAC_RETRY_COUNT"] = "retry.count"

	cmd.SetValuesFromEnv(envs, managerCmd.Flags())
	cmd.RootCmd.AddCommand(managerCmd)
}
