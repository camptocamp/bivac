package volumes

import (
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "List volumes",
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	cmd.RootCmd.AddCommand(volumesCmd)
}
