package main

import (
  "fmt"
  "os"
  "unicode/utf8"
  "github.com/fsouza/go-dockerclient"
  "github.com/fgrehm/go-dockerpty"
)

func main() {
  fmt.Println("Starting backup...")

  hostname, _ := os.Hostname()

  endpoint := "unix:///var/run/docker.sock"
  client, _ := docker.NewClient(endpoint)
  vols, _ := client.ListVolumes(docker.ListVolumesOptions{})
  for _, vol := range vols {
    if utf8.RuneCountInString(vol.Name) == 64 {
      fmt.Println("Ignoring volume ", vol.Name)
      continue
    }

    // TODO: detect if it's a Database volume (PostgreSQL, MySQL, OpenLDAP...) and launch DUPLICITY_PRECOMMAND instead of backuping the volume
    fmt.Println("ID: ", vol.Name)
    fmt.Println("Driver: ", vol.Driver)
    fmt.Println("Mountpoint: ", vol.Mountpoint)
    fmt.Println("Creating duplicity container...")
    container, err := client.CreateContainer(
      docker.CreateContainerOptions{
        Config: &docker.Config{
          Cmd: []string{
            "--full-if-older-than", "15D",
            "--s3-use-new-style",
            "--no-encryption",
            "--allow-source-mismatch",
            "/var/backups",
            os.Getenv("DUPLICITY_TARGET_URL")+"/"+hostname+"/"+vol.Name,
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
          OpenStdin:    true,
          StdinOnce:    true,
          AttachStdin:  true,
          AttachStdout: true,
          AttachStderr: true,
          Tty:          true,
        },
      },
    )

    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }

    defer func() {
      client.RemoveContainer(docker.RemoveContainerOptions{
        ID: container.ID,
        Force: true,
      })
    }()

    binds := []string{
      vol.Mountpoint+":/var/backups:ro",
    }

    fmt.Println("Starting duplicity container...")
    err = dockerpty.Start(client, container, &docker.HostConfig{
      Binds: binds,
    })

    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  }

  fmt.Println("End backup...")
}
