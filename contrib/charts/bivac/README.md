# Helm chart for Bivac

> Backup Interface for Volumes Attached to Containers

## Configuration

The following tables list the configurable parameters of the Bivac chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `image.repository` | Repository for the Bivac image. | `camptocamp/bivac` |
| `image.tag` | Tag of the Bivac image. | `2.2` |
| `image.pullPolicy` | Pull policy for the Bivac image. | `IfNotPresent` |
| `orchestrator` | Orchestrator Bivac will run on. | `kubernetes` |
| `watchAllNamespaces` | Let Bivac backup volumes from all namespaces. | `true` |
| `targetURL` | URL where to Restic should push the backups. This field is required. | `nil` |
| `resticPassword` | Password used by Restic to encrypt the backups. If left empty, a generated one will be used. | `nil` |
| `serverPSK` | Pre-shared key which protect the Bivac server. If left empty, a generated one will be used. | `nil` |
| `extraEnv` | Additional environment variables. | `[]` |
| `service.type` | Bivac server type. | `ClusterIP` |
| `service.port` | Port to expose Bivac. | `8182` |
| `resources` | Resource limits for Bivac. | `{}` |
| `nodeSelector` | Define which Nodes the Pods are scheduled on. | `{}` |
| `tolerations` | If specified, the pod's tolerations. | `[]` |
| `affinity` | Assign custom affinity rules. | `{}` |

