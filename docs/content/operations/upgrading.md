---
title: Upgrading Gloo Mesh Enterprise
menuTitle: Upgrading Gloo Mesh Enterprise
weight: 100
description: Best practices for upgrading Gloo Mesh Enterprise
---

Upgrade Gloo Mesh Enterprise from version 1.1.0 or beta tag to version 1.1.1 or later. If you are on version 1.0 or earlier, uninstall and re-install Gloo Mesh Enterprise, as described in one of the [setup guides]({{% versioned_link_path fromRoot="/getting_started/#deploying-gloo-mesh" %}}).

{{< notice warning >}}
During the upgrade, the data plane continues to run, but you might not be able to modify the configurations through the management plane. Because zero downtime is not guaranteed, try testing the upgrade in a staging environment before upgrading your production environment.
{{< /notice >}}

## Upgrade from 1.1.0 to 1.1.1 or later

1.  In your terminal, set environment variables for your current and target Gloo Mesh Enterprise versions to use in the subsequent steps.

    ```shell
    # Set the 1.1.x version to upgrade to
    export UPGRADE_VERSION=1.1.1

    # Specify the values of your Gloo Mesh installation
    export NAMESPACE=gloo-mesh
    export RELEASE_NAME=gloo-mesh-enterprise
    export GLOO_MESH_LICENSE_KEY=<your-key>
    ```

2.  Set your Kubernetes context to the management cluster, and upgrade the Helm installation.

    ```shell
    helm upgrade $RELEASE_NAME --namespace $NAMESPACE \
    https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise/gloo-mesh-enterprise-$UPGRADE_VERSION.tgz \
    --set licenseKey=$GLOO_MESH_LICENSE_KEY \
    --reuse-values \
    ```

3.  For each remote cluster, set your Kubernetes context to the cluster, and upgrade the Helm installation.
    ```shell
    helm upgrade enterprise-agent --namespace $NAMESPACE https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent/enterprise-agent-$UPGRADE_VERSION.tgz
    ```

4.  Set the Kubernetes context to the management cluster and check that your Gloo Mesh resources are in a healthy state. Refer to the [Troubleshooting Guide]({{% versioned_link_path fromRoot="/operations/troubleshooting" %}}) for more details.

    ```shell
    meshctl check server
    ```

## Upgrade from 1.1.0 beta tag to 1.1.1 or later

1.  In your terminal, set environment variables for your current and target Gloo Mesh Enterprise versions to use in the subsequent steps.

    ```shell
    # Set the 1.1.x version to upgrade to
    export UPGRADE_VERSION=1.1.1

    # Specify the values of your Gloo Mesh installation
    export NAMESPACE=gloo-mesh
    export RELEASE_NAME=gloo-mesh-enterprise
    export GLOO_MESH_LICENSE_KEY=<your-key>
    ```
2.  Set your Kubernetes context to the management cluster.

    ```shell
    kubectl config set-context <cluster-name>
    ```
3.  Update the Helm repo for Gloo Mesh Enterprise.
    ```shell
    helm repo update
    ```
4.  Pull the latest Gloo Mesh Enterprise Helm chart files.
    ```shell
    helm pull gloo-mesh-enterprise/gloo-mesh-enterprise --version $UPGRADE_VERSION --untar
    ```
5.  Apply the Gloo Mesh Enterprise CRDs on your management cluster. 
    ```shell
    kubectl apply -f gloo-mesh-enterprise/charts/enterprise-networking/charts/gloo-mesh-crds/crds/
    ```
6.  Clean up the Kubernetes resources that the Gloo Mesh Helm chart does not manage. If you keep these resources, the resources might cause problems with future upgrade steps. If the resources do not exist, proceed to the next step.
    ```shell
    kubectl delete rolebinding -n gloo-mesh enterprise-networking-test enterprise-agent-test
    kubectl delete pod -n gloo-mesh enterprise-agent-test
    kubectl delete crd ratelimiterserverconfigs.networking.enterprise.mesh.gloo.solo.io
    ```
7.  Repeat the previous steps to change the context, apply the CRDs, and clean up the Kubernetes resources on all of your remote clusters.
8.  Set your Kubernetes context to the management cluster, and upgrade the Helm installation.

    ```shell
    helm upgrade $RELEASE_NAME --namespace $NAMESPACE \
    https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise/gloo-mesh-enterprise-$UPGRADE_VERSION.tgz \
    --set licenseKey=$GLOO_MESH_LICENSE_KEY \
    --set relayClientAuthority="enterprise-networking.gloo-mesh" \
    ```

9.  For each remote cluster, set your Kubernetes context to the cluster, and upgrade the Helm installation.
    ```shell
    helm upgrade enterprise-agent --namespace $NAMESPACE https://storage.googleapis.com/gloo-mesh-enterprise/enterprise-agent/enterprise-agent-$UPGRADE_VERSION.tgz
    ```

10. Set the Kubernetes context to the management cluster and check that your Gloo Mesh Enterprise resources are in a healthy state.

    ```shell
    meshctl check server
    ```