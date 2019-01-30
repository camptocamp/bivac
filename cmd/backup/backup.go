package backup

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
	force         bool
)

var envs = make(map[string]string)

var backupCmd = &cobra.Command{
	Use:   "backup [VOLUME_NAME]",
	Short: "Backup a volume",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(remoteAddress, psk)
		if err != nil {
			log.Errorf("failed to create new client: %s", err)
			return
		}
		err = c.BackupVolume(args[0], force)
		if err != nil {
			log.Errorf("failed to backup volume: %s", err)
			return
		}

		volumes, err := c.GetVolumes()
		if err != nil {
			log.Errorf("failed to get volumes: %s", err)
			return
		}

		for _, v := range volumes {
			if v.ID == args[0] {
				tbl, err := prettytable.NewTable([]prettytable.Column{
					{},
					{},
				}...)
				if err != nil {
					log.WithFields(log.Fields{
						"volume":   v.Name,
						"hostname": v.Hostname,
					}).Errorf("failed to format output: %s", err)
					return
				}
				tbl.Separator = "\t"

				fmt.Printf("ID: %s\n", v.ID)
				fmt.Printf("Name: %s\n", v.Name)
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
	},
}

func init() {
	backupCmd.Flags().StringVarP(&remoteAddress, "remote.address", "", "http://127.0.0.1:8182", "Address of the remote Bivac server.")
	envs["BIVAC_REMOTE_ADDRESS"] = "remote.address"

	backupCmd.Flags().StringVarP(&psk, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	backupCmd.Flags().BoolVarP(&force, "force", "", false, "Force backup by removing locks.")

	cmd.SetValuesFromEnv(envs, backupCmd.Flags())
	cmd.RootCmd.AddCommand(backupCmd)
}
