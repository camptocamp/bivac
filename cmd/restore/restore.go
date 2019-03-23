package restore

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/pkg/client"
	"github.com/spf13/cobra"
	"github.com/tatsushid/go-prettytable"
)

var (
	force         bool
	psk           string
	remoteAddress string
	snapshotName  string
)

var envs = make(map[string]string)

var restoreCmd = &cobra.Command{
	Use:   "restore [VOLUME_NAME]",
	Short: "Restore volumes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(remoteAddress, psk)
		if err != nil {
			log.Errorf("failed to create new client: %s", err)
			return
		}
		for _, a := range args {
			fmt.Printf("Restoring `%s'...\n", a)
			err = c.RestoreVolume(a, force, snapshotName)
			if err != nil {
				log.Errorf("failed to restore volume: %s", err)
				return
			}
		}
		volumes, err := c.GetVolumes()
		if err != nil {
			log.Errorf("failed to get volumes: %s", err)
			return
		}
		for _, a := range args {
			for _, v := range volumes {
				if v.ID == a {
					tbl, err := prettytable.NewTable(
						[]prettytable.Column{
							{},
							{},
						}...,
					)
					if err != nil {
						log.WithFields(log.Fields{
							"volume":   v.Name,
							"hostname": v.Hostname,
						}).Errorf(
							"failed to format output: %s",
							err,
						)
						return
					}
					tbl.Separator = "\t"
					fmt.Printf("ID: %s\n", v.ID)
					fmt.Printf("Name: %s\n", v.Name)
					fmt.Printf(
						"Mountpoint: %s\n",
						v.Mountpoint,
					)
					fmt.Printf(
						"Backup date: %s\n",
						v.LastBackupDate,
					)
					fmt.Printf(
						"Backup status: %s\n",
						v.LastBackupStatus,
					)
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
	restoreCmd.Flags().StringVarP(
		&remoteAddress,
		"remote.address",
		"",
		"http://127.0.0.1:8182",
		"Address of the remote Bivac server.",
	)
	envs["BIVAC_REMOTE_ADDRESS"] = "remote.address"
	restoreCmd.Flags().StringVarP(
		&psk,
		"server.psk",
		"",
		"",
		"Pre-shared key.",
	)
	envs["BIVAC_SERVER_PSK"] = "server.psk"
	restoreCmd.Flags().BoolVarP(
		&force,
		"force",
		"",
		false,
		"Force restore by removing locks.",
	)
	restoreCmd.Flags().StringVarP(
		&snapshotName,
		"snapshot",
		"s",
		"latest",
		"Name of snapshot to restore",
	)
	cmd.SetValuesFromEnv(envs, restoreCmd.Flags())
	cmd.RootCmd.AddCommand(restoreCmd)
}
