Conplicity
==========

Website: [https://camptocamp.github.io/conplicity](https://camptocamp.github.io/conplicity)


[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/conplicity.svg)](https://hub.docker.com/r/camptocamp/conplicity/)
[![Build Status](https://img.shields.io/travis/camptocamp/conplicity/master.svg)](https://travis-ci.org/camptocamp/conplicity)
[![Coverage Status](https://img.shields.io/coveralls/camptocamp/conplicity.svg)](https://coveralls.io/r/camptocamp/conplicity?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/camptocamp/conplicity)](https://goreportcard.com/report/github.com/camptocamp/conplicity)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


Conplicity lets you backup all your named Docker volumes using Duplicity or RClone.

![Conplicity](img/conplicity_small.png)


## Installing

```shell
$ go get github.com/camptocamp/conplicity
```

## Usage

```shell
Usage:
  conplicity [OPTIONS]

Application Options:
  -V, --version                Display version.
  -l, --loglevel=              Set loglevel ('debug', 'info', 'warn', 'error', 'fatal', 'panic'). (default: info)
                               [$CONPLICITY_LOG_LEVEL]
  -b, --blacklist=             Volumes to blacklist in backups. [$CONPLICITY_VOLUMES_BLACKLIST]
  -m, --manpage                Output manpage.
      --no-verify              Do not verify backup. [$CONPLICITY_NO_VERIFY]
  -j, --json                   Log as JSON (to stderr). [$CONPLICITY_JSON_OUTPUT]
  -E, --engine=                Backup engine to use. (default: duplicity) [$CONPLICITY_ENGINE]
  -u, --target-url=            The target URL to push to. [$CONPLICITY_TARGET_URL]
  -H, --hostname-from-rancher  Retrieve hostname from Rancher metadata. [$CONPLICITY_HOSTNAME_FROM_RANCHER]

Duplicity Options:
      --duplicity-image=       The duplicity docker image. (default: camptocamp/duplicity:latest) [$DUPLICITY_DOCKER_IMAGE]
      --full-if-older-than=    The number of days after which a full backup must be performed. (default: 15D)
                               [$CONPLICITY_FULL_IF_OLDER_THAN]
      --remove-older-than=     The number days after which backups must be removed. (default: 30D) [$CONPLICITY_REMOVE_OLDER_THAN]

RClone Options:
      --rclone-image=          The rclone docker image. (default: camptocamp/rclone:latest) [$RCLONE_DOCKER_IMAGE]

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

Help Options:
  -h, --help                   Show this help message
```

## Examples

### Backup all named volumes to S3

```shell
$ conplicity \
  -u s3+http://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
  --aws-access-key-id=<my_key_id> \
  --aws-secret-key-id=<my_secret_key>
```


### Using docker

```shell
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro  --rm -ti \
   -e CONPLICITY_TARGET_URL=s3+http://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
   -e AWS_ACCESS_KEY_ID=<my_key_id> \
   -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
     camptocamp/conplicity
```


## Controlling backup parameters

The parameters used to backup each volume can be fine-tuned using volume labels (requires Docker 1.11.0 or greater):

- `io.conplicity.ignore=true` ignores the volume
- `io.conplicity.no_verify=true` skips verification of the volume's backup (faster)
- `io.conplicity.duplicity.full_if_older_than=<value>` sets the time period after which a full backup is performed. Defaults to the `CONPLICITY_FULL_IF_OLDER_THAN` environment variable value
- `io.conplicity.duplicity.remove_older_than=<value>` sets the time period after which to remove older backups. Defaults to the `CONPLICITY_REMOVE_OLDER_THAN` environment variable value

If you cannot use volume labels, you can drop a `.conplicity.overrides` file at the root of the volume:

```ini
engine = "rclone"
no_verify = true
ignore = false
target_url = "s3+http://s3-us-east-1.amazonaws.com/foo/bar"

[duplicity]
full_if_older_than = "3D"
remove_older_than = "5D"
```


## Providers


Conplicity detects automatically the kind of data that is stored on a volume and adapts its backup strategy to it. The following providers and associated strategies are currently supported:

* PostgreSQL: Run `pg_dumpall` before backup
* MySQL: Run `mysqldump` before backup
* OpenLDAP: Run `slapcat` before backup
* Default: Backup volume data as is

**Note:** in order to detect providers, conplicity needs to access the files in the
volume. When running in a Docker container, you need to mount the Docker
volumes directory for this feature to work, by adding `-v
/var/lib/docker/volumes:/var/lib/docker/volumes:ro` to the Docker command line.


## Engines

Conplicity supports various engines for performing the backup:

* Duplicity
* RClone: use for heavy data that Duplicity cannot manage efficiently

You can set the engine with either:

* an `io.conplicity.engine` volume label (requires Docker 1.11.0 or great)
* a global setting using the `CONPLICITY_ENGINE` environment variable
* the `engine` parameter in the `.conplicity.overrides` file at the root of the volume


## Return code

Conplicity returns:

* `0` if nothing failed
* `1` if a backup failed
* `2` if pushing metrics to Prometheus failed

