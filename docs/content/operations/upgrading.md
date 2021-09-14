---
title: Upgrading Gloo Mesh
weight: 30
description: Upgrade the version of Gloo Mesh in your management and remote clusters
---

Upgrade the version of Gloo Mesh Open Source in your management and remote clusters. If you are on version 1.1.0 or earlier, [uninstall]({{% versioned_link_path fromRoot="/operations/uninstall/" %}}) and [re-install]({{% versioned_link_path fromRoot="/setup/" %}}) Gloo Mesh.

{{< notice warning >}}
During the upgrade, the data plane continues to run, but you might not be able to modify the configurations through the management plane. Because zero downtime is not guaranteed, try testing the upgrade in a staging environment before upgrading your production environment.
{{< /notice >}}

1.  In your terminal, set environment variables for your current and target Gloo Mesh Open Source versions to use in the subsequent steps.
    ```shell
    # Set the 1.1.x version to upgrade to
    export UPGRADE_VERSION=1.1.1

    # Specify the values of your Gloo Mesh installation
    export NAMESPACE=gloo-mesh
    export RELEASE_NAME=gloo-mesh
    ```

2.  Set your Kubernetes context to the management cluster, and upgrade the Helm installation.
    ```shell
    helm upgrade $RELEASE_NAME --namespace $NAMESPACE \
    https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-$UPGRADE_VERSION.tgz \
    --reuse-values \
    ```

3.  For each remote cluster, set your Kubernetes context to the cluster, and upgrade the Helm installation.
    ```shell
    helm upgrade cert-agent --namespace $NAMESPACE https://storage.googleapis.com/gloo-mesh/cert-agent/cert-agent-$UPGRADE_VERSION.tgz
    ```

4.  Set the Kubernetes context to the management cluster and check that your Gloo Mesh resources are in a healthy state. Refer to the [Troubleshooting Guide]({{% versioned_link_path fromRoot="/operations/troubleshooting" %}}) for more details.
    ```shell
    meshctl check server
    ```