# New Relic Kubernetes Infrastructure Integration

New Relic Kubernetes Infrastructure Integration instruments the container orchestration layer by reporting metrics from Kubernetes objects. It gives you visibility about Kubernetes namespaces, deployments, replica sets, nodes, pods, and containers. Metrics are collected from different sources. 
* [kube-state-metrics service](https://github.com/kubernetes/kube-state-metrics) provides information about state of Kubernetes objects like namespace, replicaset, deployments and pods (when they are not in running state)
* `/stats/summary` kubelet endpoint gives information about network, errors, memory and CPU usage
* `/pods` kubelet endpoint provides information about state of running pods and containers
* `/metrics/cadvisor` cAdvisor endpoint provides missing data that is not included in the previous sources.

Check [documentation](https://docs.newrelic.com/docs/kubernetes-integration-new-relic-infrastructure) in order to find out more how to install and configure the integration, learn what metrics are captured and how to view them.

## Table of Contents

* [Configuration and installation](#configuration-and-installation)
* [Usage](#usage)
* [Integration Development usage](#integration-development-usage)

## Configuration and installation

For installing Kubernetes Infrastructure Integration deploy provided `newrelic-infra` file. This will install required roles and `newrelic` daemon set, which deploys the Infrastructure agent and the Kubernetes Infrastructure Integration.
Firstly check [compatibility and requirements](https://docs.newrelic.com/docs/kubernetes-monitoring-integration#compatibility) and then follow the
[installation steps](https://docs.newrelic.com/docs/kubernetes-monitoring-integration#install).
<!-- TODO: specify link to troubleshooting guide, once is done -->

## Usage

Check how to [find and use data](https://docs.newrelic.com/docs/kubernetes-monitoring-integration#view-data) and description of all [captured data](https://docs.newrelic.com/docs/kubernetes-monitoring-integration#metrics).

## Development

For development process [Skaffold](https://github.com/GoogleCloudPlatform/skaffold) tool and [minikube](https://kubernetes.io/docs/getting-started-guides/minikube) are used.

### Prerequisites

[Skaffold](https://github.com/GoogleCloudPlatform/skaffold#installation) correctly installed.

### Configuration

* Copy the daemonset file `deploy/newrelic-infra.yaml` to `deploy/local.yaml`.
* Edit the file and set the following value as container image: `newrelic/infrastructure-k8s-dev`.

```yaml
  containers:
    - name: newrelic-infra
      image: newrelic/infrastructure-k8s-dev
      resources:
```

* Edit the file and specify the following `CLUSTER_NAME` and `NRIA_LICENSE_KEY` on the `env` section.

 ```yaml
 env:
 - name: "CLUSTER_NAME"	
   value: "<YOUR_CLUSTER_NAME>"
 - name: "NRIA_LICENSE_KEY"
   value: "<YOUR_LICENSE_KEY>"
 ```

### Run

Run `make deploy-dev`. This will compile your integration binary with compatibility for the container OS architecture, build a temporary docker image and finally deploy it to your Minikube. 

Then you can [view your data](#usage) or run the integration standalone. To do so follow the steps:

* Run

```bash
NR_POD_NAME=$(kubectl get pods -l name=newrelic-infra -o jsonpath='{.items[0].metadata.name}')
```
This will retrieve the name of a pod where the Infrastructure agent and Kuberntetes Infrastructure Integration are installed.

* Enter to the pod

```bash
kubectl exec -it $NR_POD_NAME -- /bin/bash
```

* Execute the Kubernetes integration

```bash
/var/db/newrelic-infra/newrelic-integrations/bin/nr-kubernetes -pretty
```

* For running test, use

```bash
make test
```