---
title: Set-Style API (Enterprise)
description: Set-style rate limiting API to better limit on descriptor subsets
weight: 15
---

{{% notice note %}}
Set-style rate limiting was introduced with **Gloo Mesh Enterprise**, release `v1.1.0`.
If you are using an earlier version, this feature will not be available.
{{% /notice %}}

Gloo Mesh Enterprise exposes a fine-grained API that allows you to configure a vast number of rate limiting use cases
by defining actions that specify an ordered tuple of descriptor keys to attach to a request and descriptors that match
an ordered tuple of descriptor keys and apply an associated rate limit.

Although powerful, this API has some drawbacks.
We only limit requests whose ordered descriptors match a rule exactly.
If, for example, we want to limit requests with an `x-type` header but limit requests differently
that have an `x-type` header as well as an `x-number` header equal to `5`, we need two sets of actions on each request-
one that gets only the value of `x-type` and another that gets the value of both `x-type` and `x-number`.
While this is certainly doable, it can quickly become verbose with enough descriptor keys.
We might need to enumerate all the combinations of descriptors when we want to rate limit based on several different subsets.

To address these shortcomings, we introduced a new API.
You can define rate limits using set-style descriptors.
These are treated as an unordered set such that a given rule will apply if all the specified descriptors match,
regardless of the presence and value of the other descriptors and regardless of descriptor order.
For example, a rule may match `type: a` and `number: one` but the `color` descriptor can have any value.
This can also be understood as `color: *` where * is a wildcard.

### SetActions
`setActions` have the same structure as the `actions` already used for rate limiting but must be listed under `setActions`
to indicate to the rate limit server that they should be treated as a set and not an ordered tuple.

### SetDescriptors
`setDescriptors` specify a rate limit along with any number of `simpleDescriptors` which, like `descriptors`, must include a key
and can optionally include a value.

### Simple Example
Let's run through a simple example that uses set-style rate limiting.

#### Initial setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`). Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/installation/enterprise" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section](#bookinfo-deployment).

Now let's create a simple VirtualGateway routing to this application. (It may take a few seconds to be Accepted.)

```bash
kubectl apply -f - << EOF
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: test-gateway
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
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
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
EOF
```

To verify that the VirtualGateway works, let's send a request:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1
```

It should return the expected response:
```
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

#### Add rate limit configuration
Now, let's add a `RateLimitServerConfig` resource to include a `setDescriptor` rate limiting rule:

```bash
kubectl apply -f - << EOF
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
    setDescriptors:
      - simpleDescriptors:
          - key: type
            value: a
          - key: number
            value: one
        rateLimit:
          requestsPerUnit: 1
          unit: MINUTE
EOF
```

Now edit the VirtualGateway to include `setActions`:

```bash
kubectl edit vs test-gateway -n bookinfo
```

and add `setActions` capturing the `x-number` and `x-type` headers on the virtualHost:

{{< highlight yaml "hl_lines=18-26" >}}
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
                    - setActions:
                      - requestHeaders:
                          descriptorKey: number
                          headerName: x-number
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

Note that the descriptor order doesn't match, but this is irrelevant for set-style rate limiting.

#### Test our configuration
Let's verify that our rate limit policy is correctly enforced.

Let's try sending some requests to the Bookinfo app. Submit the following command twice:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -v -H "x-type: a" -H "x-number: one"
```

On the second attempt you should receive the following response:

```shell script
< HTTP/1.1 429 Too Many Requests
< x-envoy-ratelimited: true
< date: Tue, 14 Jul 2020 23:13:18 GMT
< server: envoy
< content-length: 0
```

This demonstrates that the rate limit is enforced.

#### Understanding set-style rate limiting functionality

Now modify the VirtualGateway `setActions` to add another descriptor. For example:
```yaml
          - requestHeaders:
              descriptorKey: color
              headerName: x-color
```

Send the following `curl` request a few times.

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -v -H "x-type: a" -H "x-number: one"  -H "x-color: blue"
```
You should see that the request is still rate limited. Since the `setDescriptor` rule only looks for two descriptors,
it still matches whether more descriptors are present or not.

However, if you modify the VirtualGateway `setActions` to remove the `type` or `number` descriptor, the request will no
longer be rate limited.

### Rule Priority

By default, `setDescriptor` rules are evaluated in the order they are listed. If a rule matches, later rules are ignored.

For example, consider the following rules:

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
    setDescriptors:
    - simpleDescriptors:
      - key: type
      - key: number
      rateLimit:
        requestsPerUnit: 10
        unit: MINUTE
    - simpleDescriptors:
      - key: type
      rateLimit:
        requestsPerUnit: 5
        unit: MINUTE
```

If the type and number are both present on a request, we want the limit to be 10 per minute.
However, if only the type is present on a request, we want the limit to be 5 per minute.

You can also specify the `alwaysApply` flag. This tells the server to consider a rule even if an earlier rule has already matched.

For example, if we have the same configuration as above but with the `alwaysApply` flag set to true,
a request with both type and number present will be limited after just 5 requests per minute, as both rules below are now considered.

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
    setDescriptors:
    - simpleDescriptors:
      - key: type
      - key: number
      rateLimit:
        requestsPerUnit: 10
        unit: MINUTE
    - simpleDescriptors:
      - key: type
      rateLimit:
        requestsPerUnit: 5
        unit: MINUTE
      alwaysApply: true
```

### All-Encompassing Rules

We can also create rules that match all requests by omitting `simpleDescriptors` altogether.
Any `setDescriptor` rule should match requests whose descriptors contain the rule's `simpleDescriptors` as a subset.
If `simpleDescriptors` is omitted from the rule, requests whose descriptors contain the empty set as a subset should match,
i.e., all requests.

These rules should be listed after all other rules without `alwaysApply` set to true, or later rules will not be considered
due to rule priority, as explained above.

An all-encompassing rule without `simpleDescriptors` would look like this:

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
    setDescriptors:
    - rateLimit:
        requestsPerUnit: 10
        unit: MINUTE
```

This rule will limit all requests to at most 10 per minute.
