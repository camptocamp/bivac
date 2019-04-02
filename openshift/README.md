# Bivac Openshift Template
Template to install bivac in openshift without helm

* Install template:
```bash
oc create -f bivac-template.yaml
```
* Instanciate template:
```bash
oc process bivac-template.yaml \
  -p BIVAC_TARGET_URL=s3:s3.amazonaws.com/<BUCKET NAME> \
  -p AWS_ACCESS_KEY_ID=<AWS ACCESS KEY> \
  -p AWS_SECRET_ACCESS_KEY=<AWS SECRET KEY> \
  -p RESTIC_PASSWORD=<RESTIC PASSWORD> \
  -p NAMESPACE=<OPENSHIFT NAMESPACE> | oc create -f -
```
* Delete bivac and all related resources:
```bash

oc delete clusterrolebinding bivac
oc delete clusterrole bivac
oc delete -n bivac-test serviceaccount bivac
oc delete -n bivac-test deploymentconfig bivac
oc delete -n bivac-test secret bivac
oc delete -n bivac-test service bivac
oc delete -n bivac-test route bivac
oc delete namespace <OPENSHIFT NAMESPACE>
```
