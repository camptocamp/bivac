#!/bin/bash

export VAGRANT_VAGRANTFILE=Vagrantfile.builder

if [ -f bivac-cattle.box ]; then
  rm bivac-cattle.box
fi

vagrant up -d
vagrant package --output bivac-cattle.box
vagrant destroy -f

vagrant box list | grep bivac-cattle && vagrant box remove bivac-cattle
