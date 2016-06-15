Conplicity
==========

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/conplicity.svg)](https://hub.docker.com/r/camptocamp/conplicity/)
[![Build Status](https://img.shields.io/travis/camptocamp/conplicity/master.svg)](https://travis-ci.org/camptocamp/conplicity)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


conplicity lets you backup all your named docker volumes using duplicity.


## Installing

```shell
$ go get github.com/camptocamp/conplicity
```

## Usage

```shell
$ conplicity -h
Usage:
  conplicity [OPTIONS]

Application Options:
  -i, --image=              The duplicity docker image. (default: camptocamp/duplicity:latest) [$DUPLICITY_DOCKER_IMAGE]
  -u, --url=                The duplicity target URL to push to. [$DUPLICITY_TARGET_URL]
  -g, --gateway_url=        The prometheus push gateway URL to use. [$PUSHGATEWAY_URL]
      --aws_access_key_id=  The AWS access key ID. [$AWS_ACCESS_KEY_ID]
      --aws_secret_key_id=  The AWS secret access key. [$AWS_SECRET_ACCESS_KEY]
      --swift_username=     The Swift user name. [$SWIFT_USERNAME]
      --swift_password=     The Swift password. [$SWIFT_PASSWORD]
      --swift_auth_url=     The Swift auth URL. [$SWIFT_AUTHURL]
      --swift_tenant_name=  The Swift tenant name. [$SWIFT_TENANTNAME]
      --swift_region_name=  The Swift region name. [$SWIFT_REGIONNAME]
      --full_if_older_than= The number of days after which a full backup must be performed. (default: 15D) [$CONPLICITY_FULL_IF_OLDER_THAN]
      --remove_older_than=  The number days after which backups must be removed. (default: 30D) [$CONPLICITY_REMOVE_OLDER_THAN]
  -b, --volumes_blacklist=  Volumes to blacklist in backups. [$CONPLICITY_VOLUMES_BLACKLIST]

Help Options:
  -h, --help                Show this help message

exit status 1
```

## Examples

### Backup all named volumes to S3

```shell
$ DUPLICITY_TARGET_URL=s3://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
  AWS_ACCESS_KEY_ID=<my_key_id> \
  AWS_SECRET_ACCESS_KEY=<my_secret_key> \
    conplicity
```


### Using docker

```shell
$ docker run -v /var/run/docker.sock:/var/run/docker.sock:ro  --rm -ti \
   -e DUPLICITY_TARGET_URL=s3://s3-eu-west-1.amazonaws.com/<my_bucket>/<my_dir> \
   -e AWS_ACCESS_KEY_ID=<my_key_id> \
   -e AWS_SECRET_ACCESS_KEY=<my_secret_key> \
     camptocamp/conplicity
```

## Environment variables

### CONPLICITY_FULL_IF_OLDER_THAN

Perform a full backup if an incremental backup is requested, but the latest full backup in the collection is older than the given time. Defaults to 15D.

### CONPLICITY_REMOVE_OLDER_THAN

Delete all backup sets older than the given time. Defaults to 30D.

### CONPLICITY_VOLUMES_BLACKLIST

Comma separated list of named volumes to blacklist.

### DUPLICITY_DOCKER_IMAGE

The image to use to launch duplicity. Default is `camptocamp/duplicity:latest`.

### DUPLICITY_TARGET_URL

Target URL passed to duplicity.
The hostname and the name of the volume to backup
are added to the path as directory levels.

### PUSHGATEWAY_URL

The Url of a Prometheus pushgateway to push metrics to.

### S3 credentials

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

### Swift credentials

- SWIFT_USERNAME
- SWIFT_PASSWORD
- SWIFT_AUTHURL
- SWIFT_TENANTNAME
- SWIFT_REGIONNAME

## Controlling backup parameters

The parameters used to backup each volume can be fine-tuned using volume labels (requires Docker 1.11.0 or greater):

- `io.conplicity.ignore=true` ignores the volume
- `io.conplicity.full_if_older_than=<value>` sets the time period after which a full backup is performed. Defaults to the `CONPLICITY_FULL_IF_OLDER_THAN` environment variable value


## Providers


Conplicity detects automatically the kind of data that is stored on a volume and adapts its backup strategy to it. The following providers and associated strategies are currently supported:

* PostgreSQL: Run `pg_dumpall` before backup
* MySQL: Run `mysqldump` before backup
* OpenLDAP: Run `slapcat` before backup
* Default: Backup volume data as is

