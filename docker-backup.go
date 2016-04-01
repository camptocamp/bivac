package main

import (
  "fmt"
  "github.com/fsouza/go-dockerclient"
)

func main() {
  fmt.Println("Starting backup...")

  endpoint := "unix:///var/run/docker.sock"
  client, _ := docker.NewClient(endpoint)
  vols, _ := client.ListVolumes(docker.ListVolumesOptions{})
  for _, vol := range vols {
    fmt.Println("ID: ", vol.Name)
    fmt.Println("Driver: ", vol.Driver)
    fmt.Println("Mountpoint: ", vol.Mountpoint)
  }

  fmt.Println("End backup...")
}
