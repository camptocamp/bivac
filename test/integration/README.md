# Integration testing

## Usage

To run the tests, you can use the script `tests.sh`. The following list describe the accepted arguments:

* `-i <docker image>`: the docker image to test.
* `-o <orchestrator>`: the orchestrator on which the tests will be run. You can leave the field blank if you want to test your code on every orchestrator.
* `-l <log level>`: the Bivac log level (default is `error`).
* `-b`: build the virtual machines before running the tests.

### Examples

Test the latest version of Bivac (unstable) on Docker:

`./tests.sh -i camptocamp/bivac:latest -o docker`

Build and test the latest stable release of Bivac on Docker, Cattle and Kubernetes:

`./tests.sh -i camptocamp/bivac:stable -b`

## Development

### Architecture

The main script is `tests.sh` and it allows you to easily run the tests. Under the directories `docker`, `cattle` and `kubernetes`, you can find the following content:

* `build.sh`: shell script that build a box from the file `Vagrantfile.builder`.
* `Vagrantfile.builder`: Vagrant file used to build a VM on which the tests will be run. 
* `Vagrantfile.runner`: Vagrant file which is used to run the tests. This VM is based on the box created from `Vagrantfile.builder`. Using two separated Vagrant files allows us to create and destroy the `runner` without having to rebuild the VM from scratch.
* `tests`: directory that contains shell scripts, the test scenarios.
