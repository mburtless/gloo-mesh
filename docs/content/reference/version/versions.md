---
title: "Supported versions"
description: | 
  Information about supported Gloo Mesh Open Source versions and related tools such as Istio.
weight: 2
---

Review the following information about supported release versions for Gloo Mesh Open Source, including dependencies on other open source projects like Istio.

{{% notice tip %}}
Want to try Gloo Mesh in production? Try [Gloo Mesh Enterprise](https://docs.solo.io/gloo-mesh-enterprise/), which includes a quarterly release cadence, additional features and mesh capabilities such as Gloo Mesh Gateway, enterprise support, and hardened versions of Istio for CVE and security patching.
{{% /notice %}}

## Supported versions

The following versions of Gloo Mesh Open Source are supported with the compatible open source projects versions of Istio and Kubernetes. Later versions of the open source projects that are released after Gloo Mesh Open Source might also work, but are not tested as part of the Gloo Mesh Open Source release. 

| Gloo Mesh Open Source | Istio`*` | Kubernetes`*` |
| -------------------- | --------- | ------------- |
| 1.1 | 1.7 - 1.10 | 1.16 - 1.21 |
| 1.0 | 1.7 - 1.10 | 1.16 - 1.21 |

{{% notice note %}}
`*` **Istio and Kubernetes**: Supported Kubernetes versions are based on the version of Istio that is installed. For example, you cannot use Gloo Mesh Open Source with Istio 1.9 on a Kubernetes 1.22 cluster, because Istio 1.9 does not support Kubernetes 1.22. To review Istio support of Kubernetes versions, see the [Istio documentation](https://istio.io/latest/docs/releases/supported-releases/#support-status-of-istio-releases).

**OpenShift and Kubernetes**: The Istio and Kubernetes versions also determines which version of OpenShift you can run. For example, if you have Istio 1.11 you can run OpenShift 4.8, which uses Kubernetes 1.21. To review OpenShift Kubernetes support, see the [OpenShift changelog documentation for the version you want to use](https://docs.openshift.com/container-platform/4.8/release_notes/ocp-4-8-release-notes.html).
{{% /notice %}}

### Version skew policy for management and remote clusters

Ideally, run the same versions of Gloo Mesh Open Source and Kubernetes in your management and remote clusters. To give you time to upgrade all of the remote clusters, your remote clusters can run one version behind the management clusters (`n-1`). Do not plan to run different versions of the Gloo Mesh deployments on your management and remote clusters for longer than you need to complete the upgrade.

Remote clusters can run different versions of Istio. However, if you want to apply policies or other resources that require a certain version of Istio across remote clusters, make sure that the clusters run a supported version.

### Upgrading versions

The upgrade process depends on which software you need to upgrade and your infrastructure provider.
* **Gloo Mesh Open Source**: See the [Upgrading guide]({{< versioned_link_path fromRoot="/operations/upgrading/" >}}).
* **Istio**: See the [Istio documentation](https://istio.io/latest/docs/setup/upgrade/).
* **Kubernetes or OpenShift**: Consult your infrastructure provider's upgrade process. For example, you might use [Amazon Elastic Kubernetes Service (EKS)](https://docs.aws.amazon.com/eks/latest/userguide/update-cluster.html), [Google Kubernetes Engine (GKE)](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-upgrades), [IBM Cloud Kubernetes Service](https://cloud.ibm.com/docs/containers?topic=containers-update), or [Microsoft Azure Kubernetes Service (AKS)](https://docs.microsoft.com/en-us/azure/aks/upgrade-cluster).

## Open source packages in Gloo Mesh Open Source

For specific versions of open sources packages that are bundled with Gloo Mesh Open Source, see the entries in the [Open Source Attribution topic]({{< versioned_link_path fromRoot="/reference/osa/" >}}). For more information on where these open source packages are retrieved from, see the [go.mod documentation](https://golang.org/ref/mod#vcs-find).
