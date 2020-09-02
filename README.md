# OpenShift Projected Resource "CSI DRIVER"

The Work In Progress implementation for [this OpenShift Enhancement Proposal](https://github.com/openshift/enhancements/blob/master/enhancements/cluster-scope-secret-volumes/csi-driver-host-injections.md),
this repository borrows from this reference implementation:

- [CSI Hostpath Driver](https://github.com/kubernetes-csi/csi-driver-host-path)
- [Node Driver Registrar Sidecar Container](https://github.com/kubernetes-csi/node-driver-registrar)

As part of forking these into this repository, function not required for the projected resources scenarios have 
been removed, and the images containing these commands are built off of RHEL based images instead of the non-Red Hat
images used upstream.

And will in the near future cherry pick overlapping CSI API implementation logic from:

- [Kubernetes-Secrets-Store-CSI-Driver](https://github.com/kubernetes-sigs/secrets-store-csi-driver)

As the enhancement proposal reveals, this is not a fully compliant CSI Driver implementation.  This repository
solely provides the minimum amounts of the Kubernetes / CSI contract to achive the goals stipulated in the 
Enhancement proposal.

## Current status with respect to the enhancement propsal

**NOT FULLY IMPLEMENTED**

The latest commit of the master branch solely introduces both the `Share` CRD and the `projectedresoure.storage.openshift.io`
API group and version `v1alpha1`.  

A controller exists for watching this new CRD.

The reference to the `share` object in the `volumeAttributes` in a declared CSI volume within a `Pod` is used to 
fuel a `SubectAccessReview` check.  The `ServiceAccount` for the `Pod` must have `get` access to the `Share` in
order for the referenced `ConfigMap` and `Secret` to be mounted in the `Pod`.

 


## Deployment
The easiest way to test the Hostpath driver is to run the `deploy.sh`.

```
# deploy csi projectedresource driver
$ deploy/deploy.sh
```

You should see an output similar to the following printed on the terminal showing the application of rbac rules and the
result of deploying the hostpath driver, external provisioner, external attacher and snapshotter components. Note that the following output is from Kubernetes 1.17:

```shell
deploying hostpath components
   deploy/hostpath/00-namespace.yaml
kubectl apply -f deploy/hostpath/00-namespace.yaml
namespace/csi-driver-projected-resource created
   deploy/hostpath/01-service-account.yaml
kubectl apply -f deploy/hostpath/01-service-account.yaml
serviceaccount/csi-driver-projected-resource-plugin created
   deploy/hostpath/03-cluster-role-binding.yaml
kubectl apply -f deploy/hostpath/03-cluster-role-binding.yaml
clusterrolebinding.rbac.authorization.k8s.io/projected-resource-privileged created
   deploy/hostpath/csi-hostpath-driverinfo.yaml
kubectl apply -f deploy/hostpath/csi-hostpath-driverinfo.yaml
csidriver.storage.k8s.io/csi-driver-projected-resource.openshift.io created
   deploy/hostpath/csi-hostpath-plugin.yaml
kubectl apply -f deploy/hostpath/csi-hostpath-plugin.yaml
service/csi-hostpathplugin created
daemonset.apps/csi-hostpathplugin created
```

## Run example application and validate

First, let's validate the deployment.  Ensure all expected pods are running for the driver plugin, which in a 
3 node OCP cluster will look something like:

```shell
$ kubectl get pods -n csi-driver-projected-resource
NAME                       READY   STATUS    RESTARTS   AGE
csi-hostpathplugin-c7bbk   2/2     Running   0          23m
csi-hostpathplugin-m4smv   2/2     Running   0          23m
csi-hostpathplugin-x9xjw   2/2     Running   0          23m
```

Next, let's start up the simple test application.  From the root directory, deploy from the `./examples` directory the 
application `Pod`, along with the associated test namespace, `Share`, `ClusterRole`, and `ClusterRoleBinding` definitions
needed to illustrate the mounting of one of the API types (in this instance a `ConfigMap` from the `openshift-config`
namespace) into the `Pod`:

```shell
$ kubectl apply -f ./examples
namespace/my-csi-app-namespace created
clusterrole.rbac.authorization.k8s.io/projected-resource-my-share created
clusterrolebinding.rbac.authorization.k8s.io/projected-resource-my-share created
share.projectedresource.storage.openshift.io/my-share created
pod/my-csi-app created
```

Ensure the `my-csi-app` comes up in `Running` state.

Finally, if you want to validate the volume, inspect the application pod `my-csi-app`:

```shell
$ kubectl describe pods/my-csi-app
Name:         my-csi-app
Namespace:    csi-driver-projected-resource
Priority:     0
Node:         ip-10-0-163-121.us-west-2.compute.internal/10.0.163.121
Start Time:   Wed, 05 Aug 2020 14:23:57 -0400
Labels:       <none>
Annotations:  k8s.v1.cni.cncf.io/network-status:
                [{
                    "name": "",
                    "interface": "eth0",
                    "ips": [
                        "10.129.2.16"
                    ],
                    "default": true,
                    "dns": {}
                }]
              k8s.v1.cni.cncf.io/networks-status:
                [{
                    "name": "",
                    "interface": "eth0",
                    "ips": [
                        "10.129.2.16"
                    ],
                    "default": true,
                    "dns": {}
                }]
              openshift.io/scc: node-exporter
Status:       Running
IP:           10.129.2.16
IPs:
  IP:  10.129.2.16
Containers:
  my-frontend:
    Container ID:  cri-o://cf4cd4f202d406153e3a067f6f6926ae93dd9748923a5116b2e2ee27e00d33e6
    Image:         busybox
    Image ID:      docker.io/library/busybox@sha256:400ee2ed939df769d4681023810d2e4fb9479b8401d97003c710d0e20f7c49c6
    Port:          <none>
    Host Port:     <none>
    Command:
      sleep
      1000000
    State:          Running
      Started:      Wed, 05 Aug 2020 14:24:03 -0400
    Ready:          True
    Restart Count:  0
    Environment:    <none>
    Mounts:
      /data from my-csi-volume (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from default-token-xbsjd (ro)
Conditions:
  Type              Status
  Initialized       True 
  Ready             True 
  ContainersReady   True 
  PodScheduled      True 
Volumes:
  my-csi-volume:
    Type:              CSI (a Container Storage Interface (CSI) volume source)
    Driver:            csi-driver-projected-resource.openshift.io
    FSType:            
    ReadOnly:          false
    VolumeAttributes:  <none>
  default-token-xbsjd:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  default-token-xbsjd
    Optional:    false
QoS Class:       BestEffort
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute for 300s
                 node.kubernetes.io/unreachable:NoExecute for 300s
Events:
  Type    Reason          Age        From                                                 Message
  ----    ------          ----       ----                                                 -------
  Normal  Scheduled       <unknown>                                                       Successfully assigned csi-driver-projected-resource/my-csi-app to ip-10-0-163-121.us-west-2.compute.internal
  Normal  AddedInterface  28m        multus                                               Add eth0 [10.129.2.16/23]
  Normal  Pulling         28m        kubelet, ip-10-0-163-121.us-west-2.compute.internal  Pulling image "busybox"
  Normal  Pulled          28m        kubelet, ip-10-0-163-121.us-west-2.compute.internal  Successfully pulled image "busybox" in 3.626604306s
  Normal  Created         28m        kubelet, ip-10-0-163-121.us-west-2.compute.internal  Created container my-frontend
  Normal  Started         28m        kubelet, ip-10-0-163-121.us-west-2.compute.internal  Started container my-frontend
```


## Confirm openshift-config ConfigMap data is present

This current version of the driver as POC also watches the `ConfigMaps` and `Secrets` in the `openshift-config` 
namespace and places that data in the provide `Volume` as well.

To verify, go back into the `Pod` named `my-csi-app` and list the contents:

  ```shell
  $ kubectl exec  -n my-csi-app-namespace -it my-csi-app /bin/sh
  / # ls -lR /data
  / # exit
  ```

You should see contents like:

```shell
ls -lR /data
ls -lR /data
/data:
total 0
drwxr-xr-x    2 root     root            60 Sep  2 19:38 configmaps

/data/configmaps:
total 4
-rw-r--r--    1 root     root           970 Sep  2 19:38 openshift-config:openshift-install
/ # 
```

To facilitate validation of the contents, including post-creation updates, the data is currently 
stored as formatted `json`.

If you want to try other `ConfigMaps` or a `Secret`, first clear out the existing application:

```shell
$ oc delete -f ./examples 
``` 

And the edit `./examples/02-csi-share.yaml` and change the `backingResource` stanza to point to the item 
you want to share, and then re-run `oc apply -f ./examples`.
