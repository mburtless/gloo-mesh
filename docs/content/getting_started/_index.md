---
title: "Getting Started"
menuTitle: Getting Started
description: Get started with Gloo Mesh Open Source
weight: 10
---

Welcome to Gloo Mesh, the open-source, multicluster, multimesh management plane. Gloo Mesh simplifies service mesh operations and lets you manage service meshes across multiple clusters from a centralized management plane.

To quickly get started with a Gloo Mesh Open Source demo environment, you can use Kind to run local Kubernetes clusters in Docker.

## Before you begin

Install the following tools:

* [Docker Desktop](https://www.docker.com/products/docker-desktop). In **Preferences > Resources > Advanced**, ensure that [at least 10 CPUs and 8 GiB of memory are available](https://kind.sigs.k8s.io/docs/user/quick-start/#settings-for-docker-desktop).
* Version 1.7, 1.8, or 1.9 of [`istioctl`](https://istio.io/latest/docs/setup/getting-started/#download), the Istio command line tool.
* [`meshctl`]({{< versioned_link_path fromRoot="/setup/meshctl_cli_install/" >}}), the Gloo Mesh command line tool for bootstrapping Gloo Mesh, registering clusters, describing configured resources, and more.
* [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start#installation), a tool for running local Kubernetes clusters by using Docker containers.
* [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command line tool. Download the `kubectl` version that is within one minor version of your Kubernetes cluster. For example, this demo environment creates clusters that run Kubernetes version 1.17.17.

## Set up the Gloo Mesh demo environment

1. Use `kind` to set up the Gloo Mesh demonstration environment. This command completes the following operations:
   * Creates two clusters named `mgmt-cluster` and `remote-cluster`.
   * Installs Istio in both clusters.
   * Deploys the Istio `Bookinfo` sample application to the `bookinfo` namespace in each cluster.
   * Installs Gloo Mesh in `mgmt-cluster`, which serves as the _management cluster_ in this setup.
   * Registers both clusters with Gloo Mesh so that both clusters are _remote clusters_ in this setup.

   ```
   meshctl demo istio-multicluster init
   ```

   Example output for the creation of `mgmt-cluster` and `remote-cluster`:
   ```
   Creating cluster mgmt-cluster with ingress port 32001
   Creating cluster "mgmt-cluster" ...
    ✓ Ensuring node image (kindest/node:v1.17.17) 🖼 
    ✓ Preparing nodes 📦  
    ✓ Writing configuration 📜 
    ✓ Starting control-plane 🕹️ 
    ✓ Installing CNI 🔌 
    ✓ Installing StorageClass 💾 
   Set kubectl context to "kind-mgmt-cluster"
   You can now use your cluster with:

   kubectl cluster-info --context kind-mgmt-cluster

   ...

   Creating cluster remote-cluster with ingress port 32000
   Creating cluster "remote-cluster" ...
    ✓ Ensuring node image (kindest/node:v1.17.17) 🖼 
    ✓ Preparing nodes 📦  
    ✓ Writing configuration 📜 
    ✓ Starting control-plane 🕹️ 
    ✓ Installing CNI 🔌 
    ✓ Installing StorageClass 💾 
   Set kubectl context to "kind-remote-cluster"
   You can now use your cluster with:

   kubectl cluster-info --context kind-remote-cluster
   ```

2. Save the context for each cluster as an environment variable.
   ```shell
   export CONTEXT_1=kind-mgmt-cluster
   export CONTEXT_2=kind-remote-cluster
   ```

3. List the pods in the `gloo-mesh` namespace of `mgmt-cluster`.
   ```shell
   kubectl --context $CONTEXT_1 get po -n gloo-mesh
   ```

   Example output:
   ```
   NAME                          READY   STATUS    RESTARTS   AGE
   cert-agent-7d79bf9f44-8pl25   1/1     Running   0          3m28s
   discovery-7bbb5bdc6c-rh59b    1/1     Running   0          3m32s
   networking-7fb9847967-vqbt5   1/1     Running   0          3m32s
   ```

4. To verify that the Gloo Mesh setup is complete, check the status of the Gloo Mesh pods and connectivity of Gloo Mesh agents in the managed clusters.
   ```shell
   meshctl check server
   ```

   Example output:
   ```
   Gloo Mesh Management Cluster Installation

   🟢 Gloo Mesh Pods Status
   +----------------+------------+-------------------------------+-----------------+
   |    CLUSTER     | REGISTERED | DASHBOARDS AND AGENTS PULLING | AGENTS PUSHING  |
   +----------------+------------+-------------------------------+-----------------+
   | mgmt-cluster   | true       |                             2 |               1 |
   +----------------+------------+-------------------------------+-----------------+
   | remote-cluster | true       |                             2 |               1 |
   +----------------+------------+-------------------------------+-----------------+

   🟢 Gloo Mesh Agents Connectivity

   Management Configuration

   🟢 Gloo Mesh CRD Versions

   🟢 Gloo Mesh Networking Configuration Resources
   ```

Your Gloo Mesh management cluster is set up, and your remote clusters are registered. Your demo evironment is now ready to go!

## Next steps

Check out the following guides to explore more of Gloo Mesh's capabilities in your demo environment.
* [Mesh discovery]({{% versioned_link_path fromRoot="/guides/discovery_intro/" %}}): Enable Gloo Mesh to automatically discover both service mesh installations on registered clusters by using control plane and sidecar discovery.
* [Traffic policies]({{% versioned_link_path fromRoot="/guides/traffic_policy/" %}}): Configure traffic policies, including properties such as timeouts, retries, CORS, and header manipulation, for services that are associated with a service mesh installation.
* [Federated trust and identity]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}): Unify the root identity between multiple service mesh installations so that any intermediates are signed by the same root certificate authority and end-to-end mTLS between clusters and destinations can be established.

To set up and manage the configuration of Gloo Mesh on your existing clusters, follow the steps in the [Setup documentation]({{% versioned_link_path fromRoot="/setup/" %}}).

{{% notice tip %}}
Running into issues? For troubleshooting help, join the [Solo.io Slack workspace](https://slack.solo.io).
{{% /notice %}}

## Clean up

If you no longer need your Gloo Mesh demo environment, you can run the following command to clean up the `mgmt-cluster` and `remote-cluster` clusters and all associated resources.

```shell
meshctl demo istio-multicluster cleanup
```