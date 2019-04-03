# Bivac Openshift Template
Templates to install bivac in openshift/OKD without helm

* Install template:
```bash
oc create -f bivac-template.yaml
```
* Instanciate template directly from file:
```bash
oc process -f bivac-template.yaml \
  -p BIVAC_TARGET_URL=s3:s3.amazonaws.com/<BUCKET NAME> \
  -p AWS_ACCESS_KEY_ID=<AWS ACCESS KEY> \
  -p AWS_SECRET_ACCESS_KEY=<AWS SECRET KEY> \
  -p RESTIC_PASSWORD=<RESTIC PASSWORD> \
  -p NAMESPACE=<BIVAC-MANAGER PROJECT> | oc create -f -
```

This will create a new namespace with a bivac-manager deployment, including all required resources like serviceaccount and secret. Note that you will need to create serviceaccounts & rolebindings for all namespaces in which you wish to backup PVCs. You can use the second file (bivac2-agent.template.yaml) for this:
```bash
oc process -f bivac2-agent.template.yaml -p NAMESPACE=<TARGET NAMESPACE>
```

* To delete bivac and all related resources:
```bash
oc delete -n <BIVAC-MANAGER PROJECT> serviceaccount bivac
oc delete -n <BIVAC-MANAGER PROJECT> deploymentconfig bivac
oc delete -n <BIVAC-MANAGER PROJECT> secret bivac
oc delete -n <BIVAC-MANAGER PROJECT> service bivac
oc delete -n <BIVAC-MANAGER PROJECT> route bivac
oc delete clusterrolebinding bivac
oc delete clusterrole bivac
oc delete namespace <BIVAC-MANAGER PROJECT>
```
