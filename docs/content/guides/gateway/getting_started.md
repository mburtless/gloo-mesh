---
title: Getting Started
menuTitle: Getting Started
description: Introductory guide for getting started with Gloo Mesh Gateway
weight: 20
---

{{% notice note %}} Basic Gateway features are available to anyone with a Gloo Mesh Enterprise license. Some advanced features require a Gloo Mesh Gateway license. {{% /notice %}}

## Before you begin

This guide assumes the following:

  * Gloo Mesh Enterprise is [installed in relay mode and running on the `cluster-1`]({{% versioned_link_path fromRoot="/setup/installation/enterprise_installation/" %}})
  * `gloo-mesh` is the installation namespace for Gloo Mesh
  * `enterprise-networking` is deployed on `cluster-1` in the `gloo-mesh` namespace and exposes its gRPC server on port 9900
  * `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
  * Both `cluster-1` and `cluster-2` are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
  * Istio is [installed on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
  * `istio-system` is the root namespace for both Istio deployments
  * `istio-ingressgateway` is deployed on `cluster-1` (this is the default with an istio install)
  * The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace
  * the following environment variables are set:
    ```shell
    CONTEXT_1=cluster_1_context
    CONTEXT_2=cluster_2_context
    ```
  * It is recommended you read the [Gateway Concepts Overview]({{% versioned_link_path fromRoot="/guides/gateway/concepts" %}}) beforehand to understand the custom resources being used in this guide.


## Environment Overview

Assuming we have all of the above set up from the previous guides, our environment should look like this:

![Bookinfo Mutlicluster]({{% versioned_link_path fromRoot="/img/gateway/bookinfo-topology.png" %}})

We have most of the `bookinfo` app installed on `cluster1`, with the exception of `reviews-v3` which is deployed exclusively on `cluster2`. We have a copy of `ratings-v1` deployed on both clusters. Also note that we are starting off with an `istio-ingressgateway` installed in `cluster1`, but no routes are configured yet. We will use the Gloo Mesh Gateway API to configure them in this guide.

## Basic Route Setup

To start, we are going to set up a route to call the `ratings` service from `bookinfo` on `cluster-1`. In order to do this, we will need to create a VirtualGateway resource. At a minimum, we'll need to specify which gateway workload that this configuration will be used by, which port the listener will run on, and what kind of listener we want (eg HTTP).

In our case, we are selecting a single gateway - the `istio-ingressgateway` in `cluster-1`. Later we will see that this configuration can also be deployed to multiple gateways at once. Note we're also specifing port `8081` for our `HTTP` listener.

Finally, note that we've in-lined our virtualHost, and added a route to `/ratings` for host `www.example.com` which routes traffic to our `ratings` service.

Run the command below to create our `demo-gateway` VirtualGateway resource:

```shell
cat << EOF | kubectl apply --context $CONTEXT_1 -f -
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  deployToIngressGateways:
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
    bindPort: 8081
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
                  clusterName: cluster-1
                  name: ratings
                  namespace: bookinfo
EOF
```

Your istio ingressgateway should be exposed as a LoadBalancerIP service, which you can check by running the command below:

```shell
kubectl get service -n istio-system
```

You should be able to the load balancer's external IP address in the output, in the example below it's `32.12.34.555`:

{{< highlight shell "hl_lines=2" >}}
NAME                   TYPE           CLUSTER-IP      EXTERNAL-IP    PORT(S)                                                                      AGE
istio-ingressgateway   LoadBalancer   10.96.229.177   32.12.34.555   15021:31911/TCP,80:30166/TCP,443:32302/TCP,15012:30471/TCP,15443:31931/TCP   10m
istiod                 ClusterIP      10.96.180.254   <none>         15010/TCP,15012/TCP,443/TCP,15014/TCP                                        10m
{{< /highlight >}}

You may see `<pending>` instead of an External IP address depending on which platform you are running Kubernetes on. For example, if you are testing this out locally in KinD or minikube, they won't provide you with an external IP address. Similarly if you are running on a platform like GKE and you don't have sufficient permissions or resources, you may not be assigned an IP address. If you run into this problem and are walking through this tutorial to familiarize yourself with the system, rather than a secure production deployment, you can simply skip the load balancer by port-forwarding the service port of the ingress:

```shell
kubectl --context $CONTEXT_1 -n istio-system port-forward deploy/istio-ingressgateway 8081
```

Now we should be able to hit our ratings service, remembering to pass our hostname as a header so that the route matches, and assuming everything's working correctly, we should get the response from our ratings service:

```shell
curl -H "Host: www.example.com" localhost:8081/ratings/1
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

Congratulations, you have just configured your first route with Gloo Mesh Gateway!

## Splitting out the VirtualHost

We just saw a basic example where the VirtualHost was in-lined inside of the VirtualGateway. This works great for simple set-ups or quickly trying new things. However, as your deployment grows in complexity and you find yourself adding more routes, more route options, more matchers, etc - having both the Gateway and the VirtualHost config in a single resource can become a little unwieldy. The solution to this is to break up the VirtualGateway and the VirtualHost into two separate resources. Not only is this easier to maintain, but it also allows organizations to split responsability of gateway configuration and host configuration across multiple teams, if desired.

From a configuration perspective, the VirtualHost looks very similar to the settings we in-lined in our first example, we simply reference it using a `virtualHostSelector` in the VirtualGateway now instead:

{{< highlight yaml "hl_lines=10-12" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - "gloo-mesh"
  deployToIngressGateways:
    bindPort: 8081
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system

---
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
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
{{< /highlight >}}

Functionally, this is equivalent to our first basic example, but it allows us a lot more flexibility with our configuration. This is the recommended pattern for production deployments.


## Routing to multiple clusters

So far we have routed traffic to services in the same cluster that the gateway itself lives in. Routing traffic to services in other clusters is just as simple however. If you've been following our examples so far, you will have a VirtualHost called `demo-virtualhost` which routes to the ratings service in `cluster-1`. In our environment, we also happen to have a copy of this service living in cluster-2. If we want to route to it, it's as simple as changing the `clusterName` in the `kubeService` definition in our `VirtualHost`:

{{< highlight yaml "hl_lines=17" >}}
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
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-2
          name: ratings
          namespace: bookinfo
{{< /highlight >}}

Similarly, we can easily split traffic across both clusters, and specify weights for the traffic ratio sent to each:


{{< highlight yaml "hl_lines=15-25" >}}
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
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
        weight: 75
      - kubeService:
          clusterName: cluster-2
          name: ratings
          namespace: bookinfo
        weight: 25
{{< /highlight >}}

The above will result in about three quarters of the traffic going to the `ratings` service in `cluster-1`, with the remaining quarter going to the `ratings` service in `cluster-2`. Note, these weights are relative, and if ommitted then traffic will be split evenly across all destinations.

## Delegating to a RouteTable

Above, we have separated the `VirtualHost` out from the `VirtualGateway`. We can also optionally go one step further and delegate some of the routing to a separate `RouteTable` object. Here's an example of the same configuration above, but split across a `VirtualGateway`, `VirtualHost`, and `RouteTable`:

```yaml
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - "gloo-mesh"
  deployToIngressGateways:
    bindPort: 8081
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system

---
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
        prefix: /
    delegateAction:
      selector:
        namespaces:
        - "gloo-mesh"


---
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
        weight: 75
      - kubeService:
          clusterName: cluster-2
          name: ratings
          namespace: bookinfo
        weight: 25
```

Functionally this is still the same routing configuration as our `VirtualGateway`/`VirtualHost` configuration before, but we have now given ourselves the flexibility to break up the configuration into different logical areas. The team configuring the RouteTable may be better suited to determine all of the routes within their application, and the operations team may want to keep control over which domains that application is served on, for example.

## Configure Route Options

Routes can have policies configured using route options. Anything that can be set in a Gloo Mesh `TrafficPolicy` can also be set here. For example, here is some configuration to add an arbitrary header to the response:


{{< highlight yaml "hl_lines=18-21" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
  routes:
  - matchers:
    - uri:
        prefix: /ratings
    name: ratings
    routeAction:
      destinations:
      - kubeService:
          clusterName: cluster-1
          name: ratings
          namespace: bookinfo
    options:
      headerManipulation:
        appendResponseHeaders:
          "x-my-custom-header": "example"
{{< /highlight >}}

Configuring these route policies at the route level is simply one option. Alternatively, or in addition, these options can be configured at the `VirtualHost` level, or from a `TrafficPolicy` resource.

Here is how we would set the same options using a TrafficPolicy instead:


{{< highlight yaml "hl_lines=19-32" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: RouteTable
metadata:
  name: demo-routetable
  namespace: gloo-mesh
spec:
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
---
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: add-header-policy
spec:
  routeSelector:
  - routeTableRefs:
    - name: demo-routetable
    - namespace: gloo-mesh
  policy:
    headerManipulation:
      appendResponseHeaders:
        "x-my-custom-header": "example"
{{< /highlight >}}

The benefit of defining such policies in a `TrafficPolicy` resource, is that they can be reused to configure both ingress traffic and east-west traffic within the mesh. To accomplish this policy sharing, we simply need to ensure the `TrafficPolicy` has a `routeSelector` for ingress traffic, and a `destinationSelector` for east-west traffic:

{{< highlight yaml "hl_lines=7-16" >}}
apiVersion: networking.mesh.gloo.solo.io/v1
kind: TrafficPolicy
metadata:
  namespace: gloo-mesh
  name: add-header-policy
spec:
  destinationSelector:
  - kubeServiceRefs:
    services:
      - clusterName: cluster-1
        name: ratings
        namespace: bookinfo
  routeSelector:
  - routeTableRefs:
    - name: demo-routetable
    - namespace: gloo-mesh
  policy:
    headerManipulation:
      appendResponseHeaders:
        "x-my-custom-header": "example"
{{< /highlight >}}

## Multiple Ingress Gateways

If we look at our `VirtualGateway`, under `deployToIngressGateways.gatewayWorkloads.kubeWorkloadMatcher.clusters` , we see that we're matching the istio ingress-gateway on `cluster-1`. Notice however that this selector is an array. If we add the istio ingressgateway from `cluster-2`, these gateway settings will also apply there. If you send requests to this second gateway, you should still be able to hit all of the routes, regardless of what cluster the service is ultimately served from.

Here's our updated VirtualGateway which now configures the ingress gateways on both `cluster-1` and `cluster-2`:

{{< highlight yaml "hl_lines=19" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  connectionHandlers:
  - http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - "gloo-mesh"
  deployToIngressGateways:
    bindPort: 8081
    gatewayWorkloads:
    - kubeWorkloadMatcher:
        clusters:
        - cluster-1
        - cluster-2
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
{{< /highlight >}}

You should now be able to make requests (eg using `curl`) to either ingress gateway, and both will have identical routes configured, even though the routes may point to services split across both clusters.