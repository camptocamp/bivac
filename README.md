Conplicity
==========

[![Docker Pulls](https://img.shields.io/docker/pulls/camptocamp/conplicity.svg)](https://hub.docker.com/r/camptocamp/conplicity/)
[![By Camptocamp](https://img.shields.io/badge/by-camptocamp-fb7047.svg)](http://www.camptocamp.com)


conplicity lets you backup all your named docker volumes using duplicity.


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

### DUPLICITY_DOCKER_IMAGE

The image to use to launch duplicity. Default is `camptocamp/duplicity:latest`.

### DUPLICITY_TARGET_URL

Target URL passed to duplicity.
The hostname and the name of the volume to backup
are added to the path as directory levels.

### S3 credentials

- AWS_ACCESS_KEY_ID
- AWS_SECRET_ACCESS_KEY

### Swift credentials

- SWIFT_USERNAME
- SWIFT_PASSWORD
- SWIFT_AUTHURL
- SWIFT_TENANTNAME
- SWIFT_REGIONNAME
