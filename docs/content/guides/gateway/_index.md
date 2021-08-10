---
title: "Gateway"
menuTitle: Gateway
description: Guides on Gloo Mesh Gateway features
weight: 30
---

{{% notice note %}} Basic Gateway features are available to anyone with a Gloo Mesh Enterprise license. Some advanced features require a Gloo Mesh Gateway license. {{% /notice %}}

{{% notice note %}} Gloo Mesh Gateway requires Gloo Mesh Enterprise version 1.1.0 or higher. {{% /notice %}}

Gloo Mesh Gateway allows users to harness the power of Envoy proxy as an Ingress Gateway for their service mesh, while keeping the simplicity of the Gloo Mesh API. This allows for simplifying your setup by writing advanced configuration once, and re-using it in multiple places and different contexts. For example, a rate limit policy could be written once, and applied to both traffic coming into the cluster via an ingress gateway as well as traffic inside the cluster travelling east to west.

It is recommended to start off by reading the [Gateway Concepts]({{% versioned_link_path fromRoot="/guides/gateway/concepts" %}}) page to familiarize yourself with the Gateway custom resources, as well as potential deployment patterns.

Once you're comfortable with the concepts, you can start trying them out yourself by walking through the [Getting Started Guide]({{% versioned_link_path fromRoot="/guides/gateway/getting_started" %}}).


Gateway Documentation Resources: 
{{% children description="true" %}}
