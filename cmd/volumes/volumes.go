package volumes

import (
	"fmt"

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
	Short: "Show volumes",
	Args:  cobra.ArbitraryArgs,
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

		if len(args) == 0 {
			tbl, err := prettytable.NewTable([]prettytable.Column{
				{Header: "ID"},
				{Header: "Name"},
				{Header: "Hostname"},
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
				tbl.AddRow(v.ID, v.Name, v.Hostname, v.Mountpoint, v.LastBackupDate, v.LastBackupStatus)
			}

			tbl.Print()
			return
		}

		for _, a := range args {
			for _, v := range volumes {
				if v.ID == a {
					tbl, err := prettytable.NewTable([]prettytable.Column{
						{},
						{},
					}...)
					if err != nil {
						log.Errorf("failed to format output: %s", err)
						return
					}
					tbl.Separator = "\t"

					fmt.Printf("ID: %s\n", v.ID)
					fmt.Printf("Name: %s\n", v.Name)
					fmt.Printf("Hostname: %s\n", v.Hostname)
					fmt.Printf("Mountpoint: %s\n", v.Mountpoint)
					fmt.Printf("Backup date: %s\n", v.LastBackupDate)
					fmt.Printf("Backup status: %s\n", v.LastBackupStatus)
					fmt.Printf("Logs:\n")
					for stepKey, stepValue := range v.Logs {
						tbl.AddRow(stepKey, stepValue)
					}
					tbl.Print()
				}
			}
		}
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
