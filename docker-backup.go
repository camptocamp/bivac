package main

import (
  "fmt"
  "os"
  "github.com/fsouza/go-dockerclient"
)

func main() {
  fmt.Println("Starting backup...")

  endpoint := "unix:///var/run/docker.sock"
  client, _ := docker.NewClient(endpoint)
  vols, _ := client.ListVolumes(docker.ListVolumesOptions{})
  for _, vol := range vols {
    // TODO: filter out unnamed volumes (name has 64 characters?)
    // TODO: detect if it's a Database volume (PostgreSQL, MySQL, OpenLDAP...) and launch DUPLICITY_PRECOMMAND instead of backuping the volume
    fmt.Println("ID: ", vol.Name)
    fmt.Println("Driver: ", vol.Driver)
    fmt.Println("Mountpoint: ", vol.Mountpoint)
    fmt.Println("Creating duplicity container...")
    client.CreateContainer(
      docker.CreateContainerOptions{
        Config: &docker.Config{
          Cmd: []string{
            "--full-if-older-than", "15D",
            "--s3-use-new-style",
            "--no-encryption",
            "--allow-source-mismatch",
            "/var/backups",
            os.Getenv("DUPLICITY_TARGET_URL"),
          },
          Env: []string{
            "AWS_ACCESS_KEY_ID="+os.Getenv("AWS_ACCESS_KEY_ID"),
            "AWS_SECRET_ACCESS_KEY="+os.Getenv("AWS_SECRET_ACCESS_KEY"),
            "SWIFT_USERNAME="+os.Getenv("SWIFT_USERNAME"),
            "SWIFT_PASSWORD="+os.Getenv("SWIFT_PASSWORD"),
            "SWIFT_AUTHURL="+os.Getenv("SWIFT_AUTHURL"),
            "SWIFT_TENANTNAME="+os.Getenv("SWIFT_TENANTNAME"),
            "SWIFT_REGIONNAME="+os.Getenv("SWIFT_REGIONNAME"),
            "SWIFT_AUTHVERSION=2",
          },
          Image: "camptocamp/duplicity",
          Volumes: map[string]struct{}{/* What should I put here? */},
        },
      },
    )
  }

  fmt.Println("End backup...")
}
