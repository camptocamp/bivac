package agent

import (
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/internal/agent"
)

var (
	targetURL    string
	backupPath   string
	hostname     string
	force        bool
	logReceiver  string
	snapshotName string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run Bivac agent",
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "backup":
			agent.Backup(targetURL, backupPath, hostname, force, logReceiver)
		case "restore":
			agent.Restore(targetURL, backupPath, hostname, force, logReceiver, snapshotName)
		}
	},
}

func init() {
	agentCmd.Flags().StringVarP(&targetURL, "target.url", "r", "", "The target URL to push the backups to.")
	agentCmd.Flags().StringVarP(&backupPath, "backup.path", "p", "", "Path to the volume to backup.")
	agentCmd.Flags().StringVarP(&hostname, "host", "", "", "Custom hostname.")
	agentCmd.Flags().BoolVarP(&force, "force", "", false, "Force a backup by removing all locks.")
	agentCmd.Flags().StringVarP(&logReceiver, "log.receiver", "", "", "Address where the manager will collect the logs.")
	agentCmd.Flags().StringVarP(&snapshotName, "snapshot", "s", "latest", "Name of snapshot to restore")
	cmd.RootCmd.AddCommand(agentCmd)
}
