# TODO: mix KSM and Kubelet documents

# New Relic kube-state-metrics Infrastructure Integration

New Relic kube-state-metrics Infrastructure Integration collects the most important data from [kube-state-metrics service](https://github.com/kubernetes/kube-state-metrics) about the health of the Kubernetes objects: namespaces, deployments, replica sets, pod and containers from the Kubernetes cluster.

## Configuration
It is required to have kube-state-metrics service configured (version v1.1.0). Once kube-state-metrics is deployed, check that you have access to the `/metrics` endpoint. Typically, if you deploy kube-state-metrics to the Kubernetes cluster and you want to access the `/metrics` endpoint being inside the cluster, then the endpoint is: `http://localhost:8080/metrics`, but it depends on your configuration. You need to know the `/metrics` endpoint, because if it's different than `http://localhost:8080/metrics` (which is used by default in the kube-state-metrics integration) you need to specify it in the configuration of the integration.

## Installation
* download an archive file for the kube-state-metrics Integration
* extract `kube-state-metrics-definition.yml` and `/bin` directory into `/var/db/newrelic-infra/newrelic-integrations`
* check that you can successfully execute the the binary file `nr-kube-state-metrics`
* extract `kube-state-metrics-config.yml.sample` into `/etc/newrelic-infra/integrations.d`

## Usage
This is the description about how to run the kube-state-metrics Integration with New Relic Infrastructure agent, so it is required to have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

In order to use the kube-state-metrics Integration it is required to configure `kube-state-metrics-config.yml.sample` file. Firstly, rename the file to `kube-state-metrics-config.yml`. Then, depending on your needs, specify all instances that you want to monitor with correct information (`/metrics` endpoint). Once this is done, restart the Infrastructure agent.

You can view your data in Insights by creating your own custom NRQL queries. To do so specify the event type. Available event types:
- K8sNamespaceSample
- K8sDeploymentSample
- K8sReplicasetSample
- K8sPodSample
- K8sContainerSample

## Build & Release

**Note:** Remove sensitive data before moving this integration to the public
repository

We are using Docker Hub as a docker registry and Jenkins to build and push our
Docker image. We are in private beta but we need a public repository to share
with customers, for that reason, we are using a hacky mechanism: using a
meaningless name for the docker image ["ohaik"](https://hub.docker.com/r/newrelic/ohaik/)

* Update integration version number (source code)
* Update daemon set container image (deploy/newrelic-infra.yaml)
* Update CHANGELOG.md with information about the changes
* Create a PR with all the changes in the infra-integrations-beta and make sure that is approved and merged
* Go to Jenkins -> infra-integrations-beta -> infra-integrations-beta-kubernetes-docker-image
* Click on 'Build with parameters' and enter the branch you want to build and the version number of the Docker image
* Once the job is executed, a new version of the image should be uploaded to
  docker hub. You can check it this [link](https://hub.docker.com/r/newrelic/ohaik/)

## Compatibility
New Relic kube-state-metrics Integration is compatible with kube-state-metrics service version: v1.1.0

## Integration Development usage
Assuming that you have the source code and Go tool installed you can build and run the kube-state-metrics Integration locally.
* Go to the directory of the kube-state-metrics integration and build it
```bash
$ make
```
* The command above will execute the tests for the kube-state-metrics integration and build an executable file called `nr-kube-state-metrics` under `bin` directory. Run `nr-kube-state-metrics`:
```bash
$ ./bin/nr-kube-state-metrics
```
* If you want to know more about usage of `./bin/nr-kube-state-metrics` check
```bash
$ ./bin/nr-kube-state-metrics -help
```

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to specific version (if possible) into vendor directory.

New Relic Kubelet Infrastructure Integration
============================================

New Relic Kubelet Infrastructure Integration is the next integration (apart from [kube-state-metrics Infrastructure Integration](https://github.com/newrelic/infra-integrations-beta/tree/master/integrations/kube-state-metrics) for monitoring Kubernetes cluster. Kubelet Infrastructure provides information about:

* network and errors for pods,
* memory and CPU usage for containers.

Data is collected from kubelet endpoint: `/stats/summary`. 

Table of Contents
-----------------

* [Configuration](#configuration)
* [Installation](#installation)
* [Usage](#usage)
* [Integration Development usage](#integration-development-usage)

Configuration
-----------------

Check that you have access to `/stats/summary` endpoint (typically: `http://<node_ip>:10255/stats/`).

Installation
-----------------

* [Install the Infrastructure integrations package.](https://docs.newrelic.com/docs/infrastructure/host-integrations/installation/install-host-integrations-built-new-relic)

* Via the command line, change directory to the integration's folder

```bash
$ cd /etc/newrelic-infra/integrations.d
```

* Create a copy of the sample configuration file by running

```bash
$ sudo cp kubelet-config.yml.sample kubelet-config.yml
```

* Check that you can successfully execute the the binary file `nr-kubelet`

```bash
$ sudo ./var/db/newrelic-infra/newrelic-integrations/bin/nr-kubelet
```

Usage
-----------------

This is the description about how to run the kubelet Integration with New Relic Infrastructure agent. It is required to have the agent installed (see [agent installation](https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/install-infrastructure-linux)).

Configure `kubelet.yml.sample` file. Depending on your needs, specify all instances that you want to monitor with correct information (`/stats/summary` kubelet endpoint). Once this is done, restart the Infrastructure agent.

You can view your data in Insights by creating your own custom NRQL queries. To do so specify the event type. Available event types:

* K8sPodSample
* K8sContainerSample

Integration Development usage
-----------------

Assuming that you have the source code and Go tool installed you can build and run the kubelet Integration locally.

* Go to the directory of the kubelet integration and build it

```bash
$ make
```

* The command above will execute the tests for the kubelet integration and build an executable file called `nr-kubelet` under `bin` directory. Run `nr-kubelet`:

```bash
$ ./bin/nr-kubelet
```

* If you want to know more about usage of `./bin/nr-kubelet` check

```bash
$ ./bin/nr-kubelet -help
```

For managing external dependencies [govendor tool](https://github.com/kardianos/govendor) is used. It is required to lock all external dependencies to specific version (if possible) into vendor directory.
