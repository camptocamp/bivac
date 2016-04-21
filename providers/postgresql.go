package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

type PostgreSQLProvider struct {
	handler   *handler.Conplicity
	vol       *docker.Volume
	backupDir string
}

func (p *PostgreSQLProvider) PrepareBackup() (err error) {
	log.Infof("PG_VERSION file found, this should be a PostgreSQL datadir")
	log.Infof("Searching postgres container using this volume...")
	c := p.handler.Client
	vol := p.vol
	containers, _ := c.ListContainers(docker.ListContainersOptions{})
	for _, container := range containers {
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.Infof("Volume %v is used by container %v", vol.Name, container.ID)
				log.Infof("Launch pg_dumpall in container %v...", container.ID)
				exec, err := c.CreateExec(
					docker.CreateExecOptions{
						Container: container.ID,
						Cmd: []string{
							"sh",
							"-c",
							"mkdir -p " + mount.Destination + "/backups && pg_dumpall -Upostgres > " + mount.Destination + "/backups/all.sql",
						},
					},
				)

				checkErr(err, "Failed to create exec", 1)

				err = c.StartExec(
					exec.ID,
					docker.StartExecOptions{},
				)

				checkErr(err, "Failed to create exec", 1)

				p.backupDir = "backups"
			}
		}
	}
	return
}

func (p *PostgreSQLProvider) BackupVolume() (err error) {
	return
}
