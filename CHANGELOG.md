# [[0.7.0](https://github.com/camptocamp/conplicity/releases/tag/0.7.0)](https://github.com/camptocamp/conplicity/releases/tag/[0.7.0](https://github.com/camptocamp/conplicity/releases/tag/0.7.0)) (2016-05-XX)

* Features:

  - Persist duplicity cache again (fixes #12)
  - Do not backup duplicity cache (fixes #16)
  - Remove temporary volumes (fixes #23)
  - Add support for removing old backups (fixes #4)
  - Launch duplicity cleanup after backup (fixes #19)

* Internals:

  - Pass `--name vol.Name` to duplicity (fixes #17)
  - Refactor provider code with a `PrepareBackup()` interface method (fixes #24)
  - Refactor launching duplicity container into handler (fixes #26)


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
