---
title: Upgrading Gloo Mesh Enterprise
menuTitle: Upgrading Gloo Mesh Enterprise
weight: 100
description: Best practices for upgrading Gloo Mesh Enterprise
---

{{< notice warning >}}
This upgrade process is for Gloo Mesh Enterprise, from v1.1.0-betax to v1.1.0. During the upgrade, the data plane continues to run, but you might not be able to modify the configurations through the management plane. Because zero downtime is not guaranteed, try testing the upgrade in a staging environment before upgrading your production environment.

If you are on a previous version, uninstall and re-install Gloo Mesh Enterprise, as described in one of the [setup guides]({{% versioned_link_path fromRoot="/getting_started/#deploying-gloo-mesh" %}}).
{{< /notice >}}

1\. In your terminal, set environment variables for your current and target Gloo Mesh Enterprise versions to use in the subsequent steps.

```shell
# Set your current version here (Must be a v1.1.0-betax version)
CURRENT_VERSION=v1.1.0-beta35
# Add your desired version here (Must be a v1.1.x version)
UPGRADE_VERSION=v1.1.0
```

2\. Upgrade the Gloo Mesh CRDs on your management cluster and all of your data plane clusters. You can change contexts before applying the CRDs by running `kubectl config set-context <cluster-name>`.
```shell
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/admin.enterprise.mesh.gloo.solo.io_v1alpha1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/discovery.mesh.gloo.solo.io_v1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/multicluster.solo.io_v1alpha1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/networking.enterprise.mesh.gloo.solo.io_v1beta1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/networking.mesh.gloo.solo.io_v1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/observability.enterprise.mesh.gloo.solo.io_v1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/rbac.enterprise.mesh.gloo.solo.io_v1_crds.yaml
kubectl apply -f https://raw.githubusercontent.com/solo-io/gloo-mesh/$UPGRADE_VERSION/install/helm/gloo-mesh-crds/crds/settings.mesh.gloo.solo.io_v1_crds.yaml
```

3\. On each cluster, clean up resources that Helm does not manage. If left behind, they
may cause problems with future upgrade steps. If these resources do not exist, the following
commands will simply be a no-op.
```shell
kubectl delete rolebinding -n gloo-mesh enterprise-networking-test enterprise-agent-test
kubectl delete pod -n gloo-mesh enterprise-agent-test
kubectl delete crd ratelimiterserverconfigs.networking.enterprise.mesh.gloo.solo.io
```

On just the management cluster, clean up resources that Helm doees not manage.
```shell
helm upgrade gloo-mesh --namespace gloo-mesh --reuse-values --set gloo-mesh-ui.enabled=false
```
Then, when reinstalling, set `gloo-mesh-ui.enabled` back to true if desired.

4\. Set your Kubernetes context to the management plane cluster, and upgrade the Helm installation.
```shell
helm upgrade --install gloo-mesh --namespace gloo-mesh \
  'https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise/gloo-mesh-enterprise-$UPGRADE_VERSION.tgz' \
  --set licenseKey=$LICENSE_KEY
  --set relayClientAuthority="enterprise-networking.gloo-mesh"
```

5\. For each data plane cluster, set your Kubernetes context to the cluster, and upgrade the Helm installation.
```shell
helm upgrade --install enterprise-agent --namespace gloo-mesh https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent/enterprise-agent-$UPGRADE_VERSION.tgz
```

6\. Set the Kubernetes context to the management plane cluster and then check that your Gloo Mesh Enterprise resources are in a healthy state. Refer to our
[Troubleshooting Guide]({{% versioned_link_path fromRoot="/operations/troubleshooting" %}}) for more details.
