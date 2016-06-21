# [0.10.1](https://github.com/camptocamp/conplicity/releases/tag/0.10.1) (2016-06-21)

* Bugfix:

  - Specify env-delim for `CONPLICITY_VOLUMES_BLACKLIST`

# [0.10.0](https://github.com/camptocamp/conplicity/releases/tag/0.10.0) (2016-06-16)

* Internals:

  - Refactor various parameters
  - Set version from Makefile + git

# [0.9.0](https://github.com/camptocamp/conplicity/releases/tag/0.9.0) (2016-06-16)

* Features:

  - Add loglevel flag and environment variable
  - Add version flag

* Internals:

  - Refactor getVolumeLabel in util

# [0.8.1](https://github.com/camptocamp/conplicity/releases/tag/0.8.1) (2016-06-15)

* Features:

  - Add flags to the command line (fixes [#49](https://github.com/camptocamp/conplicity/issues/49))

* Bugfix:

  - Send metrics to pushgateway only if slice is not empty (fixes [#47](https://github.com/camptocamp/conplicity/issues/47))

* Internals:

  - Use github.com/jessevdk/go-flags to manage flags and environment variables

# [0.7.1](https://github.com/camptocamp/conplicity/releases/tag/0.7.1) (2016-05-23)

* Bugfix:

  - Add --ssh-options="-oStrictHostKeyChecking=no" to duplicity parameters (fixes [#32](https://github.com/camptocamp/conplicity/issues/32))
  - Add missing LICENSE file (fixes [#30](https://github.com/camptocamp/conplicity/issues/30))

# [0.8.0](https://github.com/camptocamp/conplicity/releases/tag/0.8.0) (2016-06-14)

* Features:

  - Push metrics to Prometheus Pushgateway defined by PUSHGATEWAY_URL
  - Generate verifyExitCode metric (fixes [#38](https://github.com/camptocamp/conplicity/issues/38))
  - Generate lastBackup and lastFullBackup metrics (fixes [#40](https://github.com/camptocamp/conplicity/issues/40))

* Internals:

  - Fix Makefile to detect changes in all source files (fixes [#33](https://github.com/camptocamp/conplicity/issues/33))
  - Refactor BackupVolume (fixes [#44](https://github.com/camptocamp/conplicity/issues/44))

# [0.7.1](https://github.com/camptocamp/conplicity/releases/tag/0.7.1) (2016-05-23)

* Bugfix:

  - Add --ssh-options="-oStrictHostKeyChecking=no" to duplicity parameters (fixes [#32](https://github.com/camptocamp/conplicity/issues/32))
  - Add missing LICENSE file (fixes [#30](https://github.com/camptocamp/conplicity/issues/30))

# [0.7.0](https://github.com/camptocamp/conplicity/releases/tag/0.7.0) (2016-05-03)

* Breaking changes:

  - environment variable `FULL_IF_OLDER_THAN` renamed to `CONPLICITY_FULL_IF_OLDER_THAN` (fixes [#28](https://github.com/camptocamp/conplicity/issues/28))
  - environment variable `REMOVE_OLDER_THAN` renamed to `CONPLICITY_REMOVE_OLDER_THAN` (fixes [#28](https://github.com/camptocamp/conplicity/issues/28))

* Features:

  - Persist duplicity cache again (fixes [#12](https://github.com/camptocamp/conplicity/issues/12))
  - Do not backup duplicity cache (fixes [#16](https://github.com/camptocamp/conplicity/issues/16))
  - Remove temporary volumes (fixes [#23](https://github.com/camptocamp/conplicity/issues/23))
  - Add support for removing old backups (fixes [#4](https://github.com/camptocamp/conplicity/issues/4))
  - Launch duplicity cleanup after backup (fixes [#19](https://github.com/camptocamp/conplicity/issues/19))
  - Add support for volumes blacklisting with environment (fixes [#21](https://github.com/camptocamp/conplicity/issues/21))

* Internals:

  - Pass `--name vol.Name` to duplicity (fixes [#17](https://github.com/camptocamp/conplicity/issues/17))
  - Refactor provider code with a `PrepareBackup()` interface method (fixes [#24](https://github.com/camptocamp/conplicity/issues/24))
  - Refactor launching duplicity container into handler (fixes [#26](https://github.com/camptocamp/conplicity/issues/26))


# [0.6.2](https://github.com/camptocamp/conplicity/releases/tag/0.6.2) (2016-05-03)

* Bugfix:

  - Revert persist duplicity cache

# [0.6.1](https://github.com/camptocamp/conplicity/releases/tag/0.6.1) (2016-04-21)

* Bugfix:

  - Fix for docker < 1.11
  - Code refactoring
  - Persist duplicity cache

# [0.6.0](https://github.com/camptocamp/conplicity/releases/tag/0.6.0) (2016-04-21)

* Features:

  - Add providers for PostgreSQL, MySQL, OpenLDAP, and Default backup strategies
  - Refactor code
  - Update github.com/fsouza/go-dockerclient

# [0.5.0](https://github.com/camptocamp/conplicity/releases/tag/0.5.0) (2016-04-15)

* Features

  - Use github.com/caarlos0/env to manage environment variables cleanly
  - Support volume labels to tune backup behavior (close #2)
  - Add `FULL_IF_OLDER_THAN` environment variable

* Build chain

  - Update github.com/fsouza/go-dockerclient


# [0.4.0](https://github.com/camptocamp/conplicity/releases/tag/0.4.0) (2016-04-08)

* Features

  - Only pull image when it's not already present

* Build chain

  - Add Godeps
  - Automatic build on Travis CI

* Docker image

  - Reduce docker image size by using scratch

# [0.3.1](https://github.com/camptocamp/conplicity/releases/tag/0.3.1) (2016-04-06)

* Internals: 

  - Lint with `golint` and `goimports`

# [0.3.0](https://github.com/camptocamp/conplicity/releases/tag/0.3.0) (2016-04-05)

* Features:

  - Add `DUPLICITY_DOCKER_IMAGE` environment variable

* Internals:

  - Improve code organization

# [0.2.0](https://github.com/camptocamp/conplicity/releases/tag/0.2.0) (2016-04-05)

* Features:

  - Pull image before starting backup

* Internals:

  - Use implicit composition for `*docker.Client` in `Conplicity` struct

# [0.1.0](https://github.com/camptocamp/conplicity/releases/tag/0.1.0) (2016-04-04)

* Initial release
