---
title: Envoy API (Enterprise)
description: Fine-grained rate limit API.
weight: 10
---

## Table of Contents

- [Overview](#overview)
    - [Descriptors](#descriptors)
    - [Actions](#actions)
- [Simple Examples](#simple-examples)
    - [Generic Key](#generic-key)
    - [Header Values](#header-values)
    - [Remote Address](#remote-address)
- [Advanced Concepts](#advanced-concepts)
    - [Defining limits for tuples of key-value pairs](#defining-limits-for-tuples-of-key-value-pairs)
    - [Nested Limits](#nested-limits)
    - [Rule Priority and Weights](#rule-priority-and-weights)
    - [Customizing Routes](#customizing-routes)
- [Advanced Use Cases](#advanced-use-cases)
    - [Configuring multiple limits per remote address](#configuring-multiple-limits-per-remote-address)
    - [Traffic prioritization based on HTTP method](#traffic-prioritization-based-on-http-method)
    - [Securing rate limit actions with JWTs](#securing-rate-limit-actions-with-jwts)
    - [Improving security further with WAF and authorization](#improving-security-further-with-waf-and-authorization)

## Overview

In this document, we show how to use Gloo Mesh with
[Envoy's rate-limit API](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto).

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

{{% notice note %}}
Gloo Mesh Gateway Enterprise includes a rate limit server based on the implementation [here](https://github.com/envoyproxy/ratelimit).
It is already installed when Gloo Mesh Gateway with `meshctl` or helm. To get your trial license key, go to <https://www.solo.io/gloo-trial>.

<br />

Open source Gloo Mesh does not support rate limiting.

{{% /notice %}}

### Descriptors

Rate limiting descriptors define an ordered tuple of keys that must match for the associated rate limit to be applied.
The tuple of keys are expressed as a hierarchy to make configuration easy, but it's the complete tuple of keys
matching or not that is important. Each descriptor key can have an associated value that is matched as a literal.
You can define rate limits on a key matching a specific value, or you can omit the value to have the limit applied to
any unique value for that key.
See the Envoy rate limiting [configuration doc](https://github.com/envoyproxy/ratelimit#configuration) for full details.

Rate limit descriptors live in `RateLimitServerConfig` crd, so the examples below will reflect applying the `RateLimitServerConfig` configuration.

### Actions

The [Envoy rate limiting actions](https://www.envoyproxy.io/docs/envoy/v1.14.1/api-v2/api/v2/route/route_components.proto#envoy-api-msg-route-ratelimit-action)
associated with the VirtualGateway or the individual routes allow you to specify how parts of the request are
associated to rate limiting descriptor keys defined in the `RateLimitServerConfig`. Essentially, these actions tell Gloo Mesh which rate limit counters
to increment for a particular request.

You can specify more than one rate limit action, and the request is throttled if any one of the actions triggers
the rate limiting service to signal throttling, i.e., the rate limiting actions are effectively OR'd together.

## Simple Examples

Let's go through a series of simple rate limiting examples to understand the basic options for defining rate limiting
descriptors and actions. Then, we'll go through more complex examples that use nested tuples of keys, to express more
realistic use cases.

### Generic Key

A generic key is a specific string literal that will be used to match an action to a descriptor. For instance, we could
use these descriptors in the `RateLimitServerConfig`:

```yaml
cat << EOF | kubectl apply -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: generic_key
        value: per-second
        rateLimit:
          requestsPerUnit: 2
          unit: SECOND
EOF
```

This defines a limit of 2 requests per second for any request that triggers an action on the generic key called `per-second`.
We could define that action on a VirtualGateway like so:

{{< highlight yaml "hl_lines=22-30" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            options:
              - rateLimit:
                  raw:
                    - rateLimits:
                      - actions:
                        - genericKey:
                            descriptorValue: per-second
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

{{% notice note %}}
In Envoy, the rate limit config is typically written with snake case keys ("example_config") in the YAML, whereas in Gloo Mesh and Kubernetes
YAML keys typically use camel case ("exampleConfig"). We'll use camel case notation when writing YAML keys in Gloo Mesh config here.
{{% /notice %}}

### Header Values

It may be desirable to create actions based on the value of a header, which is dynamic based on the request, rather than
a generic key, that is static based on the route. The following configuration will define a descriptor that limits requests to 2 per minute
for any unique value for `type`:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
    - key: type
      rateLimit:
        requestsPerUnit: 2
        unit: MINUTE
```

Now we can create a route that triggers a rate limit action for this descriptor:

{{< highlight yaml "hl_lines=18-24" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          options:
            trafficPolicy:
              rateLimit:
                raw:
                  rateLimits:
                  - actions:
                    - requestHeaders:
                        descriptorKey: type
                        headerName: x-type
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

With this config, a rate limit of 2 per minute will be enforced for requests depending on the value of the `x-type` header,
for every unique value. If we only wanted to enforce this limit for a specific value, we could write that value into our descriptor:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: type
        value: example
        rateLimit:
          requestsPerUnit: 2
          unit: MINUTE
```

Now, requests that are routing using the VirtualGateway above will be rate limited after 2 requests per second only if the
request includes a header `x-type: example`.

### Remote Address

A common use case is to rate limit based on client IP address, also referred to as the downstream remote address. To utilize this, we
can define a descriptor called `remote_address` in the `RateLimitServerConfig`:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: remote_address
        rateLimit:
          requestsPerUnit: 2
          unit: MINUTE
```

On the route, we can define an action to count against this descriptor in the following way:

{{< highlight yaml "hl_lines=18-22" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          options:
            trafficPolicy:
              rateLimit:
                raw:
                  rateLimits:
                  - actions:
                    - remoteAddress: {}
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

{{% notice warning %}}
You may need to make additional configuration changes to Gloo Mesh in order for the `remote_address` value to be the real
client IP address, and not an address that is internal to the Kubernetes cluster, or that is from a cloud load balancer.
To address these, check out the advanced use case below.
{{% /notice %}}

## Advanced Concepts

Now that you understand the basic ways to define descriptors and link those to rate limit actions on your routes, we can
dig into some more advanced concepts.

### Defining limits for tuples of key-value pairs

In the `RateLimitServerConfig`, you can define nested descriptors to start to express rules based on tuples instead of a single value.

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: remote_address
        descriptors:
          - key: type
            descriptors:
              - key: number
                rateLimit:
                  requestsPerUnit: 1
                  unit: MINUTE
```

This rule enforces a limit of 1 request per minute for any unique combination of `type` and `number` values. We can define
multiple actions on our routes to apply this rule:

{{< highlight yaml "hl_lines=18-27" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          options:
            trafficPolicy:
              rateLimit:
                raw:
                  rateLimits:
                  - actions:
                    - requestHeaders:
                        descriptorKey: type
                        headerName: x-type
                    - requestHeaders:
                        descriptorKey: number
                        headerName: x-number
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

If a request is routed using this virtual service, and the `x-type` and `x-number` headers are both present on the request,
then it will be counted towards the limit we defined above.
If one or both headers are not present on the request, then no rate limit will be enforced.

{{% notice warning %}}
The order of actions must match the order of nesting in the descriptors. So in this example, if the actions were reversed,
with the number action before the type action, then the request would not count towards the rate limit quota defined above.
{{% /notice %}}

### Nested Limits

We can define limits at each level of a descriptor tuple. For instance, we may want to enforce the same limit if the type
is provided but the number is not:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: type
        rateLimit:
          requestsPerUnit: 1
          unit: MINUTE
        descriptors:
          - key: number
            rateLimit:
              requestsPerUnit: 1
              unit: MINUTE
```

This time, on our VirtualGateway, we'll define actions for two separate rate limits - one that increments the counter
for the `type` limit specifically, and another to increment the counter for the `type` and `number` pair, when present.

{{< highlight yaml "hl_lines=18-31" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          options:
            trafficPolicy:
              rateLimit:
                raw:
                  rateLimits:
                  - actions:
                    - requestHeaders:
                        descriptorKey: type
                        headerName: x-type
                  - actions:
                    - requestHeaders:
                        descriptorKey: type
                        headerName: x-type
                    - requestHeaders:
                        descriptorKey: number
                        headerName: x-number
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

Note that we now have two different rate limits defined for this VirtualGateway. One contributes to the counter for just `type`,
if the `x-type` header is present. The other contributes to the counter for the `type` and `number` pair, if both headers are present.
The request will result in a 429 rate limit response if either limit is reached.

### Rule Priority and Weights

We may run into cases where we need fine-grained control over the order of how rules are evaluated. For example, consider
this set of descriptors:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: type
        rateLimit:
          requestsPerUnit: 1
          unit: MINUTE
        descriptors:
          - key: number
            rateLimit:
              requestsPerUnit: 10
              unit: MINUTE
```

If the type and number are both present on a request, we want the limit to be 10 per minute. However, with the VirtualGateway
from above, we would observe a limit of 1 per minute - the request would be rate limited based on the first rule, for just matching
on the `type` descriptor.

Starting in Gloo Mesh 1.1, you can now specify weights on rules. For a particular request that has multiple sets of actions,
it will evaluate each and then increment only the matching rules with the highest weight. By default, the weight is 0, so we could
fix our config above by adding a weight to the nested descriptor:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: type
        rateLimit:
          requestsPerUnit: 1
          unit: MINUTE
        descriptors:
          - key: number
            weight: 1
            rateLimit:
              requestsPerUnit: 10
              unit: MINUTE
```

Based on the virtual service defined above, when a request has both the `x-type` and `x-number` headers, then it will evaluate both
limits - the limit on type alone, and the limit on the combination of type and number. Since the latter has a higher weight, then
only that counter will be incremented. In that way, requests with a unique `type` and `number` will be allowed 10 requests per minute,
but requests that only have a type will be limited to 1 per minute.

This logic can be bypassed by using the `alwaysApply` flag. So this configuration would behave equivalently to the example before we added
the weight:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: type
        alwaysApply: true
        rateLimit:
          requestsPerUnit: 1
          unit: MINUTE
        descriptors:
          - key: number
            weight: 1
            rateLimit:
              requestsPerUnit: 10
              unit: MINUTE
```

### Customizing Routes

So far, we have been configuring rate limit actions on our route as an option under the `route`.
Alternatively, we can define this as an option on the VirtualHost:

{{< highlight yaml "hl_lines=22-35" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            options:
              - rateLimit:
                  raw:
                    - rateLimits:
                      - actions:
                        - requestHeaders:
                            descriptorKey: type
                            headerName: x-type
                        - actions:
                          - requestHeaders:
                              descriptorKey: type
                              headerName: x-type
                          - requestHeaders:
                              descriptorKey: number
                              headerName: x-number
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

## Advanced Use Cases

Here are a few more advanced cases that may be more representative of a realistic configuration.

#### Configuring multiple limits per remote address

Now, using the config from the simple example, we can use the `remote_address` descriptor to rate limit based on the real
downstream client address. However, in practice, we may want to express multiple rules, such as a per-second and per-minute
limit.

We can model this by making `remote_address` a nested descriptor, and using a generic key that is distinct. For instance,
we could model our `RateLimitServerConfig` like this:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      - key: generic_key
        value: "per-minute"
        descriptors:
          - key: remote_address
            rateLimit:
              requestsPerUnit: 20
              unit: MINUTE
      - key: generic_key
        value: "per-second"
        descriptors:
          - key: remote_address
            rateLimit:
              requestsPerUnit: 2
              unit: SECOND
```

And we can configure a route to count towards both limits:

{{< highlight yaml "hl_lines=20-32" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            options:
              - rateLimit:
                  raw:
                    - rateLimits:
                      - actions:
                        - genericKey:
                          descriptorValue: "per-minute"
                        - remoteAddress: {}
                      - actions:
                        - genericKey:
                            descriptorValue: "per-second"
                        - remoteAddress: {}
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

Now, we'll increment a per-minute and per-second rate limit counter based on the client remote address.

### Traffic prioritization based on HTTP method

A useful tactic for building resilient, distributed systems is to implement different rate limits for different "priorities" or "classes" of traffic. This practice is strongly related to the concept of [_load shedding_](https://landing.google.com/sre/workbook/chapters/managing-load/).

Suppose you have exposed an API that supports both `GET` and `POST` methods for listing data and creating  resources respectively. While both pieces of functionality are important, ultimately the `POST` action is more important to your business, so you want to protect the availability of the `POST` function at the expense of the less important `GET` function.

To implement this, we will build on the previous example and provide a global limit per remote client for all traffic classes as well a smaller limit for the less important `GET` method. This allows our system to drop the lower priority traffic and protect the higher priority traffic.

We can implement this in Gloo Mesh using a descriptor for the method of incoming request in conjunction with the remote client:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  raw:
    descriptors:
      # allow 5 calls per minute for any unique host
      - key: remote_address
        rateLimit:
          requestsPerUnit: 5
          unit: MINUTE
      # specifically limit GET requests from unique hosts to 2 per min
      - key: method
        value: GET
        descriptors:
          - key: remote_address
            rateLimit:
              requestsPerUnit: 2
              unit: MINUTE
```

With these limits in place, we are ensuring that the server doesn't get overwhelmed with `GETs`. Now we can add an action to extract the method from the `:method` psuedo-header:

{{< highlight yaml "hl_lines=22-31" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: my-gateway
  namespace: bookinfo
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHost:
          domains:
          - www.example.com
          routes:
          - matchers:
            - uri:
                prefix: /ratings
            name: ratings
            options:
              - rateLimit:
                  raw:
                    - rateLimits:
                      - actions:
                        - remoteAddress: {}
                      - actions:
                        - requestHeaders:
                            descriptorKey: method
                            headerName: :method
                        - remoteAddress: {}
                ratelimitServerConfigSelector:
                  namespaces:
                  - bookinfo
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  deployToIngressGateways:
    bindPort: 8080
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
{{< /highlight >}}

How the route will have a per-client limit for general protection while a smaller limit is in place for `GET` requests to prevent lower priority traffic from overwhelming the system.
