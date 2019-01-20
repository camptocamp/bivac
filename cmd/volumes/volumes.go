package volumes

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/pkg/client"
)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "List volumes",
	Run: func(cmd *cobra.Command, args []string) {
		_, err := client.NewClient("http://127.0.0.1:8182", "foo")
		if err != nil {
			log.Errorf("failed to create new client: %s", err)
			return
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(volumesCmd)
}
