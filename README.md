# Barometer Collectd Operator

A Kubernetes operator for collectd daemon which collects system and application performance metrics periodically and provides mechanisms to store the values in a variety of ways.

## Introduction

Collectd gathers metrics from various sources, e.g. the operating system, applications, logfiles and external devices, and stores this information or makes it available over the network. Those statistics can be used to monitor systems, find performance bottlenecks (i.e. performance analysis) and predict future system load (i.e. capacity planning). 

## Project: Barometer 

The ability to monitor the Network Function Virtualization Infrastructure (NFVI) where VNFs are in operation will be a key part of Service Assurance within an NFV environment, in order to enforce SLAs or to detect violations, faults or degradation in the performance of NFVI resources so that events and relevant metrics are reported to higher level fault management systems. If fixed function appliances are going to be replaced by virtualized appliances the service levels, manageability and service assurance needs to remain consistent or improve on what is available today.

As such, the NFVI needs to support the ability to monitor:


### Deploy Barometer Operator

The `deploy` directory contains the manifests needed to properly install the
Operator.

 Build and push the barometer-operator image to a public registry such as quay.io(make sure to mark it as public in setting of your quay repo after pushing.)
 ```
$ operator-sdk build quay.io/YOUR-NAMESPACE/barometer-operator
$ docker push quay.io/YOUR-NAMESPACE/barometer-operator

```

 Update the operator manifest to use the built image name

 ```
$ sed -i 's|REPLACE_IMAGE|quay.io/YOUR-NAMESPACE/barometer-operator|g' deploy/operator.yaml
 On OSX use:
$ sed -i "" 's|REPLACE_IMAGE|quay.io/YOUR-NAMESPACE/barometer-operator|g' deploy/operator.yaml

```
Create the service account for the operator.

```
$ kubectl create -f deploy/service_account.yaml
```

Create the RBAC role and role-binding that grants the permissions
necessary for the operator to function.

```
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
```

Deploy the CRD to the cluster that defines the Collectd resource.

```
$ kubectl create -f deploy/crds/collectd.opnfv.org_barometers_crd.yaml
```
You will be able to confirm that the new CRD has been registered in the cluster and you can review its details.

```
$ kubectl get crd
$ kubectl describe crd barometers.collectd.opnfv.org
```


Next, deploy the operator into the cluster.

```
$ kubectl create -f deploy/operator.yaml
```
 The default controller will watch for Collectd objects and create a pod for each CR

This step will create a pod on the Kubernetes cluster for the barometer Operator.
Observe the `barometer-operator` pod and verify it is in the running state.

```
$ kubectl get pods -l name=barometer-operator
```

If for some reason, the pod does not get to the running state, look at the
pod details to review any event that prohibited the pod from starting.

```
$ kubectl describe pod -l name=barometer-operator
```

 ```
$ kubectl create -f deploy/crds/collectd.opnfv.org_v1alpha1_barometer_cr.yaml
```

 ```
$ cat <<EOF | kubectl create -f - 
apiVersion:collectd.opnfv.org/v1alpha1
kind: Barometer
metadata:
  name: barometer-collectd-ds
  labels:
    app: barometer
spec:
  # Add fields here
  deploymentPlan: 
    image: opnfv/barometer-collectd
    size: 1
    configname: barometer-config    
```


The operator will create a deployment of  barometer as daemon using default collectd.conf,To make changes to collectd configuration , you may apply configuration via configmap , where configmap name is matched with deployment.configname .
which can be viewed by running .

```
kubectl get barometer -A
```



To  add/modifying collectd configurations:
```
kubectl apply -f examples/configmap.yaml

```





## Development

This Operator is built using the [Operator SDK](https://github.com/operator-framework/operator-sdk). Follow the [Quick Start](https://github.com/operator-framework/operator-sdk) instructions to checkout and install the operator-sdk CLI.

Local development may be done with [minikube](https://github.com/kubernetes/minikube) or [minishift](https://www.okd.io/minishift/).

#### Source Code

Clone this repository to a location on your workstation such as `$GOPATH/src/github.com/ORG/REPO`. Navigate to the repository and install the dependencies.

```
$ cd $GOPATH/src/github.com/ORG/REPO/barometer-operator
```

#### Run Operator Locally

Ensure the service account, role, role bindings and CRD are added to  the local cluster.

```
$ kubectl create -f deploy/service_account.yaml
$ kubectl create -f deploy/role.yaml
$ kubectl create -f deploy/role_binding.yaml
$ kubectl create -f deploy/crds/collectd.opnfv.org_barometers_crd.yaml
```

Start the operator locally for development.

```
$ export OPERATOR_NAME=barometer-operator
$ operator-sdk  run local
```

Create a  resource to observe and test your changes.

```console
 cat <<EOF | kubectl create -f -
apiVersion: collectd.barometer.com/v1alpha1
kind: Collectd
metadata:
  name: barometer-collectd-ds
spec:
  # Add fields here
  deploymentPlan: 
    image: opnfv/barometer-collectd
    size: 1
    configname: barometer-config    
EOF
```

As you make local changes to the code, restart the operator to enact the changes.

### Clean up
```
# Cleanup
$ kubectl delete -f deploy/crds/collectd.opnfv.org_v1alpha1_barometer_cr.yaml
$ kubectl delete -f deploy/operator.yaml
$ kubectl delete -f deploy/role.yaml
$ kubectl delete -f deploy/role_binding.yaml
$ kubectl delete -f deploy/service_account.yaml
$ kubectl delete -f deploy/crds/collectd.opnfv.org_barometers_crd.yaml

```

#### Build

The Makefile will do the dependency check, operator-sdk generate k8s, run local test, and finally the operator-sdk build. Please ensure any local docker server is running.

```
make
```

#### Test

Before submitting PR, please test your code. 

File or local validation.
```
$ make test
```

Cluster-based test. 
Ensure there is a cluster running before running the test and replace image in operator.yml

```
$ make cluster-test
```

## Manage the operator using the Operator Lifecycle Manager

Ensure the Operator Lifecycle Manager is installed in the local cluster.  By default, the `catalog-source.sh` will install the operator catalog resources in `operator-lifecycle-manager` namespace.  You may also specify different namespace where you have the Operator Lifecycle Manager installed.

```
$ ./hack/catalog-source.sh [namespace]
$ oc apply -f deploy/olm-catalog/barometer-operator/0.1.0/catalog-source.yaml
```
