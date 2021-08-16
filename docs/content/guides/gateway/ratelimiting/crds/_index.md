---
title: Ratelimit Configs (Enterprise)
description: Powerful, reusable configuration for Gloo Mesh Gateway Enterprise's rate-limit service.
weight: 20
---

{{% notice note %}}
Rate limit configuration via `RateLimitServerConfig` and `RateLimitClientConfig` resources was introduced with **Gloo Mesh Gateway Enterprise**, release `v1.1.0`.
If you are using an earlier version, this feature will not be available.
{{% /notice %}}

Gloo Mesh Gateway Enterprise exposes a fine-grained API that allows you to configure a vast number of rate limiting use cases.
The two main objects that make up the API are:
1. the [`descriptors`]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/envoy//#descriptors" %}})
   and/or [`setDescriptors`]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/set//#setdescriptors" %}}),
   which configure the rate limit server and are defined on the `RateLimitServerConfig` resource, and
2. the [`actions`]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/envoy//#actions" %}})
   and/or [`setActions`]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/set//#setactions" %}})that
   determine how Envoy composes the descriptors that are sent to the server to check whether a request should be
   rate-limited; `actions` and `setActions` are defined either on the `Route` or on the `VirtualHost` `options`.

### Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`) with Ratelimit enabled. Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

### RateLimitServerConfig resource
Starting with Gloo Mesh Enterprise `v1.1.0` you can define rate limits by creating `RateLimitServerConfig` and `RateLimitClientConfig` resources.
A `RateLimitServerConfig` resource represents a the rate limit server policy; this means that Gloo Mesh will use the resource
to configure only the Gloo Mesh Enterprise rate limit server it communicates with. A `RateLimitClientConfig` resource represents a the Envoy client policy;
this means that Gloo Mesh will use the resource to configure only the Envoy proxies.

Gloo Mesh guarantees that rate limit rules defined on different `RateLimitServerConfig` resources are completely independent of each other.

Here is a simple example of a `RateLimitServerConfig` resource:

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
        rateLimit:
          requestsPerUnit: 4
          unit: MINUTE
        value: counter
```

Here is a simple example of a `RateLimitClientConfig` resource:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitClientConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: rl-config
  namespace: bookinfo
spec:
  rateLimits:
    raw:
      - rateLimits:
        - actions:
          - genericKey:
              descriptorValue: counter
```

Once an `RateLimitServerConfig` and `RateLimitClientConfig` are created, it can be used to enforce rate limits by referencing the resource at two different levels:

- on **VirtualHosts** and
- on **Routes**,

The configuration format is the same in both cases. It must be specified under the relevant TrafficPolicy `options` attribute
(on VirtualHosts or Routes). This snippet shows an example configuration that uses the above `RateLimitClientConfig`:

```yaml
rateLimit:
  ratelimitClientConfigRef:
    name: rl-config
    namespace: bookinfo
  ratelimitConfigLabels:
    labels:
      rl-config: bookinfo
```

`RateLimitClientConfig`s defined on a `VirtualHost` is inherited by all the `Route`s that belong to that `VirtualHost`,
unless a route itself references its own `RateLimitClientConfig`s.

#### Configuration format
Each `RateLimitServerConfig` is an instance of one specific configuration type. Currently, only the `raw` configuration type
is implemented, but we are planning on adding more high-level configuration formats to support specific use cases
(e.g. limiting requests based on the presence and value of a header, or on a per-destination, per-client basis, etc.).

The `raw` configuration allows you to specify rate limit policies using the raw configuration formats used by the
server and the client (Envoy). It consists of two elements:

- a list of `descriptors`,
- a list of `setDescriptors`

These objects have the exact same format as the `descriptors` and `setDescriptors` that are explained in detail in the
[Envoy API guide]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/envoy/" %}})
and [Set-Style API guide]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/set/" %}}).

### Example
Let's run through an example that uses `RateLimitServerConfig` and `RateLimitClientConfig` resources to enforce rate limit policies on your `VirtualGateway`.
As mentioned earlier, all the examples that are listed in the [Envoy API guide]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/envoy/" %}})
and [Set-Style API guide]({{% versioned_link_path fromRoot="/guides/security/rate_limiting/set/" %}})
apply to `RateLimitServerConfig`s and `RateLimitClientConfig`'s as well, so please be sure to check them out.

#### Initial setup
First, we need to install Gloo Mesh Enterprise (minimum version `v1.1.0`). Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/installation/enterprise" >}}) for details.

Next, make sure you have the Bookinfo sample application setup. You can install the application by following the steps in the [Bookinfo deployment section](#bookinfo-deployment).

Now let's create a VirtualGateway with two different routes. Requests with the `/ratings` and `/reviews` path prefixes
will be routed to the Bookinfo service.

```yaml
kubectl apply -f - << EOF
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
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: ratings
                  namespace: bookinfo
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: reviews
                  namespace: bookinfo
  ingressGatewaySelectors:
    portName: http2
    destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
EOF
```

To verify that the VirtualGateway works, let's send a request to bookinfo:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1
```

This should return the expected response from ratings:

```
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

This should return the expected response from reviews:

```
{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!", "rating": {"stars": 5, "color": "black"}},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.", "rating": {"stars": 4, "color": "black"}}]}
```

#### Apply rate limit policies
Now let's create two `RateLimitServerConfig` resources.

```yaml
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
    descriptors:
    - key: generic_key
      rateLimit:
        requestsPerUnit: 4
        unit: MINUTE
      value: count

---
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RateLimitServerConfig
metadata:
  labels:
    app: bookinfo-policies
    app.kubernetes.io/name: bookinfo-policies
  name: per-destination-counter
  namespace: bookinfo
spec:
  raw:
    descriptors:
    - key: destination_cluster
      rateLimit:
        requestsPerUnit: 3
        unit: MINUTE
EOF
```

Let’s see what each of these resources represents:

* rl-config defines a simple counter. Every time a request matches a route that references this resource, the counter will be increased. After the counter has been increased 4 times within a 1-minute time window, successive requests in the same time window will be rejected with a 429 response code;

* per-destination-counter defines a set of counters. Each counter tracks requests to a specific cluster. After a destination has received 3 requests within a 1-minute time window, successive requests to the same destination in the same time window will be rejected with a 429 response code.

Now let's apply these policies to our `VirtualGateway`:

```shell script
kubectl apply -f - << EOF
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
                    - genericKey:
                        descriptorValue: count
                  - actions:
                    - destinationCluster: {}
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
          - matchers:
            - uri:
                prefix: /reviews
            name: reviews
            routeAction:
              destinations:
              - kubeService:
                  clusterName: mgmt-cluster
                  name: reviews
                  namespace: bookinfo
      routeOptions:
        rateLimit:
          denyOnFail: true
  ingressGatewaySelectors:
    portName: http2
    destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - mgmt-cluster
        labels:
          istio: ingressgateway-ns
        namespaces:
        - istio-system
EOF
```

We have applied these two policies to the `VirtualHost`, so they apply to both of the routes that belong to
the `VirtualHost`. This will cause requests to be rate-limited either when:

- one of the two destination is hit more than **3 times within a minute**, or
- the aggregate of both destination is hit more than **4 times within a minute**.

You can verify that Gloo Mesh has been correctly configured by port-forwarding the rate limit server and requesting a
config dump. First run:

```shell script
kubectl port-forward -n gloo-mesh deploy/rate-limiter 9091
```

Then - from a separate shell - run:

```shell script
curl http://localhost:9091/rlconfig/
```

You should get the following response:

```
domain: solo.io
  treeDescriptors:
    - solo.io|generic_key^bookinfo.per-destination-counter|destination_cluster: unit=MINUTE requests_per_unit=3 weight=0 always_apply=false
    - solo.io|generic_key^bookinfo.rl-config|generic_key^count: unit=HOUR requests_per_unit=1 weight=0 always_apply=false
  setDescriptors:
```

#### Test our configuration
Let's verify that our rate limit policies are correctly enforced.

First, let's try sending some requests to the `ratings` destination. Submit the following command multiple times in rapid succession:

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1
```

On the **fourth attempt** you should receive the following response:

```shell script
< HTTP/1.1 429 Too Many Requests
< x-envoy-ratelimited: true
< date: Tue, 14 Jul 2020 23:13:18 GMT
< server: envoy
< content-length: 0
```

This demonstrates that the global rate limit is enforced. Now let's wait for a minute for the counter to reset
and then submit the same command again, but this time only **3 times**:

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1
```

You should get three successful responses.
After the third attempt, let's start sending requests to the `reviews` destination:

```shell script
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/reviews/1
```

The **fifth attempt** should return the `429 Too Many Reqeusts` response:

```shell script
< HTTP/1.1 429 Too Many Requests
< x-envoy-ratelimited: true
< date: Tue, 14 Jul 2020 23:13:18 GMT
< server: envoy
< content-length: 0
```

This is because although we get 3 requests per minute on the destination, we have a global-limit of 4 requests per minute across both destinations.
