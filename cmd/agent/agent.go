package agent

import (
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/agent"
)

var (
	targetURL  string
	backupPath string
	hostname   string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run Bivac agent",
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "backup":
			agent.Backup(targetURL, backupPath, hostname)
		case "restore":
			agent.Restore(targetURL, backupPath, hostname)
		}
	},
}

func init() {
	agentCmd.Flags().StringVarP(&targetURL, "target.url", "r", "", "The target URL to push the backups to.")
	agentCmd.Flags().StringVarP(&backupPath, "backup.path", "p", "", "Path to the volume to backup.")
	agentCmd.Flags().StringVarP(&hostname, "host", "", "", "Custom hostname.")
	cmd.RootCmd.AddCommand(agentCmd)
}
