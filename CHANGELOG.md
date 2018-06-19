# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

### Unreleased

### Added
- Add `hostNetwork: true` option to daemonset file. This is a requirement for the Infrastructure Agent to report the proper hostname in New Relic.

### Changed
- Update newrelic-infra.yaml to force our objects to be deployed in `default` namespace.

### 1.0.0-beta2.3

### Added
- Add configurable flag for kube-state-metrics endpoint (only HTTP).
- Add additional label `app` for discovering kube-state-metrics endpoint.

### Changed
- Kubelet discovery process fetches now the nodeName directly from the spec using downward API.

### 1.0.0-beta2.2

### Fixed
- Fix bug in error handling where recoverable errors made the integration to panic.

### 1.0.0-beta2.1

### Added
- Allow direct connection to cAdvisor by specifying the port.

### Fixed
- Call to CAdvisor was failing when Kubelet was secure.

### 1.0.0-beta2.0

### Added
- nodes/metrics resource was added to the newrelic cluster role.

### Changed
- CAdvisor call is now bypassing Kubelet endpoint talking then directoy to CAdvisor port

## 1.0.0-beta1.0

Initial public beta release.

## 1.0.0-alpha5.1

### Changed
- TransformFunc now handles errors.
- Add checks for missing data coming from kube-state-metrics.
- Boolean values have changed from `"true"` and `"false"` to `1` and `0` respectively from the following metrics:
  1. isReady and isScheduled for pods.
  2. isReady for containers.
- Update metrics
  1. `errorCountPerSecond` to `errorsPerSecond` for pods and nodes.
  2. `usageCoreSeconds` to `cpuUsedCoreMilliseconds` for nodes.
  3. `memoryMajorPageFaults` to `memoryMajorPageFaultsPerSecond` for nodes.

### Fixed
- Calculate properly RATE metrics.

## 1.0.0-alpha5

### Added
- TypeGenerator for entities.
- Caching discovered endpoints on disk.
- Implementation of Time-To-Live (TTL) cache expiry functionality.
- Added the concept of Leader and Follower roles.
  - Leader represents the node where Kube State Metrics is installed (so only 1 by cluster).
  - Follower represents any other node.
- Both Follower and Leader call kubelet /pods endpoint in order to get metrics that were previously fetched from KSM.
- Fetch metrics from KSM about pods with status "Pending".
- Prometheus TextToProtoHandleFunc as http.HandlerFunc.
  Useful for serving a Prometheus payload in protobuf format from a plain text reader.
- Both Follower and Leader call kubelet /metrics/cadvisor endpoint in order to fill some missing metrics coming from Kubelet.

### Changed
- Rename `endpoints` package to `client` package.
- Moved a bunch of functions related to `Prometheus` from `ksm` package to `prometheus` one.
- Renamed the recently moved `Prometheus` functions. Removed **Prometheus** word as it is considered redundant.
- Containers objects reported as their own entities (not as part of pod entities).
- NewRelic infra Daemonset updateStrategy set to RollingUpdate in newrelic-infra.yaml.
- Prometheus CounterValue type changed from uint to float64.
- Change our daemonset file to deploy the integration in "default" namespace.
- Prometheus queries now require to use an operator.
- Prometheus Do method now requires a metrics endpoint.

### Removed
- Follower does not call KSM endpoints anymore.
- Config package with default unknown namespace value
- Removed legacy Kubernetes spec files.

### Fixed
- Replace `log.Fatal()` by `log.Panic()` in order to call all defer statements. 
- Skip missing data from /stats/summary endpoint, instead of reporting them as zero values.
- Entities not reported in case of problem with setting their name or type.

## 1.0.0-alpha4

### Added
- Adding node metrics. Data is fetched from Kubelet and kube-state-metrics.
- Adding toleration for the "NoSchedule" taint, so the integration is deployed on all nodes.
- Adding new autodiscovery flow with authentication and authorization mechanisms.

### Removed
- Custom arguments for kubelet and kube-state-metrics endpoints.

### Fixed
- Integration stops on KSM or Kubelet connection error, instead of continuing.

## 1.0.0-alpha3

### Changed
- `updatedAt` metric was renamed to `podsUpdated`.
- `cpuUsedCores` has been divided by 10^9, to show actual cores instead of nanocores.
- Update configurable timeout flag using it to connect to kubelet and kube-state-metrics.

### Fixed
- Fix debug log level when verbose. Some parts of the code didn't log debug information.

## 1.0.0-alpha2

### Added
- Metrics for unscheduled Pods.

### Fixed
- Fix format of inherited labels. Remove unnecessary prefix `label_` included by kube-state-metrics.
- Fix labels inheritance. Labels weren't propagating between "entities" correctly.

## 1.0.0-alpha

### Added
- Initial version reporting metrics about Namespaces, Deployments, ReplicaSets,
  Pods and Containers. This data is fetched from two different sources: Kubelet
  and kube-state-metrics.
