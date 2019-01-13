package manager

import (
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/manager"
	"github.com/camptocamp/bivac/pkg/orchestrators"
)

var (
	// Server stores informations relative the Bivac server
	Server struct {
		Address string
		PSK     string
	}

	// Orchestrator is the name of the orchestrator on which Bivac should connect to
	Orchestrator string

	Orchestrators struct {
		// Docker stores informations relative to the Docker orchestrator
		Docker orchestrators.DockerConfig
	}
)
var envs = make(map[string]string)

// TODO: Rename this command to something more explicit
var managerCmd = &cobra.Command{
	Use:   "manager",
	Short: "Start Bivac backup manager",
	Run:   manager.Start,
}

func init() {
	managerCmd.Flags().StringVarP(&Server.Address, "server.address", "", "0.0.0.0:8182", "Address to bind on.")
	envs["BIVAC_SERVER_ADDRESS"] = "server.address"
	managerCmd.Flags().StringVarP(&Server.PSK, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	managerCmd.Flags().StringVarP(&Orchestrator, "orchestrator", "o", "", "Orchestrator on which Bivac should connect to.")
	envs["BIVAC_ORCHESTRATOR"] = "orchestrator"

	managerCmd.Flags().StringVarP(&Docker.Endpoint, "docker.endpoint", "", "unix:///var/run/docker.sock", "Docker endpoint.")
	envs["BIVAC_DOCKER_ENDPOINT"] = "docker.endpoint"

	cmd.SetValuesFromEnv(envs, managerCmd.Flags())
	cmd.RootCmd.AddCommand(managerCmd)
}
