#!/bin/bash

docker() {
  if [ "$build" = true ]; then
    echo "[*] Docker : Building environment..."
    (cd docker && ./build.sh)
  else
    echo "[*] Docker : Build skipped."
  fi

  echo "[*] Docker : Running tests..."

  (cd docker && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant up)
  (cd docker && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant ssh -c "run-parts -v /vagrant/tests -a $1 -a $2")
  (cd docker && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant destroy -f)
}

cattle() {
  if [ "$build" = true ]; then
    echo "[*] Cattle: Building environment..."
    (cd cattle && ./build.sh)
  else
    echo "[*] Cattle: Build skipped."
  fi

  echo "[*] Cattle: Running tests..."

  (cd cattle && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant up)
  (cd cattle && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant ssh -c "run-parts -v /vagrant/tests -a $1 -a $2")
  (cd cattle && VAGRANT_VAGRANTFILE=Vagrantfile.runner vagrant destroy -f)
}

usage() { echo "Usage: $0 [-i <docker image>] [-o <orchestrator>] [-l <log level>] [-b]" 1>&2; exit 1; }

log_level=error
while getopts ":i:o:l:b" o; do
  case "${o}" in
    i)
      image=${OPTARG}
      ;;
    o)
      orchestrator=${OPTARG}
      ;;
    l)
      log_level=${OPTARG}
      ;;
    b)
      build=true
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND-1))

if [ -z "${image}" ]; then
  usage
fi


case "$orchestrator" in
  docker)
    docker $image $log_level
    ;;
  cattle)
    cattle $image $log_level
    ;;
  *)
    docker $image $log_level
    cattle $image $log_level
    ;;
esac
