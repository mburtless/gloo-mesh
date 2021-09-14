---
title: "Prepare clusters"
description: Use existing clusters or deploy Kind clusters for a Gloo Mesh setup
weight: 20
---

Choose existing clusters or deploy clusters locally for your Gloo Mesh setup, and set the contexts for your clusters. For example, you can use a managed Kubernetes environment, such as clusters in Google Kubernetes Engine (GKE) or Amazon Elastic Kubernetes Service (EKS). Alternatively, you can deploy two local clusters by using Kubernetes in Docker (Kind). 

## Using managed clusters

Choose or create two clusters.

Throughout the Gloo Mesh documentation, the name `cluster-1` is used for the _management cluster_, and the name `cluster-2` is used for a _remote cluster_. Save the name and context of each cluster in environment variables. If your clusters have different names, specify those names instead.
```shell
export MGMT_CLUSTER=cluster-1
export MGMT_CONTEXT=<management-cluster-context>
export REMOTE_CLUSTER=cluster-2
export REMOTE_CONTEXT=<remote-cluster-context>
```

## Deploying clusters locally with Kind

Deploy two local clusters by using Kubernetes in Docker (Kind). Note that the clusters use a significant amount of RAM to run Istio and Gloo Mesh, so use a workstation that has a minimum of 16GB of memory.

**Before you begin**, install the following tools:

* [Docker Desktop](https://www.docker.com/products/docker-desktop). In **Preferences > Resources > Advanced**, ensure that [at least 10 CPUs and 8 GB of memory are available](https://kind.sigs.k8s.io/docs/user/quick-start/#settings-for-docker-desktop).
* [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start#installation), a tool for running local Kubernetes clusters by using Docker containers.
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command line tool. Download the `kubectl` version that is within one minor version of your Kubernetes cluster. For example, this demo environment creates clusters that run Kubernetes version 1.21.2.

1. Create `cluster-1`, which will serve as both the _management cluster_ and a _remote cluster_ to be managed by Gloo Mesh in this setup.
   ```shell
   cat <<EOF | kind create cluster --name cluster-1 --image kindest/node:v1.21.2 --config=-
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   nodes:
   - role: control-plane
     extraPortMappings:
     - containerPort: 32001
       hostPort: 32001
       protocol: TCP
     kubeadmConfigPatches:
     - |
       kind: InitConfiguration
   kubeadmConfigPatches:
   - |
     kind: InitConfiguration
     nodeRegistration:
       kubeletExtraArgs:
         authorization-mode: "AlwaysAllow"
   EOF
   ```

2. Create `cluster-2`, which will serve as a _remote cluster_ to be managed by Gloo Mesh in this setup.
   ```shell
   cat <<EOF | kind create cluster --name cluster-2 --image kindest/node:v1.21.2 --config=-
   kind: Cluster
   apiVersion: kind.x-k8s.io/v1alpha4
   nodes:
   - role: control-plane
     extraPortMappings:
     - containerPort: 32000
       hostPort: 32000
       protocol: TCP
     kubeadmConfigPatches:
     - |
       kind: InitConfiguration
   kubeadmConfigPatches:
   - |
     kind: InitConfiguration
     nodeRegistration:
       kubeletExtraArgs:
         authorization-mode: "AlwaysAllow"
   EOF
   ```

3. Save the name and context of each cluster in environment variables.
   ```shell
   export MGMT_CLUSTER=cluster-1
   export MGMT_CONTEXT=kind-cluster-1
   export REMOTE_CLUSTER=cluster-2
   export REMOTE_CONTEXT=kind-cluster-2
   ```

4. Switch to the context for the management cluster, `cluster-1`.
   ```shell
   kubectl config use-context $MGMT_CONTEXT
   ```

## Next Steps
Now that you have two clusters ready, you can [install the Gloo Mesh management components]({{% versioned_link_path fromRoot="/setup/community_installation" %}}).