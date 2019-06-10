#!/bin/bash

export VAGRANT_VAGRANTFILE=Vagrantfile.builder

if [ -f bivac-docker.box ]; then
  rm bivac-docker.box
fi

vagrant up -d
vagrant package --output bivac-docker.box
vagrant destroy -f

vagrant box list | grep bivac-docker && vagrant box remove bivac-docker
