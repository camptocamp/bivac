Bivac : Backup Interface for Volumes Attached to Containers
===========================================================

Website: [https://camptocamp.github.io/bivac](https://camptocamp.github.io/bivac)


[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/bivac.svg)](https://hub.docker.com/r/camptocamp/bivac/)
[![Build Status](https://img.shields.io/travis/camptocamp/bivac/master.svg)](https://travis-ci.org/camptocamp/bivac)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/bivac.svg)](https://coveralls.io/r/camptocamp/bivac?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/bivac)](https://goreportcard.com/report/github.com/camptocamp/bivac)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Bivac lets you backup all your containers volumes deployed on Docker Engine or Kubernetes using Restic, Duplicity or RClone.

![Bivac](img/bivac_small.png)


## Installing

```shell
$ go get github.com/camptocamp/bivac
```

## Usage

```shell
Usage:
  bivac [OPTIONS]

Application Options:
  -V, --version                Display version.
  -l, --loglevel=              Set loglevel ('debug', 'info', 'warn', 'error', 'fatal', 'panic'). (default: info) [$BIVAC_LOG_LEVEL]
  -b, --blacklist=             Volumes to blacklist in backups. [$BIVAC_VOLUMES_BLACKLIST]
  -m, --manpage                Output manpage.
      --no-verify              Do not verify backup. [$BIVAC_NO_VERIFY]
  -j, --json                   Log as JSON (to stderr). [$BIVAC_JSON_OUTPUT]
  -E, --engine=                Backup engine to use. (default: restic) [$BIVAC_ENGINE]
  -o, --orchestrator=          Container orchestrator to use. (default: docker) [$BIVAC_ORCHESTRATOR]
  -u, --target-url=            The target URL to push to. [$BIVAC_TARGET_URL]
  -H, --hostname-from-rancher  Retrieve hostname from Rancher metadata. [$BIVAC_HOSTNAME_FROM_RANCHER]
      --check-every=           Time between backup checks. (default: 24h) [$BIVAC_CHECK_EVERY]
      --remove-older-than=     Remove backups older than the specified interval. (default: 30D) [$BIVAC_REMOVE_OLDER_THAN]
      --label-prefix=          The volume prefix label. [$BIVAC_LABEL_PREFIX]

Restic Options:
      --restic-image=          The restic docker image. (default: restic/restic:latest) [$RESTIC_DOCKER_IMAGE]
      --restic-password=       The restic backup password. [$RESTIC_PASSWORD]

RClone Options:
      --rclone-image=          The rclone docker image. (default: camptocamp/rclone:1.33-1) [$RCLONE_DOCKER_IMAGE]

Duplicity Options:
      --duplicity-image=       The duplicity docker image. (default: camptocamp/duplicity:latest) [$DUPLICITY_DOCKER_IMAGE]
      --full-if-older-than=    The number of days after which a full backup must be performed. (default: 15D) [$BIVAC_FULL_IF_OLDER_THAN]

Metrics Options:
  -g, --gateway-url=           The prometheus push gateway URL to use. [$PUSHGATEWAY_URL]

AWS Options:
      --aws-access-key-id=     The AWS access key ID. [$AWS_ACCESS_KEY_ID]
      --aws-secret-key-id=     The AWS secret access key. [$AWS_SECRET_ACCESS_KEY]

Swift Options:
      --swift-username=        The Swift user name. [$SWIFT_USERNAME]
      --swift-password=        The Swift password. [$SWIFT_PASSWORD]
      --swift-auth_url=        The Swift auth URL. [$SWIFT_AUTHURL]
      --swift-tenant-name=     The Swift tenant name. [$SWIFT_TENANTNAME]
      --swift-region-name=     The Swift region name. [$SWIFT_REGIONNAME]

Docker Options:
  -e, --docker-endpoint=       The Docker endpoint. (default: unix:///var/run/docker.sock) [$DOCKER_ENDPOINT]

Kubernetes Options:
      --k8s-namespace=         Namespace where you want to run Bivac. [$K8S_NAMESPACE]
      --k8s-kubeconfig=        Path to your kubeconfig file. [$K8S_KUBECONFIG]

Help Options:
  -h, --help                   Show this help message
```

## Examples

### Backup all named volumes to S3

```shell
$ bivac \
  -u s3+http://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
  --aws-access-key-id=<my_key_id> \
  --aws-secret-key-id=<my_secret_key>
```


### Using docker

```shell
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro  --rm -ti \
   -e BIVAC_TARGET_URL=s3+http://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
   -e AWS_ACCESS_KEY_ID=<my_key_id> \
   -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
     camptocamp/bivac
```


## Controlling backup parameters

The parameters used to backup each volume can be fine-tuned using volume labels (requires Docker 1.11.0 or greater):

- `io.bivac.ignore=true` ignores the volume
- `io.bivac.no_verify=true` skips verification of the volume's backup (faster)
- `io.bivac.duplicity.full_if_older_than=<value>` sets the time period after which a full backup is performed. Defaults to the `BIVAC_FULL_IF_OLDER_THAN` environment variable value
- `io.bivac.duplicity.remove_older_than=<value>` sets the time period after which to remove older backups. Defaults to the `BIVAC_REMOVE_OLDER_THAN` environment variable value

If you cannot use volume labels, you can drop a `.bivac.overrides` file at the root of the volume:

```ini
engine = "rclone"
no_verify = true
ignore = false
target_url = "s3+http://s3-us-east-1.amazonaws.com/foo/bar"

[duplicity]
full_if_older_than = "3D"
remove_older_than = "5D"
```

## Orchestrators

Bivac supports runing on either Docker Engine (using the Docker API) or Kubernetes (using the Kubernetes API).

### Docker

Bivac will backup all named volumes by default.

### Kubernetes

Bivac will backup all Persistent Volume Claims by default.

## Providers


Bivac detects automatically the kind of data that is stored on a volume and adapts its backup strategy to it. The following providers and associated strategies are currently supported:

* PostgreSQL: Run `pg_dumpall` before backup
* MySQL: Run `mysqldump` before backup
* OpenLDAP: Run `slapcat` before backup
* Default: Backup volume data as is

**Note:** in order to detect providers, bivac needs to access the files in the
volume. When running in a Docker container, you need to mount the Docker
volumes directory for this feature to work, by adding `-v
/var/lib/docker/volumes:/var/lib/docker/volumes:ro` to the Docker command line.


## Engines

Bivac supports various engines for performing the backup:

* Restic
* RClone: use for heavy data that Restic or Duplicity cannot manage efficiently
* Duplicity

You can set the engine with either:

* an `io.bivac.engine` volume label (requires Docker 1.11.0 or great)
* a global setting using the `BIVAC_ENGINE` environment variable
* the `engine` parameter in the `.bivac.overrides` file at the root of the volume


## Return code

Bivac returns:

* `0` if nothing failed
* `1` if a backup failed
* `2` if pushing metrics to Prometheus failed

