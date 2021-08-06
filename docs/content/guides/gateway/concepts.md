---
title: Gateway Concepts Overview
menuTitle: Concepts
description: A walkthrough of the Gloo Mesh Gateway configuration resources.
weight: 10
---

{{% notice note %}} Basic Gateway features are available to anyone with a Gloo Mesh Enterprise license. Some advanced features require a Gloo Mesh Gateway license. {{% /notice %}}

## Gloo Mesh Gateway

Gloo Mesh Gateway is an abstraction built on top of [Istio's ingress gateway model](https://istio.io/latest/docs/tasks/traffic-management/ingress/ingress-control/). It can be used as a way to simplify the configuration of ingress traffic rules while retaining seamless integration with your service mesh. It also adds powerful mutli-cluster, multi-mesh capabilities to your gateway.

## Example Architectures

There are many ways to set up ingress gateways into your service mesh, particularly when you scale up to multiple clusters and multiple meshes. Below we have a few examples of what this could look like.

### Example 1: Single cluster, Single Gateway

![Single Cluster Single Gateway]({{% versioned_link_path fromRoot="/img/gateway/gateway-single-cluster.png" %}})

In this setup, we have a single ingress gateway routing traffic to a single cluster.

### Example 2: Single cluster, Multi Gateway

![Single cluster, Multi Gateway]({{% versioned_link_path fromRoot="/img/gateway/gateway-single-cluster-multi-gateway.png" %}})

In this setup, we have a single cluster deployed with multiple gateways. The services which these gateways route traffic to can overlap, if desired.

### Example 3: Mutlicluster, Single Gateway

![Mutli cluster, Single Gateway]({{% versioned_link_path fromRoot="/img/gateway/gateway-multi-cluster-single-gateway.png" %}})

In this setup, our application has grown and now spans two meshes deployed across two clusters. We still have a single ingress gateway configured, but notably it is aware of services running in both meshes and both clusters. Having the application split like this can help with separation of concerns when multiple application teams are involved. It also offers better resilience and high availability, should an application in one cluster fail.

### Example 4: Mutlicluster, Multi Gateway

![Mutli cluster, Multi Gateway]({{% versioned_link_path fromRoot="/img/gateway/gateway-multi-cluster-multi-gateway.png" %}})

In this setup, we have added another gateway, this time to the second cluster. Having multiple ingresses like this can help with high availability, as well as location-aware routing. A network load balancer could be configured to route between both ingress gateways, and once the traffic reaches a specific gateway, it will have the ability to prioritize routing to services local to its own cluster, before falling back on routing to other clusters if those services are unavailable locally for any reason. It is worth noting that even though we have multiple gateways deployed here, they are both configured in one place - in the Gloo Mesh management plane. Similarly, both can be observed using Gloo Mesh's observability suite.

## Gloo Mesh Gateway Custom Resources

In order to make all of the above scenarios easier to configure, we have introduced three new Custom Resources (CRs):

- VirtualGateway
- VirtualHost
- RouteTable

Each of these resources will live in the Management Cluster, although they may be used to configure ingress resources in any cluster managed by Gloo Mesh. These resources are installed as part of Gloo Mesh Enterprise, and no additional settings are required to start using Gateway features.

### VirtualGateway

A VirtualGateway resource will define the [Envoy listeners](https://www.envoyproxy.io/docs/envoy/latest/configuration/listeners/listeners), i.e. the protocols and ports to listen on. It will also define which Envoy filters are attached, by allowing users to specify a [FilterChain](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/listeners/network_filter_chain) (or multiple FilterChains). It attaches to existing ingress gateway workloads via a workload selector, and specifies one or more VirtualHost configurations to configure routing rules. A single VirtualGateway can apply to multiple ingress gateway deployments across all meshes and clusters contained within a VirtualMesh. Attached Gateways will be listed in the status of the VirtualGateway. In order to perform cross-mesh routing, the Gateway Mesh and Destination Mesh _must be contained in a single VirtualMesh with federation enabled_.

### VirtualHost

VirtualHosts are selected by a VirtualGateway. A single VirtualHost can be selected by multiple VirtualGateways. VirtualHosts are responsible for configuring top level settings, such as domains. In addition, a VirtualHost can set route options that apply to all child routes, which will be inherited as a default unless explicitly overridden at the route level. A VirtualHost contains a list of Routes, which can contain various matchers and actions. 

### RouteTable

A RouteTable is effectively a list of Routes. It exists only to be delegated to from VirtualHosts, or other RouteTables. This allows users to separate configuration and ownership across the organization. For example an app-level team may want to configure all of the low-level endpoints that are sent to their service, but may not have control over which domains they can serve traffic on, or what the authorization policies are. Routes configured in RouteTables will by default inherit any options configured by their parent delegating resource (RouteTable or VirtualHost), for example, timeout or retry settings.

### Routes

Routes can be configured on a VirtualHost, and/or a RouteTable. Routes consist of options, matchers, and an action. Options are a way to specify route policies in-line, for example a retry policy or a timeout. Matchers are how the routing logic determines which route is selected. It could match on a path, http method, or header, for example. Routes can have one of four actions:

#### Route Action

Used to send traffic directly to one of three destination types:
- A Kubernetes service, the most commonly used option. Notably, this is a service in a specific _cluster_.
- A VirtualDestination, useful for locality based failover.
- A static destination, useful for routing to services outside of your service mesh.

#### Redirect Action

Redirect actions are used to redirect certain requests to alternate paths. Commonly used to configure eg HTTP 301 responses.

#### Direct Response Action

Direct response action can be used to directly return a specified body and HTTP status from the proxy, without forwarding the request upstream at all. If necessary, headers can be specified usign the Header Modification feature in the enclosing route.

#### Delegate Action

Delegate action will pass the routing responsibility to a RouteTable. This can be very useful when the number of routes configured on a given gateway starts to become large, and multiple teams need to make route edits.

## Summary

On this page we talked about the Gloo Mesh resources used to configure Gloo Mesh Gateway, as well as several network topology options commonly used when deploying it. You can check out our [Getting Started Guide]({{% versioned_link_path fromRoot="/guides/gateway/getting_started" %}}) to see some specific examples and try out these concepts for yourself!
