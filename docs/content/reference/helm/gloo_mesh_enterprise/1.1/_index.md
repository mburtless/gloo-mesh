
---
title: "v1.1.0"
description: Reference for Helm values. 
weight: 2
---

The following pages provide Helm value reference documentation for various Gloo Mesh components, including:

1. **Open source Gloo Mesh**: the OSS version of Gloo Mesh
2. **Enterprise Networking (enterprise only)**: The management plane of Gloo Mesh Enterprise, deployed on the management cluster.
3. **Enterprise Agent (enterprise only)**: The agent of Gloo Mesh Enterprise, deployed on each remote cluster.
4. **RBAC Webhook (enterprise only)**: The Kubernetes webhook that enforces Gloo Mesh Enterprise's role-based API.
5. **Gloo Mesh UI (enterprise only)**: The UI for Gloo Mesh Enterprise, which includes the Dashboard and Redis subcharts.

Note that when you provide Helm values for the [bundled Gloo Mesh Enterprise chart](https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise),
values for each subchart must be prefixed accordingly:

1. Values for the RBAC Webhook must be prefixed with "rbac-webhook".
2. Values for Enterprise Networking must be prefixed with "enterprise-networking".
3. Values for the Gloo Mesh UI must be prefixed with "gloo-mesh-ui".
  - Values for the Dashboard (a subchart of the Gloo Mesh UI chart) must be prefixed with "dashboard"
  - Values for the Redis Dashboard (a subchart of the Gloo Mesh UI chart) must be prefixed with "redis-dashboard"


The following is an example of how to set values for subcharts:


> helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise \
> --namespace gloo-mesh \
> --set licenseKey=${GLOO_MESH_LICENSE_KEY}  \
> --set rbac-webhook.enabled=true
> --set enterprise-networking.enterpriseNetworking.floatingUserId=true \
> --set gloo-mesh-ui.dashboard.floatingUserId.floatingUserId=true \
> --set gloo-mesh-ui.redis-dashboard.redisDashboard.floatingUserId=true

{{% children description="true" %}}
