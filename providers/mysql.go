package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/camptocamp/conplicity/handler"
	"github.com/fsouza/go-dockerclient"
)

type MySQLProvider struct {
	handler   *handler.Conplicity
	vol       *docker.Volume
	backupDir string
}

func (*MySQLProvider) GetName() string {
	return "MySQL"
}

func (p *MySQLProvider) PrepareBackup() (err error) {
	log.Infof("mysql directory found, this should be MySQL datadir")
	log.Infof("Searching mysql container using this volume...")
	c := p.handler.Client
	vol := p.vol
	containers, _ := c.ListContainers(docker.ListContainersOptions{})
	for _, container := range containers {
		for _, mount := range container.Mounts {
			if mount.Name == vol.Name {
				log.Infof("Volume %v is used by container %v", vol.Name, container.ID)
				log.Infof("Launch mysqldump in container %v...", container.ID)
				exec, err := c.CreateExec(
					docker.CreateExecOptions{
						Container: container.ID,
						Cmd: []string{
							"sh",
							"-c",
							"mkdir -p " + mount.Destination + "/backups && mysqldump --all-databases --extended-insert --password=$MYSQL_ROOT_PASSWORD > " + mount.Destination + "/backups/all.sql",
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

func (p *MySQLProvider) BackupVolume() (err error) {
	return
}
