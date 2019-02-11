package info

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/camptocamp/bivac/cmd"
	"github.com/camptocamp/bivac/pkg/client"
)

var (
	remoteAddress string
	psk           string
)

var envs = make(map[string]string)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Retrive Bivac informations",
	Run: func(cmd *cobra.Command, args []string) {
		c, err := client.NewClient(remoteAddress, psk)
		if err != nil {
			log.Errorf("failed to create a new client: %s", err)
			return
		}

		informations, err := c.GetInformations()
		if err != nil {
			log.Errorf("failed to get informations: %s", err)
			return
		}

		for infok, infov := range informations {
			if infok == "volumes_count" {
				infok = "Managed volumes"
			}
			fmt.Printf("%s: %s\n", strings.Title(infok), infov)
		}
	},
}

func init() {
	infoCmd.Flags().StringVarP(&remoteAddress, "remote.address", "", "http://127.0.0.1:8182", "Address of the remote Bivac server.")
	envs["BIVAC_REMOTE_ADDRESS"] = "remote.address"

	infoCmd.Flags().StringVarP(&psk, "server.psk", "", "", "Pre-shared key.")
	envs["BIVAC_SERVER_PSK"] = "server.psk"

	cmd.SetValuesFromEnv(envs, infoCmd.Flags())
	cmd.RootCmd.AddCommand(infoCmd)

}
