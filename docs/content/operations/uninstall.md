---
title: "Uninstall Gloo Mesh"
description: Deregister clusters and uninstall the Gloo Mesh management components
weight: 40
---

If you no longer need your Gloo Mesh environment, you can uninstall Gloo Mesh from your management and remote clusters.

**Before you begin**, save the name and context of each cluster in environment variables. Save the names and contexts for subsequent remote clusters as needed, such as `REMOTE_CONTEXT_2`, and so on.
```shell
export MGMT_CLUSTER=<management-cluster-name>
export MGMT_CONTEXT=<management-cluster-context>
export REMOTE_CLUSTER=<remote-cluster-name>
export REMOTE_CONTEXT=<remote-cluster-context>
```

## Deregister remote clusters

To deregister a cluster, you must uninstall the `cert-agent` that runs on the remote cluster and the corresponding `KubernetesCluster` resource that exists on the management cluster.

1. Uninstall the `agent-crds` and `cert-agent` that runs on the remote cluster.
   ```shell script
   meshctl cluster deregister community \
     --mgmt-context $MGMT_CONTEXT \
     --remote-context $REMOTE_CONTEXT \
     $REMOTE_CLUSTER
   ```

   Example output:
   ```
   Deregistering cluster: cluster-2
   Finished uninstalling release agent-crds
   Finished uninstalling release cert-agent
   Successfully deregistered cluster: cluster-2
   ```

2. Delete the Custom Resource Definitions (CRDs) that were installed on the remote cluster during registration.
   ```shell script
   for crd in $(kubectl get crd --context $REMOTE_CONTEXT | grep mesh.gloo | awk '{print $1}'); do kubectl --context $REMOTE_CONTEXT delete crd $crd; done
   ```

3. Delete the `gloo-mesh` namespace.
   ```shell
   kubectl --context $REMOTE_CONTEXT delete namespace gloo-mesh
   ```

4. Repeat these steps for each cluster that is registered with Gloo Mesh. For example, if you ran the management components in a cluster that was also registered, repeat these steps for the `MGMT_CLUSTER` and specify the `MGMT_CONTEXT`. If you registered multiple remote clusters, repeat these steps for each remote cluster.

## Uninstall management components

Uninstall the Gloo Mesh management componets from the management cluster.

1. Uninstall the Gloo Mesh management plane components.
   ```shell script
   meshctl uninstall --kubecontext $MGMT_CONTEXT
   ```

   Example output:
   ```
   Uninstalling Helm chart
   Finished uninstalling release gloo-mesh
   ```

   {{% notice note %}}
   If you [installed GlooMesh with Kubernetes resources directly]({{< versioned_link_path fromRoot="/setup/community_installation/#installing-with-kubectl-apply" >}}), you can uninstall the management components by running `meshctl install community --dry-run | kubectl delete -f -`.
   {{% /notice %}}

2. Delete the Custom Resource Definitions (CRDs) that were installed on the remote cluster during registration.
   ```shell script
   for crd in $(kubectl get crd --context $REMOTE_CONTEXT | grep mesh.gloo | awk '{print $1}'); do kubectl --context $REMOTE_CONTEXT delete crd $crd; done
   ```

3. Delete the `gloo-mesh` namespace.
   ```shell
   kubectl --context $REMOTE_CONTEXT delete namespace gloo-mesh
   ```

## Optional: Uninstall Bookinfo and Istio

Optionally uninstall Bookinfo and Istio from each remote cluster.

1. If you installed Bookinfo, the Istio sample application, run the following commands to uninstall its resources.
   ```shell script
   # Remove sidecar injection label from the default namespace
   kubectl --context $REMOTE_CONTEXT label namespace default istio-injection-
   # Remove Bookinfo application components for all versions less than v3
   kubectl --context $REMOTE_CONTEXT delete -f https://raw.githubusercontent.com/istio/istio/$ISTIO_VERSION/samples/bookinfo/platform/kube/bookinfo.yaml -l 'app,version notin (v3)'
   # Remove all Bookinfo service accounts
   kubectl --context $REMOTE_CONTEXT delete -f https://raw.githubusercontent.com/istio/istio/$ISTIO_VERSION/samples/bookinfo/platform/kube/bookinfo.yaml -l 'account'
   # Remove ingress gateway configuration for accessing Bookinfo
   kubectl --context $REMOTE_CONTEXT delete -f https://raw.githubusercontent.com/istio/istio/$ISTIO_VERSION/samples/bookinfo/networking/bookinfo-gateway.yaml
   ```

2. Uninstall Istio and delete the `istio-system` namespace.
   ```shell script
   istioctl --context $REMOTE_CONTEXT x uninstall --purge
   kubectl --context $REMOTE_CONTEXT delete namespace istio-system
   ```

3. Repeat these steps for each cluster that was registered with Gloo Mesh. For example, if you ran the management components in a cluster that was also registered, repeat these steps for the `MGMT_CONTEXT`. If you registered multiple remote clusters, repeat these steps for each remote cluster.

## Optional: Clean up Kind clusters

If you followed the Kind cluster setup, you can optionally delete those clusters.
```shell
kind delete clusters cluster-1 cluster-2
```