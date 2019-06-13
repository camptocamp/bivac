# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.hostname = "testing"

  config.ssh.insert_key = false


  config.vm.provider "virtualbox" do |v|
    v.memory = 4096
    v.cpus = 2
  end

  config.vm.provision "shell", inline: <<-SHELL
    sudo apt-get update
    sudo apt-get upgrade -y
    sudo apt-get install apt-transport-https ca-certificates curl software-properties-common jq wget net-tools iproute2 -y
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
  SHELL

  config.vm.provision "shell", path: "prepare.sh"
end
