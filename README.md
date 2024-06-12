# Beaver
## _**Distributed openvswitch controller for kubernetes.**_

> ### I works in combination with Multus CNI
> ### _It needs beaver agent to work with_

## Installation
Install the CRD at the first.

```bash
kubectl apply -f ./crd/ovsnet.yaml
kubectl apply -f ./crd/vni.yaml
```

Now can deploy the beaver controller:

```bash
kubectl apply -f ./manifests/deploymnet.yaml
```