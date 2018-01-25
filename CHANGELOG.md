# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## Unreleased

### Added
- Adding node metrics. Data is fetched from Kubelet and kube-state-metrics.

### Fixed
- Integration stops on KSM or Kubelet connection error, instead of continuing.

## 1.0.0-beta3

### Fixed
- Fix debug log level when verbose. Some parts of the code didn't log debug information.

### Changed
- `updatedAt` metric was renamed to `podsUpdated`.
- `cpuUsedCores` has been divided by 10^9, to show actual cores instead of nanocores.
- Update configurable timeout flag using it to connect to kubelet and kube-state-metrics.

## 1.0.0-beta2

### Added
- Metrics for unscheduled Pods.

### Fixed
- Fix format of inherited labels. Remove unnecessary prefix `label_` included by kube-state-metrics.
- Fix labels inheritance. Labels weren't propagating between "entities" correctly.

## 1.0.0-beta

### Added
- Initial version reporting metrics about Namespaces, Deployments, ReplicaSets,
  Pods and Containers. This data is fetched from two different sources: Kubelet
  and kube-state-metrics.
