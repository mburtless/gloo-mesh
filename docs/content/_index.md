---
weight: 99
title: Gloo Mesh
---

## What is Gloo Mesh?

Gloo Mesh is a Kubernetes-native **management plane** that enables configuration 
and operational management of multiple heterogeneous service meshes across multiple 
clusters through a unified API. The Gloo Mesh API integrates with the leading 
service meshes and abstracts away differences between their disparate APIs, allowing 
users to configure a set of different service meshes through a single API. Gloo 
Mesh is engineered with a focus on its utility as an operational management 
tool, providing both graphical and command line UIs, observability features, and 
debugging tools.

![Gloo Mesh overview]({{% versioned_link_path fromRoot="/img/gloomesh-diagram.png" %}})

For more information about how Gloo Mesh works, check out the [core concepts that underpin Gloo Mesh]({{% versioned_link_path fromRoot="/concepts/concepts/" %}}).

### Getting to know Gloo Mesh

Gloo Mesh can be run in a dedicated cluster, or can be co-located with a service mesh in an existing cluster. From the management cluster, Gloo Mesh remotely operates and drives the configuration for specific service mesh control planes. This allows Gloo Mesh to discover service meshes and workloads, establish federated identity, and enable global traffic routing and load balancing across multiple remote clusters. Additionally, Gloo Mesh provides centralized access control policies, observability and monitoring, and more.

![Gloo Mesh architecture]({{% versioned_link_path fromRoot="/img/gloomesh-3clusters.png" %}})

For more information, check out the [Gloo Mesh architecture and components]({{% versioned_link_path fromRoot="/concepts/architecture/" %}}).

### Videos: Take a dive into Gloo Mesh

We've put together [a handful of videos](https://www.youtube.com/watch?v=4sWikVELr5M&list=PLBOtlFtGznBjr4E9xYHH9eVyiOwnk1ciK) detailing the features of Gloo Mesh.

## Contribution

There are many ways to get involved in our open source community and contribute to the Gloo Mesh Open Source project. Gloo Mesh Open Source would not be possible without the valuable work of projects in the community! To get started with contributing to code, documentation, and mode, see the [contributing guidelines]({{% versioned_link_path fromRoot="/contributing/" %}}).

## Questions and resources

- If you have questions, join the [#Gloo-Mesh channel](https://solo-io.slack.com/archives/CJQGK5TQ8) in the Solo.io Slack workspace
- Discover Gloo Mesh capabilities on the [product page](https://www.solo.io/products/gloo-mesh/)
- Learn more about [Open Source at Solo.io](https://www.solo.io/open-source/)
- Follow us on Twitter at [@soloio_inc](https://twitter.com/soloio_inc)