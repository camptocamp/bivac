package manager

import (
	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	//"github.com/camptocamp/bivac/internal/server"
	//"github.com/camptocamp/bivac/pkg/orchestrators"
)

// Start starts a Bivac manager which handle backups management
func Start(cmd *cobra.Command, args []string) {
	_, _ = cmd.Flags().GetString("orchestrator")
	dockerConfig, _ := cmd.Flags().GetString("docker")
	log.Infof("%+v", dockerConfig)

	//o, err := getOrchestrator()
	//if err != nil {
	//	log.Errorf("failed to retrieve orchestrator: %s", err)
	//	return
	//}

	//err = server.Start(o)
	//if err != nil {
	//	log.Errorf("failed to start server: %s", err)
	//}
	//return
}

//func getOrchestrator() (o orchestrators.Orchestrator, err error) {
//	if manager.Orchestrator != "" {
//		log.Debugf("Choosing orchestrator based on configuration...")
//		switch manager.Orchestrator {
//		case "docker":
//			o, err = NewDockerOrchestrator(manager.Docker)
//		default:
//			err = fmt.Errorf("'%s' is not a valid orchestrator")
//			return
//		}
//	} else {
//		log.Debugf("Trying to detect orchestrator based on environment...")
//		if detectDocker() {
//			o, err = NewDockerOrchestrator(manager.Docker)
//		} else {
//			err = fmt.Errorf("no orchestrator detected")
//			return
//		}
//	}
//	if err != nil {
//		log.Infof("Using orchestrator: %s", o.GetName())
//	}
//	return
//}
