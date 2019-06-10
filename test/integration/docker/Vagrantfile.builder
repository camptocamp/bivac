# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"

  config.ssh.insert_key = false

  config.vm.provision "shell", inline: <<-SHELL
    sudo apt-get update
    sudo apt-get upgrade -y
    sudo apt-get install apt-transport-https ca-certificates curl software-properties-common jq -y
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    sudo apt-get update
    sudo apt-get install docker-ce -y
    sudo curl -L "https://github.com/docker/compose/releases/download/1.22.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    sudo usermod -aG docker vagrant

    # Add default vagrant key
    curl -k https://raw.githubusercontent.com/mitchellh/vagrant/master/keys/vagrant.pub > /home/vagrant/.ssh/authorized_keys
    chmod 0700 /home/vagrant/.ssh
    chmod 0600 /home/vagrant/.ssh/authorized_keys

    docker pull mysql
    docker pull postgres
    docker pull restic/restic

    # Start Minio
    docker run --restart=always -d -p 9000:9000 -e MINIO_ACCESS_KEY=OBQZY3DV6VOEZ9PG6NIM -e MINIO_SECRET_KEY=7e88XeX0j3YdB6b1o0zU2GhG0dX6tFMy3Haty --name minio -v /root/minio:/data minio/minio server /data
    sleep 10
    docker run --rm -e MC_HOST_minio=http://OBQZY3DV6VOEZ9PG6NIM:7e88XeX0j3YdB6b1o0zU2GhG0dX6tFMy3Haty@172.17.0.2:9000 minio/mc mb minio/bivac-testing
    SHELL
end
