package agent

import (
	"os"

	//log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/agent"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run Bivac agent",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(agent.Start())
	},
}

func init() {
	cmd.RootCmd.AddCommand(agentCmd)
}
