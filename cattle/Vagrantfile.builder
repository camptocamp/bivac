# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.network "private_network", ip: "192.168.42.3"
  config.vm.hostname = "testing"

  config.ssh.insert_key = false


  config.vm.provider "virtualbox" do |v|
    v.memory = 2048
    v.cpus = 2
  end

  config.vm.provision "shell", inline: <<-SHELL
    sudo apt-get update
    sudo apt-get upgrade -y
    sudo apt-get install apt-transport-https ca-certificates curl software-properties-common jq wget -y
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    sudo apt-get update
    sudo apt-get install docker-ce -y
    sudo curl -L "https://github.com/docker/compose/releases/download/1.22.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose
    sudo usermod -aG docker vagrant
    echo '{"dns": ["8.8.8.8","4.4.4.4"]}' | sudo tee /etc/docker/daemon.json
    sudo systemctl restart docker

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
    docker run --rm -e MC_HOSTS_minio=http://OBQZY3DV6VOEZ9PG6NIM:7e88XeX0j3YdB6b1o0zU2GhG0dX6tFMy3Haty@172.17.0.2:9000 minio/mc mb minio/bivac-testing

    # Start Rancher
    docker run -d --restart=unless-stopped -p 8080:8080 rancher/server:stable
    sleep 60
    curl 'http://localhost:8080/v2-beta/setting' -H 'Accept: application/json' -H 'content-type: application/json' --data '{"type":"setting","name":"telemetry.opt","value":"in"}'
    sleep 1
    curl 'http://localhost:8080/v2-beta/settings/api.host' -X PUT -H 'Accept: application/json' -H 'content-type: application/json' --data '{"id":"api.host","type":"activeSetting","baseType":"setting","name":"api.host","activeValue":null,"inDb":false,"source":null,"value":"http://192.168.42.3:8080"}'
    sleep 1
    curl 'http://localhost:8080/v2-beta/projects/1a5/registrationtoken' --data '{"type":"registrationToken"}'
    sleep 1
    command=$(curl -s 'http://localhost:8080/v2-beta/projects/1a5/registrationtokens?state=active&limit=-1&sort=name'  -H 'Accept: application/json'  -H 'content-type: application/json' | jq -r ".data[0].command")
    $command

    # Install Rancher CLI
    wget https://releases.rancher.com/cli/v0.6.12/rancher-linux-amd64-v0.6.12.tar.gz
    tar zxf rancher-linux-amd64-v0.6.12.tar.gz
    sudo cp ./rancher-v0.6.12/rancher /bin/rancher
    sudo chmod +x /bin/rancher
    rm -rf ./rancher-v0.6.12
    rm rancher-linux-amd64-v0.6.12.tar.gz
    SHELL
end
