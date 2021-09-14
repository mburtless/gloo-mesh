---
title: "Register clusters"
description: Register clusters to be managed by Gloo Mesh
weight: 50
---

Register clusters so that Gloo Mesh can identify and manage their service meshes.

## Before you begin

* Install the following CLI tools:
  * [`meshctl`]({{< versioned_link_path fromRoot="/setup/meshctl_cli_install/" >}}), the Gloo Mesh command line tool for bootstrapping Gloo Mesh, registering clusters, describing configured resources, and more.
  * [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command line tool. Download the `kubectl` version that is within one minor version of your Kubernetes cluster.
* [Prepare at least two clusters]({{< versioned_link_path fromRoot="/setup/kind_setup/" >}}) for your Gloo Mesh setup, [install the Gloo Mesh management components]({{< versioned_link_path fromRoot="/setup/community_installation/" >}}) into one cluster, and [install an Istio service mesh]({{< versioned_link_path fromRoot="/setup/installing_istio/" >}}) into one or more clusters.
* Save the name and context of each cluster in environment variables. Throughout the Gloo Mesh documentation, the name `cluster-1` is used for the management cluster, and the name `cluster-2` is used for a remote cluster. If your clusters have different names, specify those names instead.
  ```shell
  export MGMT_CLUSTER=cluster-1
  export MGMT_CONTEXT=<management-cluster-context>
  export REMOTE_CLUSTER=cluster-2
  export REMOTE_CONTEXT=<remote-cluster-context>
  ```

## Register clusters

In order for Gloo Mesh to manage the service mesh in a cluster, you must register the cluster with Gloo Mesh. Registration ensures that Gloo Mesh can identify the remote cluster and has the proper credentials to communicate with the Kubernetes API server in the remote cluster.

The registration process creates a service account, cluster role, and cluster role binding that grants service account the necessary permissions to monitor and make changes to the cluster. For more information about the cluster role definition, see the [reference documentation]({{% versioned_link_path fromRoot="/reference/cluster_role" %}}).

### Register a remote cluster

Start by registering a remote cluster, which is a cluster that runs a service mesh but does not run the Gloo Mesh management components.

1. Set your current context to the management cluster.
   ```shell
   kubectl config use-context $MGMT_CONTEXT
   ```

2. Register the remote cluster. The Kubernetes context for the remote cluster is specified in the `--remote-context` flag. If you use Kind clusters, use the Kind tabs to specify the IP address and port of the Kubernetes API server in the `--api-server-address` flag.
   {{< tabs >}}
   {{< tab name="Kubernetes" codelang="shell" >}}
   meshctl cluster register community $REMOTE_CLUSTER \
     --remote-context $REMOTE_CONTEXT
   {{< /tab >}}
   {{< tab name="Kind (MacOS)" codelang="shell" >}}
   ADDRESS=$(docker inspect $REMOTE_CLUSTER-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

   meshctl cluster register community $REMOTE_CLUSTER \
     --remote-context $REMOTE_CONTEXT \
     --api-server-address ${ADDRESS}:6443
   {{< /tab >}}
   {{< tab name="Kind (Linux)" codelang="shell" >}}
   ADDRESS=$(docker exec "$REMOTE_CLUSTER-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')

   meshctl cluster register community $REMOTE_CLUSTER \
     --remote-context $REMOTE_CONTEXT \
     --api-server-address ${ADDRESS}:6443
   {{< /tab >}}
   {{< /tabs >}}

   Example output:
   ```
   Registering cluster
   Finished installing chart 'agent-crds' as release gloo-mesh:agent-crds
   Finished installing chart 'cert-agent' as release gloo-mesh:cert-agent
   Successfully registered cluster: cluster-2
   ```

3. Verify that the registration is successful by listing the Custom Resources (CRs) in the management cluster. The presence of the remote cluster CR ensures that the management cluster has the necessary permissions to create the service account, cluster role, and cluster role binding in the remote cluster.
   ```shell
   kubectl get kubernetescluster -n gloo-mesh $REMOTE_CLUSTER
   ```

   Example output:
   ```
   NAME             AGE
   cluster-2        68s
   ```

4. Verify that the service account, cluster role, and cluster role binding are created in the remote cluster.
   * Service account: 
      ```shell
      kubectl get sa --context $REMOTE_CONTEXT -n gloo-mesh
      ```

      Example output:
      ```
      NAME             SECRETS   AGE
      cert-agent   1         89s
      cluster-2    1         89s
      default      1         4m9s
      ```
   * Cluster role:
      ```shell
      kubectl get clusterrole --context $REMOTE_CONTEXT gloomesh-remote-access
      ```

      Example output:
      ```
      NAME                     AGE
      gloomesh-remote-access   5m
      ```
   * Cluster role binding:
      ```shell
      kubectl get clusterrolebinding --context $REMOTE_CONTEXT \
        $REMOTE_CLUSTER-gloomesh-remote-access-clusterrole-binding
      ```

      Example output:
      ```
      NAME                                                        AGE
      cluster-2-gloomesh-remote-access-clusterrole-binding        7m
      ```

### Register the management cluster

Next, register the management cluster, which is a cluster that runs a service mesh alongside the Gloo Mesh management components. If you follow a deployment pattern in which the management components are installed in a dedicated cluster that does not run a service mesh, skip this section.

The Kubernetes context for the management cluster is specified in the `--remote-context` flag. If you use Kind clusters, use the Kind tabs to specify the IP address and port of the Kubernetes API server in the `--api-server-address` flag.
{{< tabs >}}
{{< tab name="Kubernetes" codelang="shell" >}}
meshctl cluster register community $MGMT_CLUSTER \
 --remote-context $MGMT_CONTEXT
{{< /tab >}}
{{< tab name="Kind (MacOS)" codelang="shell" >}}
ADDRESS=$(docker inspect $MGMT_CLUSTER-control-plane | jq -r '.[0].NetworkSettings.Networks.kind.IPAddress')

meshctl cluster register community $MGMT_CLUSTER \
 --remote-context $MGMT_CONTEXT \
 --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< tab name="Kind (Linux)" codelang="shell" >}}
ADDRESS=$(docker exec "$MGMT_CLUSTER-control-plane" ip addr show dev eth0 | sed -nE 's|\s*inet\s+([0-9.]+).*|\1|p')

meshctl cluster register community $MGMT_CLUSTER \
 --remote-context $MGMT_CONTEXT \
 --api-server-address ${ADDRESS}:6443
{{< /tab >}}
{{< /tabs >}}

Example output:
```
Registering cluster
Finished installing chart 'agent-crds' as release gloo-mesh:agent-crds
Finished installing chart 'cert-agent' as release gloo-mesh:cert-agent
Successfully registered cluster: cluster-1
```

## More about cluster registration

Here's what happened behind the scenes when you registered a cluster with Gloo Mesh:
* A service account was created in the `gloo-mesh` namespace of the remote cluster.
* The service account's auth token was stored in a secret in the management cluster.
* The Gloo Mesh CSR agent was deployed in the remote cluster.
* Any communication from the Gloo Mesh management components to the remote cluster's Kubernetes API server will use the service account's auth token.

## Next Steps

Now that Gloo Mesh is running in the management cluster and remote clusters are registered with Gloo Mesh, your Gloo Mesh setup is complete! Any service meshes in the registered clusters are now available for configuration by Gloo Mesh.

Check out the following guides to explore more of Gloo Mesh's capabilities.
* [Mesh discovery]({{% versioned_link_path fromRoot="/guides/discovery_intro/" %}}): Enable Gloo Mesh to automatically discover both service mesh installations on registered clusters by using control plane and sidecar discovery.
* [Traffic policies]({{% versioned_link_path fromRoot="/guides/traffic_policy/" %}}): Configure traffic policies, including properties such as timeouts, retries, CORS, and header manipulation, for services that are associated with a service mesh installation.
* [Federated trust and identity]({{% versioned_link_path fromRoot="/guides/federate_identity/" %}}): Unify the root identity between multiple service mesh installations so that any intermediates are signed by the same root certificate authority and end-to-end mTLS between clusters and destinations can be established.
