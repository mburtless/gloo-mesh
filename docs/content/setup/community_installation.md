---
title: "Install Gloo Mesh"
description: Install the Gloo Mesh management components
weight: 30
---

Use `meshctl` or Helm to install the Gloo Mesh management components into a cluster. The control plane components can be installed in a dedicated management cluster, or in a cluster that runs a service mesh.

**Dedicated management cluster**: Install Gloo Mesh in a dedicated cluster that does not run a service mesh. This cluster serves only as the management cluster in your setup, and is not a registered cluster. This deployment pattern is recommended for production-level setups, as shown in this diagram

![Management componenets installed in a dedicated management cluster]({{% versioned_link_path fromRoot="/img/gloomesh-3clusters.png" %}})

**Service mesh cluster**: Install Gloo Mesh in a cluster that also runs a service mesh. This cluster serves as both the management cluster and as a registered cluster. This deployment pattern uses less resources and is useful for getting started with and exploring Gloo Mesh's capabilities. The guides throughout the Gloo Mesh Open Source documentation set follow this deployment pattern to install Gloo Mesh into a service mesh cluster, as shown in this diagram.

![Management components installed in a registered service mesh cluster]({{% versioned_link_path fromRoot="/img/gloomesh-2clusters.png" %}})

## Before you begin

* Install the following CLI tools:
  * [`meshctl`]({{< versioned_link_path fromRoot="/setup/meshctl_cli_install/" >}}), the Gloo Mesh command line tool for bootstrapping Gloo Mesh, registering clusters, describing configured resources, and more.
  * [`kubectl`](https://kubernetes.io/docs/tasks/tools/#kubectl), the Kubernetes command line tool. Download the `kubectl` version that is within one minor version of your Kubernetes cluster.
* [Prepare at least two clusters]({{< versioned_link_path fromRoot="/setup/kind_setup/" >}}) for your Gloo Mesh setup.
* Save the context for your management cluster in an environment variable, and set the context to the management cluster.
  ```shell
  export MGMT_CONTEXT=<management-cluster-context>
  kubectl config use-context $MGMT_CONTEXT
  ```

## Install Gloo Mesh

Install Gloo Mesh in the management cluster by using `meshctl`, Kubernetes resources, or Helm.

### Using `meshctl`

Use the Gloo Mesh command line tool to install the management components.
```shell
meshctl install community --kubecontext $MGMT_CONTEXT
```

Example output:
```
Installing Helm chart
Finished installing chart 'gloo-mesh' as release gloo-mesh:gloo-mesh
```

### Using `kubectl apply`

To directly use the Kubernetes resources for installation, you can use the `--dry-run` flag in the `meshctl install` command to output YAML resource files. Then, you can use the files in automation or apply the files directly to yhe management cluster by using `kubectl apply`.

```shell
meshctl install community --dry-run
```

{{% notice note %}}
The `--dry-run` flag outputs the entire YAML for the resources, but does not properly order resources. For example, race conditions can occur between Custom Resource Definitions (CRDs) that are registered and any Custom Resources (CRs) that are created. If this error occurs, you can simply re-apply the resources by using `kubectl apply`.
{{% /notice %}}

### Using Helm

Use [Helm version 3](https://helm.sh/docs/intro/install/) to install the Gloo Mesh Open Source Helm charts.

1. Locally add and update the Gloo Mesh Helm repository.
   ```shell
   helm repo add gloo-mesh https://storage.googleapis.com/gloo-mesh/gloo-mesh
   helm repo update
   ```

2. Optional: List the Helm chart information and values.
   ```shell
   helm show all gloo-mesh/gloo-mesh
   ```

3. Install Gloo Mesh into the `gloo-mesh` namespace of the management cluster.
   ```shell
   helm install gloo-mesh gloo-mesh/gloo-mesh --namespace gloo-mesh --create-namespace
   ```

#### Helm chart customization

{{% notice note %}} This feature is available only for Gloo Mesh Open Source version v1.1.0-beta12 or later.{{% /notice %}}

If you need to tailor the installation manifests to specific requirements, you can customize the Helm release by passing in a Helm value file. You can specify a `deploymentOverrides` field for the `discovery` sub-component, and a `serviceOverrides` field for the `networking` sub-component. The values of these fields are merged with the default deployment and service fields via a [Helm library chart function](https://github.com/helm/charts/blob/master/incubator/common/templates/_util.tpl).

1. Create a `values.yaml` file, and specify your deployment or service override values. For example, the following `values.yaml` file adds a custom label to the discovery pod and replaces the service account used by the networking pod.
   ```yaml
   discovery:
     deploymentOverrides:
       spec:
         template:
           metadata:
             annotations:
               test: new-annotation
   networking:
     serviceOverrides:
       spec:
         template:
           spec:
             serviceAccountName: other-service-account
   ```

2. Create a template to see the Helm release that uses your custom override values.
   ```shell
   helm template gloo-mesh https://storage.googleapis.com/gloo-mesh/gloo-mesh/gloo-mesh-$GLOO_MESH_VERSION.tgz --namespace gloo-mesh --values values.yaml
   ```

3. Install the Helm release with your custom override values.
   ```shell
   helm install gloo-mesh gloo-mesh/gloo-mesh --namespace gloo-mesh --create-namespace --values values.yaml
   ```

## Verify installation

1. After you install Gloo Mesh into the management cluster, view the management components that are installed.
   ```shell
   kubectl get pods -n gloo-mesh
   ```

   Example output:
   ```
   NAME                          READY   STATUS    RESTARTS   AGE
   discovery-66675cf6fd-cdlpq    1/1     Running   0          32m
   networking-6d7686564d-ngrdq   1/1     Running   0          32m
   ```

2. Verify that the components are install correctly.
   ```shell
   meshctl check server
   ```

   Example output:
   ```
   Gloo Mesh Management Cluster Installation

   🟢 Gloo Mesh Pods Status

   🟢 Gloo Mesh Agents Connectivity

   Management Configuration

   🔴 Gloo Mesh CRD Versions
     * deployments.apps "enterprise-networking" not found

   🟢 Gloo Mesh Networking Configuration Resources
   ```
   The `enterprise-networking` CRD is not found because the CRD is not included in a Gloo Mesh Open Source installation.

## Next steps

Next, you can [install an Istio service mesh into remote clusters]({{% versioned_link_path fromRoot="/setup/installing_istio" %}}) and [register the remote clusters with Gloo Mesh]({{% versioned_link_path fromRoot="/setup/community_cluster_registration" %}}).