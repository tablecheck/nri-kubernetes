# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).


## Unreleased

### Kube State Metrics
- Initial version: Includes Metrics data for Kubernetes pods, containers, replicasets, namespace, deployments. Metrics are fetched from [kube-state-metrics service](https://github.com/kubernetes/kube-state-metrics)
- Includes *prometheus* parser, which is used to parse data received from kube-state-metrics

### Kubelet
- Initial version: Includes Metrics data for:
  - network and errors of Kubernetes pods
  - memory and CPU usage of Kubernetes containers
- Second version: self-discovery of kubelet

### Merged version
- Initial, simple merging of KSM and Kubelet integrations.