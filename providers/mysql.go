package providers

import (
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type MySQLProvider struct {
	*BaseProvider
}

// MySQLProvider implements a BaseProvider struct
// for MySQL backups
func (*MySQLProvider) GetName() string {
	return "MySQL"
}

func (p *MySQLProvider) PrepareBackup() (err error) {
	c := p.handler.Client
	vol := p.vol
	log.Infof("Looking for a mysql container using this volume...")
	containers, _ := c.ListContainers(docker.ListContainersOptions{})
	for _, container := range containers {
		container, err := c.InspectContainer(container.ID)
		checkErr(err, "Failed to inspect container "+container.ID+": %v", -1)
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
