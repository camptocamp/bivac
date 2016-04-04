package main

import (
  "os"
  "unicode/utf8"
  "github.com/fsouza/go-dockerclient"
  "github.com/fgrehm/go-dockerpty"
	log "github.com/Sirupsen/logrus"
)

type Conplicity struct {
  Hostname string
  Client   *docker.Client
}

func main() {
  log.Infof("Starting backup...")

  var err error

  c := &Conplicity{}
  c.Hostname, err = os.Hostname()
  checkErr(err, "Failed to get hostname: %v", 1)

  endpoint := "unix:///var/run/docker.sock"

  c.Client, err = docker.NewClient(endpoint)
  checkErr(err, "Failed to create Docker client: %v", 1)

  vols, err := c.Client.ListVolumes(docker.ListVolumesOptions{})
  checkErr(err, "Failed to list Docker volumes: %v", 1)

  for _, vol := range vols {
    err = c.backupVolume(vol)
    checkErr(err, "Failed to process volume "+vol.Name+": %v", -1)
  }

  log.Infof("End backup...")
}


func (c *Conplicity) backupVolume(vol docker.Volume) (err error) {
    if utf8.RuneCountInString(vol.Name) == 64 {
      log.Infof("Ignoring volume "+vol.Name)
      return
    }

    // TODO: detect if it's a Database volume (PostgreSQL, MySQL, OpenLDAP...) and launch DUPLICITY_PRECOMMAND instead of backuping the volume
    log.Infof("ID: "+vol.Name)
    log.Infof("Driver: "+vol.Driver)
    log.Infof("Mountpoint: "+vol.Mountpoint)
    log.Infof("Creating duplicity container...")
    container, err := c.Client.CreateContainer(
      docker.CreateContainerOptions{
        Config: &docker.Config{
          Cmd: []string{
            "--full-if-older-than", "15D",
            "--s3-use-new-style",
            "--no-encryption",
            "--allow-source-mismatch",
            "/var/backups",
            os.Getenv("DUPLICITY_TARGET_URL")+"/"+c.Hostname+"/"+vol.Name,
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

    checkErr(err, "Failed to create container for volume "+vol.Name+": %v", 1)

    defer func() {
      c.Client.RemoveContainer(docker.RemoveContainerOptions{
        ID: container.ID,
        Force: true,
      })
    }()

    binds := []string{
      vol.Mountpoint+":/var/backups:ro",
    }

    err = dockerpty.Start(c.Client, container, &docker.HostConfig{
      Binds: binds,
    })
    checkErr(err, "Failed to start container for volume "+vol.Name+": %v", -1)
    return
}

func checkErr(err error, msg string, exit int) {
  if err != nil {
    log.Errorf(msg, err)

    if exit != -1 {
      os.Exit(exit)
    }
  }
}
