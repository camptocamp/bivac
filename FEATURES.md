## Client


To connect to a remote Bivac manager:

```shell
$ BIVAC_REMOTE_SERVER=127.0.0.1:8182 bivac info
```

### `info`

Get informations regarding the Bivac manager

```shell
$ bivac info
Version: 2.0.0
Started at: 2019-01-01 19:00:00
Orchestrator: Kubernetes
Address: 0.0.0.0:8182
Managed volumes: 20
```

### `volumes`

Retrieve informations regarding the volumes detected by Bivac.

```shell
$ bivac volumes
ID           NAME              STATUS
00001        foo               backed up
00002        bar               ignored
```

```shell
$ bivac volumes foo
ID           NAME        STATUS      LAST BACKUP            SIZE
00001        foo         backed up   2019-01-01 21:00:00    10Gb 
```

### `backups [VOLUME_NAME]`

List volume's backups.

```shell
$ bivac backups foo
ID        DATE                    PATH
1a54e2f   2019-01-01 20:00:00     /data
```

### `restore`

Restore a volume content from a backup.

```shell
$ bivac restore foo 1a54e2f (--old-host node1.srv)
Restoration successfully completed.
```
