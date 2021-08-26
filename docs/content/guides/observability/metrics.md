---
title: Metrics
menuTitle: Metrics
description: Guide on Gloo Mesh's metrics features.
weight: 30
---

This guide describes how to get started with Gloo Mesh Enterprise's out of the box metrics suite.

{{% notice note %}} Gloo Mesh Enterprise is required for this feature. {{% /notice %}}

{{% notice note %}} This feature currently only supports Istio meshes. {{% /notice %}}

## Before you begin

This guide assumes the following:

* Istio is [installed on both `cluster-1` and `cluster-2`]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
    * Istio is configured according to what's described in "Environment Prerequisites" below
    * `istio-system` is the root namespace for both Istio deployments
* The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace

## Environment Prerequisites

### Istio

Each managed Istio control plane must be installed with the following configuration in the [`IstioOperator` manifest](https://istio.io/latest/docs/reference/config/istio.operator.v1alpha1/).

```yaml
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: example-istiooperator
  namespace: istio-system
spec:
  meshConfig:
    defaultConfig:
      envoyMetricsService:
        address: enterprise-agent.gloo-mesh:9977
      proxyMetadata:
        # needed for annotating Gloo Mesh cluster name on envoy requests (i.e. access logs, metrics)
        GLOO_MESH_CLUSTER_NAME: ${gloo-mesh-registered-cluster-name}
  values:
    global:
      # needed for annotating istio metrics with cluster
      multiCluster:
        clusterName: ${gloo-mesh-registered-cluster-name}
```

The `envoyMetricsService` config ensures that all Envoy proxies are configured to emit their metrics to
the Enterprise Agent, which acts as an [Envoy metrics service sink](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/metrics/v3/metrics_service.proto#extension-envoy-stat-sinks-metrics-service).
The Enterprise Agents then forward all received metrics to Enterprise Networking, where metrics across all managed clusters are centralized.

The `multiCluster` config enables Istio collected metrics to be annotated with the Gloo Mesh registered cluster name.
This allows for proper attribution of metrics in multicluster environments, and is particularly important for attributing
requests that cross cluster boundaries.

### Gloo Mesh Enterprise

When installing Gloo Mesh Enterprise, the `metricsBackend.prometheus.enabled` Helm value must be set to true. This can be done by providing
the following argument to `Helm install`, `--set metricsBackend.prometheus.enabled=true`.

This configures Gloo Mesh to install a Prometheus server which comes preconfigured to scrape the centralized metrics from the Enterprise Networking
metrics endpoint.

After installation of the Gloo Mesh management plane into `cluster-1`, you should see the following deployments:

```shell
gloo-mesh      enterprise-networking-69d74c9744-8nlkd               1/1     Running   0          23m
gloo-mesh      prometheus-server-68b58c79f8-rlq54                   2/2     Running   0          23m
```

## Functionality

### Generate Traffic

Before any meaningful metrics are collected, traffic has to be generated in the system.

Port forward the productpage deployment (the productpage workload is
convenient because it makes requests to the other workloads, but any workload of your choice will suffice).

```shell
kubectl -n bookinfo port-forward deploy/productpage-v1 9080
```

Then using a utility like [hey](https://github.com/rakyll/hey), send requests to that destination:

```shell
# send 1 request per second
hey -z 1h -c 1 -q 1 http://localhost:9080/productpage\?u\=normal
```

Note that you may need to wait a few minutes before the metrics are returned from the Gloo Mesh API discussed below.
The metrics need time to propagate from the Envoy proxies to the Gloo Mesh server, and for the Prometheus server to scrape the data from Gloo Mesh.

### Prometheus UI

The Prometheus server comes with a builtin UI suitable for basic metrics querying. You can view it with the following commands:

```shell
# port forward prometheus server
kubectl -n gloo-mesh port-forward deploy/prometheus-server 9090
```

Then open `localhost:9090` in your browser of choice.
Here is a simple promql query to get you started with navigating the collected metrics.
This query fetches the `istio_requests_total` metric (which counts the total number of requests) emitted by the
`productpage-v1.bookinfo.cluster-1` workload's Envoy proxy. You can read more about PromQL in the [official documentation](https://prometheus.io/docs/prometheus/latest/querying/basics/).

```promql
sum(
  increase(
    istio_requests_total{
      gm_workload_ref="productpage-v1.bookinfo.cluster-1",
    }[2m]
  )
) by (
  gm_workload_ref,
  gm_destination_workload_ref,
  response_code,
)
```

## Using a Custom Prometheus Instance

To integrate Gloo Mesh with an existing Prometheus server or other Prometheus-compatible solution, you must disable the default Prometheus server. Then, configure Gloo Mesh and your Prometheus server to communicate with each other.

1\. Set up Gloo Mesh Enterprise to disable the default Prometheus instance and instead read from your custom Prometheus instance's full URL, including the port number. You can include the following `--set` flags in a `helm upgrade` [command](https://helm.sh/docs/helm/helm_upgrade/), or update these fields in your [Helm values configuration file]({{% versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise/" %}}) when you install Gloo Mesh Enterprise.

```
--set enterprise-networking.metricsBackend.prometheus.enabled=false
--set enterprise-networking.metricsBackend.prometheus.url=<URL (with port) to Prometheus server>
```

2\. Configure your Prometheus server to scrape metrics from Gloo Mesh. Although each solution might have a different setup, configure your solution to scrape from the `enterprise-networking.gloo-mesh:9091` endpoint and respect the [Prometheus scrapping annotations](https://github.com/solo-io/gloo-mesh-enterprise/blob/main/enterprise-networking/install/helm/enterprise-networking/templates/deployment.yaml#L32) in the Gloo Mesh deployment.

For example, if you have the [Prometheus Community Chart](https://github.com/prometheus-community/helm-charts/tree/main/charts/prometheus), update the Helm `values.yaml` file as follows to scrape metrics from Gloo Mesh.

```yaml
serverFiles:
  prometheus.yml:
    scrape_configs:
    - job_name: gloo-mesh
      scrape_interval: 15s
      scrape_timeout: 10s
      static_configs:
      - targets:
        - enterprise-networking.gloo-mesh:9091
```


3\. **Optional**: Scrape metrics from the agents on data plane clusters. You might collect these metrics for operational awareness of the system, such as for troubleshooting purposes. Note that these metrics are not rendered in the service graph of the Gloo Mesh UI. To collect these metrics, configure your Prometheus instance to scrape the `enterprise-agent.gloo-mesh:9091/metrics` endpoint on the data plane clusters.
