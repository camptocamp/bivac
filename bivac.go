package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types/filters"
	docker "github.com/docker/docker/client"

	"github.com/camptocamp/bivac/handler"
)

var version = "undefined"
var exitCode int

func main() {

	c, err := handler.NewBivac(version)
	if err != nil {
		log.Fatalf("Failed to initialize bivac: %v", err)
		os.Exit(1)
	}

	if c.Config.ServerMode {
		startServer(c)
	} else if c.Config.CLIMode {
		//startCLI(c)
	} else {
		log.Fatalf("No mode selected.")
		os.Exit(1)
	}
	return
}

func startServer(c *handler.Bivac) {
	// Start volume manager thread
	volumesManagerStopped := make(chan struct{})
	volumesManagerStop := make(chan struct{})
	go volumesManager(c, volumesManagerStopped, volumesManagerStop)

	// Start API server thread
	apiStopped := make(chan struct{})
	apiStop := make(chan struct{})
	go apiServer(c, apiStopped, apiStop)

	defer teardownThreads(exitCode, volumesManagerStop, volumesManagerStopped, apiStop, apiStopped)

	for {
		select {
		case <-volumesManagerStopped:
			exitCode = 1
			return
		case <-apiStopped:
			exitCode = 1
			return
		default:
		}
	}
}

func volumesManager(c *handler.Bivac, thStopped, thStop chan struct{}) {
	defer close(thStopped)

	for {
		select {
		default:
			d, err := docker.NewClient("unix:///var/run/docker.sock", "", nil, nil)
			if err != nil {
				log.Errorf("%v", err)
			}
			vols, _ := d.VolumeList(context.Background(), filters.NewArgs())
			for _, v := range vols.Volumes {
				log.Infof("Volume: %s", v.Name)
			}
			time.Sleep(10000 * time.Millisecond)
			return
		case <-thStop:
			return
		}
	}
}

func apiServer(c *handler.Bivac, thStopped, thStop chan struct{}) {
	defer close(thStopped)

	router := mux.NewRouter()
	router.HandleFunc("/volumes", GetVolumes).Methods("GET")

	for {
		select {
		default:
			log.Fatal(http.ListenAndServe(":8080", router))
		case <-thStop:
			return
		}
	}
}

// GetVolumes is an HTTP endpoint that list volumes
func GetVolumes(w http.ResponseWriter, r *http.Request) {
}

func teardownThreads(exitCode int, volumesManagerStop, volumesManagerStopped, apiStop, apiStopped chan struct{}) int {
	log.Info("Stopping volumes manager...")
	close(volumesManagerStop)
	<-volumesManagerStopped
	log.Info("Volumes manager stopped.")

	log.Info("Stopping API server...")
	close(apiStop)
	<-apiStopped
	log.Info("API server stopped.")

	return exitCode
}
