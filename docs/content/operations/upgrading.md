---
title: Upgrading Gloo Mesh
weight: 30
description: Upgrade the version of Gloo Mesh in your management and remote clusters
---

Upgrade the version of Gloo Mesh Open Source in your management and remote clusters. If you are on version 1.1.0 or earlier, [uninstall]({{% versioned_link_path fromRoot="/operations/uninstall/" %}}) and [re-install]({{% versioned_link_path fromRoot="/setup/" %}}) Gloo Mesh.

{{< notice warning >}}
During the upgrade, the data plane continues to run, but you might not be able to modify the configurations through the management plane. Because zero downtime is not guaranteed, try testing the upgrade in a staging environment before upgrading your production environment.
{{< /notice >}}

## Before you begin

1.  Review the [changelog]({{% versioned_link_path fromRoot="/reference/changelog/" %}}) for new features and fixes.
2.  In your terminal, set environment variables for your current and target Gloo Mesh versions to use in the subsequent steps.
    ```shell
    # Set the 1.1.x version to upgrade to
    export UPGRADE_VERSION=1.1.1

    # Specify the values of your Gloo Mesh installation
    export NAMESPACE=gloo-mesh
    export RELEASE_NAME=gloo-mesh
    ```

## Step 1: Upgrade the CRDs for your management cluster

1.  Set your Kubernetes context to the management cluster.
    ```shell
    kubectl config set-context <cluster-name>
    ```
2.  Add and update the Helm repo for `gloo-mesh`.
    ```shell
    helm repo add gloo-mesh https://storage.googleapis.com/gloo-mesh/gloo-mesh
    helm repo update
    ```
3.  Pull the latest Gloo Mesh Helm chart files for the version.
    ```shell
    helm pull gloo-mesh/gloo-mesh --version $UPGRADE_VERSION --untar
    ```
4.  Apply the Gloo Mesh CRDs on your management cluster. 
    ```shell
    kubectl apply -f gloo-mesh/charts/gloo-mesh-crds/crds/
    ```
5.  Upgrade the Helm installation.

    ```shell
    helm upgrade $RELEASE_NAME --namespace $NAMESPACE \
    https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-$UPGRADE_VERSION.tgz \
    --reuse-values \
    ```

## Step 2: Upgrade the agent CRDs for each remote cluster

1.  Set your Kubernetes context to the remote cluster.
    ```shell
    kubectl config set-context <cluster-name>
    ```
2.  Upgrade the Helm installation for the cert agent.
    ```shell
    helm upgrade cert-agent --namespace $NAMESPACE https://storage.googleapis.com/gloo-mesh/cert-agent/cert-agent-$UPGRADE_VERSION.tgz
    ```
3.  Repeat these steps for each remote cluster.

## Step 3: Verify the Helm installation

1.  From your remote cluster, check the pods in the `gloo-mesh` namespace and note the name of the `cert-agent` pod.
    ```shell
    kubectl get po -n gloo-mesh
    ```
2.  Describe the pod and check the `Image` field. Notice that the pod runs the image of the version that you upgraded to.
    ```shell
    kubectl describe po -n gloo-mesh cert-agent-<pod-id>
    ```

    Example output:
    {{< highlight yaml "hl_lines=6" >}}
    Name:         cert-agent-649cf5bdf5-7895f
    Namespace:    gloo-mesh
    Containers:
    cert-agent:
        Container ID:  containerd://b9192312c15f2c554180818e8dced5c0d5657af2b602eb4082e5a28d9c0c0300
        Image:         gcr.io/gloo-mesh/cert-agent:1.1.1
    {{< /highlight >}}

3.  Set the Kubernetes context to the management cluster and check that your Gloo Mesh resources are in a healthy state. Refer to the [Troubleshooting Guide]({{% versioned_link_path fromRoot="/operations/troubleshooting" %}}) for more details.
    ```shell
    meshctl check server
    ```
