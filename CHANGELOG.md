# [0.26.1](https://github.com/camptocamp/bivac/releases/tag/0.26.1) (2017-05-26)

* Rclone:

  - Pin camptocamp/rclone:1.33-1 as newer images changes environment variable API


# [0.26.0](https://github.com/camptocamp/bivac/releases/tag/0.26.0) (2017-01-03)

* Metrics:

  - Use volume as primary key in metrics, avoids getting all metrics (PR #117, fix #115)

# [0.25.6](https://github.com/camptocamp/bivac/releases/tag/0.25.6) (2016-10-17)

* Bugfix:

  - Fix lint issue

# [0.25.5](https://github.com/camptocamp/bivac/releases/tag/0.25.5) (2016-10-17)

* Features:

  - Retry on errors (fix #112)

# [0.25.4](https://github.com/camptocamp/bivac/releases/tag/0.25.4) (2016-10-17)

* Bugfix:

  - Do not crash when GetMetrics fails (fix #108)

# [0.25.3](https://github.com/camptocamp/bivac/releases/tag/0.25.3) (2016-09-26)

* Dependencies:

  - Switch to docker/docker since docker/engine-api is deprecated

* Bugfix:

  - Override p.SetVolumeBackupDir() for all providers (fix #107, really fix #103)

# [0.25.2](https://github.com/camptocamp/bivac/releases/tag/0.25.2) (2016-09-05)

* Build:

  - Build with Go 1.7 in Travis CI

* Bugfix:

  - Check exit code in PrepareBackup() (fix #105)

# [0.25.1](https://github.com/camptocamp/bivac/releases/tag/0.25.1) (2016-08-31)

* Bugfix:

  - Drop databases before recreating (fix #104)
  - Call p.GetBackupDir(), not p.backupDir (fix #103)

# [0.25.0](https://github.com/camptocamp/bivac/releases/tag/0.25.0) (2016-08-22)

## Warning:

We don't use a specific path separator when using the duplicity engine for swift anymore, as the bug that prevented to backup to a pseudo-folder is fixed in duplicity 0.7.08.

# [0.24.5](https://github.com/camptocamp/bivac/releases/tag/0.24.5) (2016-08-18)

* Bugfix:

  - Really take the last chainEndTime element (fix #96)

# [0.24.4](https://github.com/camptocamp/bivac/releases/tag/0.24.4) (2016-08-17)

* Bugfix:

  - Fix c.ImageInspectWithRaw() call for new version of docker/engine-api

# [0.24.3](https://github.com/camptocamp/bivac/releases/tag/0.24.3) (2016-08-17)

* Bugfix:

  - Take the last chainEndTime element (fix #96)

* Tests:

  - Fix some tests

* Quality:

  - Add go report badge

# [0.24.2](https://github.com/camptocamp/bivac/releases/tag/0.24.2) (2016-08-10)

* Bugfix:

  - Fix UpdateEvent() not updating event

* Logging:

  - Make debug less verbose

* Tests:

  - Add tests

# [0.24.1](https://github.com/camptocamp/bivac/releases/tag/0.24.1) (2016-08-10)

* Bugfix:

  - Fix wrong condition checking on Event#Equals()

* Tests:

  - Add tests for Event#Equals()

# [0.24.0](https://github.com/camptocamp/bivac/releases/tag/0.24.0) (2016-08-10)

* Internals:

  - Linting
  - Avoid asserting config params as strings (fix #89)
  - Use a sub-struct for sub-config in Volumes (fix #90)
  - Use net/url for targetURL, fix RClone URLs (fix #92)

* Configuration:

  - Add more volume configuration (fix #91)
  - Use a global targetURL (fix #93)

* Tests:

  - Fix volume tests
  - Fix util tests

* Documentation:

  - Update README
  - Improve and update manpage


# [0.23.0](https://github.com/camptocamp/bivac/releases/tag/0.23.0) (2016-08-09)

* Internals:

  - Use reflect to compute Volume config (fix #87)

* Configuration:

  - Add a .bivac.overrides file as a replacement for Docker labels (fix #87)

* Metrics:

  - Add backupStartTime and backupEndTime metrics for volumes

# [0.22.0](https://github.com/camptocamp/bivac/releases/tag/0.22.0) (2016-08-08)

* Internals:

  - Fix volume labels fetching (fix #83)
  - Refactor metrics, using Metric and Event structs

* Metrics:

  - Change metric names according to Prometheus docs
  - Fetch existing metrics from Prometheus gateway (fix #84)

* Testing:

  - Add metrics to coverage

* Documentation:

  - Add return codes to README
  - Document engines in README

# [0.21.0](https://github.com/camptocamp/bivac/releases/tag/0.21.0) (2016-07-28)

* Internals:

  - Rename lib/ as handler/
  - Add NewBivac() method
  - Don't use panic
  - Inspect volumes in GetVolumes()
  - Propagate errors instead of crashing (fix #65)
  - Add metrics.go

# [0.20.0](https://github.com/camptocamp/bivac/releases/tag/0.20.0) (2016-07-28)

* Features:

  - Output version in logs
  - Make backup engine pluggable
  - Add RClone engine (fix #71, #76)
  - Add option to get host name from Rancher metadata (fix #73)

* Internals:

  - Check errors in PushToPrometheus
  - Read response from Prometheus

# [0.13.3](https://github.com/camptocamp/bivac/releases/tag/0.13.3) (2016-07-05)

* Bugfix:

  - Blacklist lost+found (fix #68)
  - Wait for container to exit before retrieving logs (fix #69)

# [0.13.2](https://github.com/camptocamp/bivac/releases/tag/0.13.2) (2016-07-05)

* Bugfix:

  - Fix logging in exec create
  - Create new Docker client for PrepareBackup (fix #70)

* Testing:

  - Add docker image testing

# [0.13.1](https://github.com/camptocamp/bivac/releases/tag/0.13.1) (2016-06-30)

* Bugfix:

  - Fix image pulling (fix #61)
  - Support "none" as value for last full backup date
    (avoid crashing on first backup)

* Internals:

  - Change some panic errors into fatal to avoid traces

* Testing:

  - Add more tests

# [0.13.0](https://github.com/camptocamp/bivac/releases/tag/0.13.0) (2016-06-29)

* Breaking changes:

  - JSON_OUTPUT renamed to BIVAC_JSON_OUTPUT

* Internals:

  - Reorganize packages in lib/ (fix #60)

* Bugfix:

  - Do not intent to remove links in removeContainer()

* Testing:

  - Test with verbose
  - Add more tests
  - Add coverage

# [0.12.0](https://github.com/camptocamp/bivac/releases/tag/0.12.0) (2016-06-28)

* Features:

  - Log as JSON with -j (fix #56)
  - Use logs with fields (fix #57)
  - Add stdout in debug output

* Internals:
  - Rewrite using github.com/docker/engine-api (fix #58)

* Bugfix:

  - Fix stdout redirection bug (fix #51)

# [0.11.0](https://github.com/camptocamp/bivac/releases/tag/0.11.0) (2016-06-28)

* Features:

  - Add --no-verify to skip volume backup verification
  - Add backupExitCode new metric

* Internals:

  - Lint code

* Testing:

  - Add testing (fix #55)

# [0.10.2](https://github.com/camptocamp/bivac/releases/tag/0.10.2) (2016-06-23)

* Bugfix:

  - Check errors better

# [0.10.1](https://github.com/camptocamp/bivac/releases/tag/0.10.1) (2016-06-21)

* Bugfix:

  - Specify env-delim for `BIVAC_VOLUMES_BLACKLIST`

# [0.10.0](https://github.com/camptocamp/bivac/releases/tag/0.10.0) (2016-06-16)

* Internals:

  - Refactor various parameters
  - Set version from Makefile + git

# [0.9.0](https://github.com/camptocamp/bivac/releases/tag/0.9.0) (2016-06-16)

* Features:

  - Add loglevel flag and environment variable
  - Add version flag

* Internals:

  - Refactor getVolumeLabel in util

# [0.8.1](https://github.com/camptocamp/bivac/releases/tag/0.8.1) (2016-06-15)

* Features:

  - Add flags to the command line (fixes [#49](https://github.com/camptocamp/bivac/issues/49))

* Bugfix:

  - Send metrics to pushgateway only if slice is not empty (fixes [#47](https://github.com/camptocamp/bivac/issues/47))

* Internals:

  - Use github.com/jessevdk/go-flags to manage flags and environment variables

# [0.7.1](https://github.com/camptocamp/bivac/releases/tag/0.7.1) (2016-05-23)

* Bugfix:

  - Add --ssh-options="-oStrictHostKeyChecking=no" to duplicity parameters (fixes [#32](https://github.com/camptocamp/bivac/issues/32))
  - Add missing LICENSE file (fixes [#30](https://github.com/camptocamp/bivac/issues/30))

# [0.8.0](https://github.com/camptocamp/bivac/releases/tag/0.8.0) (2016-06-14)

* Features:

  - Push metrics to Prometheus Pushgateway defined by PUSHGATEWAY_URL
  - Generate verifyExitCode metric (fixes [#38](https://github.com/camptocamp/bivac/issues/38))
  - Generate lastBackup and lastFullBackup metrics (fixes [#40](https://github.com/camptocamp/bivac/issues/40))

* Internals:

  - Fix Makefile to detect changes in all source files (fixes [#33](https://github.com/camptocamp/bivac/issues/33))
  - Refactor BackupVolume (fixes [#44](https://github.com/camptocamp/bivac/issues/44))

# [0.7.1](https://github.com/camptocamp/bivac/releases/tag/0.7.1) (2016-05-23)

* Bugfix:

  - Add --ssh-options="-oStrictHostKeyChecking=no" to duplicity parameters (fixes [#32](https://github.com/camptocamp/bivac/issues/32))
  - Add missing LICENSE file (fixes [#30](https://github.com/camptocamp/bivac/issues/30))

# [0.7.0](https://github.com/camptocamp/bivac/releases/tag/0.7.0) (2016-05-03)

* Breaking changes:

  - environment variable `FULL_IF_OLDER_THAN` renamed to `BIVAC_FULL_IF_OLDER_THAN` (fixes [#28](https://github.com/camptocamp/bivac/issues/28))
  - environment variable `REMOVE_OLDER_THAN` renamed to `BIVAC_REMOVE_OLDER_THAN` (fixes [#28](https://github.com/camptocamp/bivac/issues/28))

* Features:

  - Persist duplicity cache again (fixes [#12](https://github.com/camptocamp/bivac/issues/12))
  - Do not backup duplicity cache (fixes [#16](https://github.com/camptocamp/bivac/issues/16))
  - Remove temporary volumes (fixes [#23](https://github.com/camptocamp/bivac/issues/23))
  - Add support for removing old backups (fixes [#4](https://github.com/camptocamp/bivac/issues/4))
  - Launch duplicity cleanup after backup (fixes [#19](https://github.com/camptocamp/bivac/issues/19))
  - Add support for volumes blacklisting with environment (fixes [#21](https://github.com/camptocamp/bivac/issues/21))

* Internals:

  - Pass `--name vol.Name` to duplicity (fixes [#17](https://github.com/camptocamp/bivac/issues/17))
  - Refactor provider code with a `PrepareBackup()` interface method (fixes [#24](https://github.com/camptocamp/bivac/issues/24))
  - Refactor launching duplicity container into handler (fixes [#26](https://github.com/camptocamp/bivac/issues/26))


# [0.6.2](https://github.com/camptocamp/bivac/releases/tag/0.6.2) (2016-05-03)

* Bugfix:

  - Revert persist duplicity cache

# [0.6.1](https://github.com/camptocamp/bivac/releases/tag/0.6.1) (2016-04-21)

* Bugfix:

  - Fix for docker < 1.11
  - Code refactoring
  - Persist duplicity cache

# [0.6.0](https://github.com/camptocamp/bivac/releases/tag/0.6.0) (2016-04-21)

* Features:

  - Add providers for PostgreSQL, MySQL, OpenLDAP, and Default backup strategies
  - Refactor code
  - Update github.com/fsouza/go-dockerclient

# [0.5.0](https://github.com/camptocamp/bivac/releases/tag/0.5.0) (2016-04-15)

* Features

  - Use github.com/caarlos0/env to manage environment variables cleanly
  - Support volume labels to tune backup behavior (close #2)
  - Add `FULL_IF_OLDER_THAN` environment variable

* Build chain

  - Update github.com/fsouza/go-dockerclient


# [0.4.0](https://github.com/camptocamp/bivac/releases/tag/0.4.0) (2016-04-08)

* Features

  - Only pull image when it's not already present

* Build chain

  - Add Godeps
  - Automatic build on Travis CI

* Docker image

  - Reduce docker image size by using scratch

# [0.3.1](https://github.com/camptocamp/bivac/releases/tag/0.3.1) (2016-04-06)

* Internals: 

  - Lint with `golint` and `goimports`

# [0.3.0](https://github.com/camptocamp/bivac/releases/tag/0.3.0) (2016-04-05)

* Features:

  - Add `DUPLICITY_DOCKER_IMAGE` environment variable

* Internals:

  - Improve code organization

# [0.2.0](https://github.com/camptocamp/bivac/releases/tag/0.2.0) (2016-04-05)

* Features:

  - Pull image before starting backup

* Internals:

  - Use implicit composition for `*docker.Client` in `Bivac` struct

# [0.1.0](https://github.com/camptocamp/bivac/releases/tag/0.1.0) (2016-04-04)

* Initial release
