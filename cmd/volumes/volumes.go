package volumes

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-prettytable"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/pkg/client"
)

var (
	remoteAddress string
	psk           string
)

var envs = make(map[string]string)

var volumesCmd = &cobra.Command{
	Use:   "volumes",
	Short: "List volumes",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(remoteAddress, psk)
		if err != nil {
			log.Errorf("failed to create new client: %s", err)
			return
		}

		volumes, err := c.GetVolumes()
		if err != nil {
			log.Errorf("failed to get volumes: %s", err)
			return
		}

		tbl, err := prettytable.NewTable([]prettytable.Column{
			{Header: "ID"},
			{Header: "Name"},
			{Header: "Mountpoint"},
			{Header: "LastBackupDate"},
			{Header: "LastBackupStatus"},
		}...)
		if err != nil {
			log.Errorf("failed to format output: %s", err)
			return
		}
		tbl.Separator = "\t"

		for _, v := range volumes {
			tbl.AddRow(v.ID, v.Name, v.Mountpoint, v.LastBackupDate, v.LastBackupStatus)
		}

		tbl.Print()
	},
}

func init() {
	volumesCmd.Flags().StringVarP(&remoteAddress, "remote.address", "", "http://127.0.0.1:8182", "Address of the remote Bivac server.")
	envs["BIVAC_REMOTE_ADDRESS"] = "remote.address"

	volumesCmd.Flags().StringVarP(&psk, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	cmd.SetValuesFromEnv(envs, volumesCmd.Flags())
	cmd.RootCmd.AddCommand(volumesCmd)
}
