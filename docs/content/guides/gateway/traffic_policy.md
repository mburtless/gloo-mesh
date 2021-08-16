---
title: Traffic Policies
menuTitle: TrafficPolicies
description: A walkthrough of using traffic policies with Gloo Mesh Gateway
weight: 20
---

{{% notice note %}} Basic Gateway features are available to anyone with a Gloo Mesh Enterprise license. Some advanced features require a Gloo Mesh Gateway license. {{% /notice %}}

## Traffic Policies

Gloo Mesh makes use of [TrafficPolicies]({{% versioned_link_path fromRoot="/guides/traffic_policy/" %}}) for "East/West" communication between services within the mesh. These same `TrafficPolicies` can be applied to Gateway traffic. See the full `TrafficPolicy` API [here]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy/" %}}).

A standard East/West `TrafficPolicy` which adds a 5 second timeout to a particular destination might look something like this:

{{< highlight yaml "hl_lines=14" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: timeout-policy
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: reviews
          namespace: bookinfo
  policy:
    requestTimeout: 5s
{{< /highlight >}}

By default, if no `sourceSelector` is specified on a `TrafficPolicy`, it will apply to all East/West traffic going to the given `destinationSelector` from within the mesh. If we want this policy to _also_ apply to Gateway traffic, we use the `routeSelector` field:

{{< highlight yaml "hl_lines=13-16" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: timeout-policy
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: reviews
          namespace: bookinfo
  routeSelector:
    - virtualHostSelector:
        namespaces: 
        - gloo-mesh
  policy:
    requestTimeout: 5s
{{< /highlight >}}

Once we apply the above policy to the management plane, we can confirm that it has been applied by checking the status of the VirtualHost we have selected (this is assuming you have a `VirtualHost` named `demo-virtualhost` in the `gloo-mesh` namespace, from the [Getting Started Guide]({{% versioned_link_path fromRoot="/guides/gateway/getting_started" %}})). The status of a `VirtualHost`, or `RouteTable` will always show all `TrafficPolicies` that have been applied, as well as listing which routes the policy has been applied to, in the case that it's a subset.

```shell
kubectl get -n gloo-mesh virtualhost demo-virtualhost -o yaml
```

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  ...truncated for brevity...
status:
  appliedTrafficPolicies:
  - observedGeneration: 1
    ref:
      name: timeout-policy
      namespace: gloo-mesh
    routes:
    - '*'
    spec:
      destinationSelector:
      - kubeServiceRefs:
          services:
          - clusterName: cluster-1
            name: reviews
            namespace: bookinfo
      policy:
        requestTimeout: 5s
      routeSelector:
      - virtualHostSelector:
          namespaces:
          - gloo-mesh
...
```

Similarly, we can look at the status of our `TrafficPolicy` to see what resources it applies to:

```shell
kubectl get -n gloo-mesh trafficpolicy timeout-policy -o yaml
```

```yaml
...
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: timeout-policy
  namespace: gloo-mesh
spec:
  ...truncated for brevity...
status:
  destinations:
    reviews-bookinfo-cluster-1.gloo-mesh.:
      state: ACCEPTED
  gatewayRoutes:
    demo-virtualhost.gloo-mesh (VirtualHost):
      routes:
      - '*'
  observedGeneration: 1
  state: ACCEPTED
  workloads:
  - details-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - details-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - istio-ingressgateway-istio-system-cluster-1-deployment.gloo-mesh.
  - istio-ingressgateway-istio-system-cluster-2-deployment.gloo-mesh.
  - productpage-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - productpage-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - ratings-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - ratings-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - reviews-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v2-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v3-bookinfo-cluster-2-deployment.gloo-mesh.
```

You'll notice that we can see which destinations this policy has been applied to, as well as all workloads that have applied the policy. Now that we've added our `routeSelector`, you can also see which `gatewayRoutes` the policy has been applied to. All `RouteTables` and `VirtualHosts` that the policy has been applied to will show up in this list. Note the `routes` field, with a value of `'*'` - this means the policy has been applied to all fields in the `VirtualHost`.

### Applying a TrafficPolicy to a subset of fields

So far, the `TrafficPolicies` that we have looked at apply to all routes on the selected `VirtualHost` or `RouteTable`. If we want to have more fine-grained control over which routes a policy gets applied to, we can use route labels. Route labels are a set of key/value pairs that are specified at the route level. These labels let you select a subset of routes on a `VirtualHost` or `RouteTable`.

Let's update our `VirtualHost` example to include a route label on one of the two routes:

{{< highlight yaml "hl_lines=14-15" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    labels:
      "retryType": "severe"
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
{{< /highlight >}}

Now let's write a new `TrafficPolicy` which _only_ applies to routes with that label:

{{< highlight yaml "hl_lines=17-18" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: retries-policy
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
        - clusterName: cluster-1
          name: ratings
          namespace: bookinfo
  routeSelector:
    - virtualHostSelector:
        namespaces: 
        - gloo-mesh
      routeLabelMatcher:
        "retryType": "severe"
  policy:
    retries:
      attempts: 10
      perTryTimeout: 2s
{{< /highlight >}}

After applying the new policy, we should see in its status that it only applied to the `ratings` route, rather than all routes (`'*'`). This is because we've added the arbitrary label `retryType: "severe"` to the `ratings` route, and not the `reviews` route.

```shell
kubectl get trafficpolicy -n gloo-mesh retries-policy -o yaml
```

{{< highlight yaml "hl_lines=29-30" >}}
aapiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: retries-policy
  namespace: gloo-mesh
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-1
        name: ratings
        namespace: bookinfo
  policy:
    retries:
      attempts: 10
      perTryTimeout: 2s
  routeSelector:
  - routeLabelMatcher:
      retryType: severe
    virtualHostSelector:
      namespaces:
      - gloo-mesh
status:
  destinations:
    ratings-bookinfo-cluster-1.gloo-mesh.:
      state: ACCEPTED
  gatewayRoutes:
    demo-virtualhost.gloo-mesh (VirtualHost):
      routes:
      - ratings
  observedGeneration: 9
  state: ACCEPTED
  workloads:
  - details-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - details-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - istio-ingressgateway-istio-system-cluster-1-deployment.gloo-mesh.
  - istio-ingressgateway-istio-system-cluster-2-deployment.gloo-mesh.
  - productpage-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - productpage-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - ratings-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - ratings-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v1-bookinfo-cluster-1-deployment.gloo-mesh.
  - reviews-v1-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v2-bookinfo-cluster-2-deployment.gloo-mesh.
  - reviews-v3-bookinfo-cluster-2-deployment.gloo-mesh.
{{< /highlight >}}

Finally if we look at our `VirtualHost`'s status we should see both policies applied, and which routes each applies to:
{{< highlight yaml "hl_lines=35-39 58-62" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - labels:
      retryType: severe
    matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
  - matchers:
    - uri:
        prefix: /reviews
    name: reviews
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: reviews
          namespace: bookinfo
status:
  appliedTrafficPolicies:
  - observedGeneration: 9
    ref:
      name: retries-policy
      namespace: gloo-mesh
    routes:
    - ratings
    spec:
      destinationSelector:
      - kubeServiceRefs:
          services:
          - clusterName: cluster-1
            name: ratings
            namespace: bookinfo
      policy:
        retries:
          attempts: 10
          perTryTimeout: 2s
      routeSelector:
      - routeLabelMatcher:
          retryType: severe
        virtualHostSelector:
          namespaces:
          - gloo-mesh
  - observedGeneration: 2
    ref:
      name: timeout-policy
      namespace: gloo-mesh
    routes:
    - '*'
    spec:
      destinationSelector:
      - kubeServiceRefs:
          services:
          - clusterName: cluster-1
            name: reviews
            namespace: bookinfo
      policy:
        requestTimeout: 5s
      routeSelector:
      - virtualHostSelector:
          namespaces:
          - gloo-mesh
  attachedVirtualGateways:
  - name: demo-gateway
    namespace: gloo-mesh
  observedGeneration: 8


{{< /highlight >}}

### TrafficShift policies with Gateway

One of the features that a `TrafficPolicy` provides, is the ability to shift traffic targeting one destination to another. When it comes to gateway, there is a little bit of nuance involved in traffic shifts, as traffic can be split at two different points.

First, a regular gateway `routeAction` can provide multiple destinations, and specify weights to split the traffic between them. For example:

{{< highlight yaml "hl_lines=20 25" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - www.example.com
  routes:
  - matchers:
    - uri:
        prefix: /some-service
    name: multi-destination-route
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
        weight: 40
      - kubeService:
          clusterName: cluster-2
          name: reviews
          namespace: bookinfo
        weight: 60
{{< /highlight >}}

This will result in 40% of the traffic going to v1 of our `ratings` service, and 60% of the traffic going to our `reviews` service. Since a `TrafficPolicy` also has the ability to split traffic targeting one destination into multiple destinations, it's possible for a `TrafficPolicy` to be applied to one of these routes as well (or instead). This results in the traffic being split even further. Let's look at an example.

Here's an example `TrafficPolicy` which splits traffic to the `reviews`service into 50% `reviews-v2` and 50% `reviews-v3`:

{{< highlight yaml "hl_lines=26 33" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  name: reviews-split-policy
  namespace: gloo-mesh
spec:
  destinationSelector:
  - kubeServiceRefs:
      services:
      - clusterName: cluster-2
        name: reviews
        namespace: bookinfo
  routeSelector:
    - virtualHostSelector:
        namespaces:
          - gloo-mesh
  policy:
    trafficShift:
      destinations:
        - kubeService:
            clusterName: cluster-2
            name: reviews
            namespace: bookinfo
            subset:
              version: v2
          weight: 50
        - kubeService:
            clusterName: cluster-2
            name: reviews
            namespace: bookinfo
            subset:
              version: v3
          weight: 50
{{< /highlight >}}

When the final gateway route `multi-destination-route` (requests with a path prefix of `/some-service`) is created, the split will be 40/30/30 for `ratings`/`reviews-v2`/`reviews-v3`, respectively. That's because at the gateway level, the traffic is split 40/60 between `ratings` and `reviews`. Then once the `TrafficPolicy`'s `trafficShift` is applied, the traffic going to the `reviews` service is split 50/50 between v2 and v3. Note that this `TrafficPolicy` _also_ applies to East/West traffic, meaning requests from within the mesh to the `reviews` service on `cluster-2` will be split 50/50 across reviews-v2 and reviews-v3.

For more information on what kinds of policies can be applied, see the full `TrafficPolicy` API [here]({{% versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.networking.v1.traffic_policy/" %}}).