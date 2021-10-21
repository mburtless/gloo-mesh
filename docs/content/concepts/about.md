---
title: "About"
menuTitle: About
description: Learn about the features and benefits of Gloo Mesh
weight: 20
---

{{< readfile file="static/content/gm_about.txt" markdown="true">}}

## What is Gloo Mesh?

{{%excerpt%}}
Gloo Mesh is an open source distribution of the [Istio service mesh](https://istio.io/). The Gloo Mesh API simplifies the complexity of your service mesh by installing custom resource definitions (CRDs) that you configure. Then, Gloo Mesh translates these CRDs into Istio resources across your environment, and provides visibility across all of the resources and traffic.
{{%/excerpt%}}

{{% notice tip %}}
{{< readfile file="static/content/try_gme" markdown="true">}}
{{% /notice %}}

As shown in the following figure, Gloo Mesh manages the configuration of service meshes across clusters, so that you can control service-to-service traffic consistently. 

<figure><img src="{{< versioned_link_path fromRoot="/img/gloomesh-3clusters.png">}}">
<figcaption style="text-align:center;font-style:italic">Figure of Gloo Mesh managing service meshes across clusters.</figcaption></figure>

## Why use Gloo Mesh?

With Gloo Mesh, you get an extensible, open-source set of API tools to connect and manage your services across multiple clusters and service meshes.

### Use Gloo Mesh Open Source for demonstration purposes

{{< readfile file="static/content/gmoss_about.txt" markdown="true">}}

In the following table, review the benefits of using Gloo Mesh Open Source.

{{% notice tip %}}
Looking for a full list of features compared against what's available in open source? See the [Feature comparison](https://www.solo.io/products/gloo-mesh/) on the product website.
{{% /notice %}}

| Benefit | Community Istio | Gloo Mesh Open Source | Gloo Mesh Enterprise |
| ------- | :-------------: | :-------------------: | :------------------: |
| Upstream-first approach to feature development | ✅ | ✅ | ✅ | 
| Installation, upgrade, and management across clusters and service meshes | ❌ | ✅ | ✅ |
| Security features like self-cert signing, federated trust, and multi-tenancy | ❌ | ✅ | ✅ |
| Global service discovery across service meshes | ❌ | ✅ | ✅ |
| Reliable multicluster routing with virtual destinations and locality failover | ❌ | ❌ | ✅ |
| Dynamic scaling and global observability | ❌ | ❌ | ✅ |
| End-to-end Istio support and CVE security patching for `n-4` versions | ❌ | ❌ | ✅  |
| Specialty builds for distroless and FIPS compliance | ❌ | ❌ | ✅  |
| 24x7 production support and one-hour Severity 1 SLA | ❌ | ❌ | ✅  |
| Gateway, Web Assembly, and Portal modules to extend functionality | ❌ | ❌ | ✅  |

### Use Gloo Mesh as a multicluster management plane

Gloo Mesh consists of a set of components that run on a single cluster, often referred to as a *management plane cluster*. The management plane components are stateless and rely exclusively on declarative CRDs. Each service mesh installation that spans a deployment footprint often has its own control plane. You can think of Gloo Mesh as a management plane for multiple control planes.

You can run Gloo Mesh in a dedicated management plane cluster, or co-locate with a service mesh in an existing cluster. From the management cluster, Gloo Mesh remotely operates and drives the configuration for specific service mesh control planes. This allows Gloo Mesh to discover service meshes and workloads, establish federated identity, and enable global traffic routing and load balancing across multiple remote clusters. Additionally, Gloo Mesh provides centralized access control policies, observability and monitoring, and more, as shown in the following figure.

<figure><img src="{{< versioned_link_path fromRoot="/img/concepts-gloomesh-components.png">}}">
<figcaption style="text-align:center;font-style:italic">Figure of Gloo Mesh multicluster discovery and management components.</figcaption></figure>

### Use Gloo Mesh as a complete service mesh management solution

With an increasingly distributed environment for your apps, you need a flexible, open-source based solution to help meet your traditional and new IT requirements. Solo is laser-focused on developing the best service mesh and API gateway solutions, unlike other offerings that might be a small part of a vendor-locked in solution. Furthermore, Gloo Mesh is built with the following six principles.

1. **Secure**: You need a zero-trust model and end-to-end controls to implement best practices and internal regulations. You can start with Gloo Mesh to get familiar with using the API to set up traffic policies and other custom resources. Later in your journey, you can use Gloo Mesh Enterprise to comply with strict regulations like FIPS and to reduce the risk of running older versions with security patching. 
2. **Reliable**: You need a robust management tool with features like priority failover and locality-aware load balancing to manage your service mesh and API gateway for your mission-critical workloads.
3. **Unified**: You need one centralized tool to manage and observe your application environments and traffic policies at scale.
4. **Simplified**: Your developers need a simple, declarative, API-based method to provide services to your apps without further coding and without needing to understand the complex technologies like Istio and Kubernetes that underlie your environments. 
5. **Comprehensive**: You need east-west service traffic management across infrastructure resources on-premises and across clouds. If you choose to upgrade to Gloo Mesh Enterprise, you can add the Gloo Mesh Gateway module with north-south ingress control for a complete traffic solution. 
6. **Modern and open**: You need a solution that is designed from the ground up on open-source, cloud-native best practices and leading projects like Kubernetes and Istio to maximize the portability and scalability of your development processes.

See how Gloo Mesh is an **Outperformer** and **Leader** in the service mesh space in the following GigaOM Radar report.

{{< button href="https://lp.solo.io/report-download" >}} Download the report {{< /button >}}
