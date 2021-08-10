---
title: Passthrough Auth
weight: 10
description: Authenticating using an external grpc service that implements [Envoy's Authorization Service API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter.html). 
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

When using Gloo Mesh's external authentication server, it may be convenient to integrate authentication with a component that implements [Envoy's authorization service API](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/security/ext_authz_filter.html?highlight=authorization%20service#service-definition). This guide will walk through the process of setting up Gloo Mesh's external authentication server to pass through requests to the provided component for authenticating requests. 

## Passthrough auth vs. Custom Auth
You can also implement your own auth with Gloo Mesh with a [Custom Auth server]({{< versioned_link_path fromRoot="/guides/gateway/auth/custom_auth" >}}).

**gRPC Passthrough vs. Custom Auth server**
With gRPC passthrough, you can leverage other Gloo Mesh extauth implementations (e.g. OIDC, API key, etc.) alongside custom logic. A custom auth server is not integrated with Gloo Mesh extauth so it cannot do this.

**gRPC Passthrough Cons**
While using a gRPC Passthrough service does provide additional flexibility and convenience with auth configuration, it does require an additional network hop from Gloo Mesh's external auth service to the gRPC service. 

## Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`). Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

### Creating an authentication service
Currently, Gloo Mesh's external authentication server only supports passthrough requests to a gRPC server. For more information, view the service spec in the [official docs](https://github.com/envoyproxy/envoy/blob/main/api/envoy/service/auth/v3/external_auth.proto).

To use an example gRPC authentication service provided for this guide, run the following command to deploy the provided image. This will run a docker image that contains a simple gRPC service running on port 9001.

{{< highlight yaml >}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: extauth-grpcservice
  namespace: bookinfo # <-- this namespace assumes auto injection is enabled
spec:
  selector:
    matchLabels:
      app: grpc-extauth
  replicas: 1
  template:
    metadata:
      labels:
        app: grpc-extauth
    spec:
      containers:
        - name: grpc-extauth
          image: quay.io/solo-io/passthrough-grpc-service-example
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 9001
{{< /highlight >}}

The source code for the gRPC service can be found in the Gloo Edge repository [here](https://github.com/solo-io/gloo/tree/master/docs/examples/grpc-passthrough-auth).

Once we create the authentication service, we also want to apply the following Service to assign it a static cluster IP.
{{< highlight yaml >}}
apiVersion: v1
kind: Service
metadata:
  name: example-grpc-auth-service
  namespace: bookinfo
  labels:
      app: grpc-extauth
spec:
  ports:
  - port: 9001
    protocol: TCP
  selector:
      app: grpc-extauth
{{< /highlight >}}

## Creating a Virtual Gateway
Now let's configure Gloo Mesh Gateway to route requests to the destination we just created. To do that, we define a simple Virtual
Gateway to match all requests that:

- contain a `Host` header with value `www.example.com` and
- have a path that starts with `/ratings`

Apply the following virtual gateway:
{{< readfile file="guides/gateway/auth/extauth/basic_auth/test-no-auth-vg.yaml" markdown="true">}}

Let's send a request that matches the above route to the Gloo Mesh gateway and make sure it works:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

The above command should produce the following output:

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

If you are getting a connection error, make sure you are port-forwarding the `glooctl proxy url` port to port 8080.

# Securing the Route
As we just saw, we were able to reach the destination without having to provide any credentials. This is because by default 
Gloo Mesh allows any request on routes that do not specify authentication configuration. Let's change this behavior. 
We will update the Virtual Gateway so that all requests will be authenticated by our own auth service.

{{< highlight yaml "hl_lines=8-13" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: passthrough-auth
  namespace: bookinfo
spec:
  configs:
  - passThroughAuth:
      grpc:
        # Address of the grpc auth server to query
        address: example-grpc-auth-service.bookinfo.svc.cluster.local:9001
        # Set a connection timeout to external service, default is 5 seconds
        connectionTimeout: 3s
{{< /highlight >}}

Once the `AuthConfig` has been created, we can use it to secure our Virtual Gateway:

{{< readfile file="guides/gateway/auth/extauth/passthrough_auth/test-passthrough-auth-vg.yaml" markdown="true">}}

### Logging

The extauth server logs can be viewed with:
```
kubectl logs -n gloo-mesh deploy/ext-auth-service -c ext-auth-service -f
```
If the auth config has been received successfully, you should see the log line:
```
"logger":"extauth","caller":"runner/run.go:179","msg":"got new config"
```

## Testing the secured Route
The virtual gateway that we have created should now be secured using our external authentication service. To test this, we can try our original command, and the request should not be allowed through because of missing authentication.

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

The output should be empty. In fact, we can see the 403 (Unauthorized) HTTP status code.

The sample gRPC authentication service has been implemented such that any request with the header `authorization: authorize me` will be authorized. We can easily add this header to our curl request as follows:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "authorization: authorize me"
```

The request should now be authorized!

## Sharing state with other auth steps

A common requirement is to be able to share state between the passthrough service, and other auth steps (either custom plugins, or our built-in authentication) . When writing a custom auth plugin, this is possible, and the steps to achieve it are [outlined here]({{< versioned_link_path fromRoot="/guides/dev/writing_auth_plugins#sharing-state-between-steps" >}}). We support this requirement by leveraging request and response metadata.

We provide some example implementations in the Gloo Edge repository at `docs/examples/grpc-passthrough-auth/pkg/auth/v3/auth-with-state.go`.

### Reading state from other auth steps

State from other auth steps is sent to the passthrough service via [CheckRequest FilterMetadata](https://github.com/envoyproxy/envoy/blob/50e722cbb0486268c128b0f1d0ef76217387799f/api/envoy/service/auth/v3/external_auth.proto#L36) under a unique key: `solo.auth.passthrough`.

### Writing state to be used by other auth steps

State from the passthrough service can be sent to other auth steps via [CheckResponse DynamicMetadata](https://github.com/envoyproxy/envoy/blob/50e722cbb0486268c128b0f1d0ef76217387799f/api/envoy/service/auth/v3/external_auth.proto#L126) under a unique key: `solo.auth.passthrough`.

### Passing in custom configuration to Passthrough Auth Service from AuthConfigs

Custom config can be passed from gloo to the passthrough authentication service. This can be achieved using the `config` field under Passthrough Auth in the AuthConfig:

{{< highlight yaml "hl_lines=14-16" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
  metadata:
    name: passthrough-auth
    namespace: bookinfo
  spec:
    configs:
    - passThroughAuth:
        grpc:
          # Address of the grpc auth server to query
          address: example-grpc-auth-service.default.svc.cluster.local:9001
          # Set a connection timeout to external service, default is 5 seconds
          connectionTimeout: 3s
      config:
        customKey1: "customConfigStringValue"
        customKey2: false
{{< /highlight >}}

This config is accessible via the [CheckRequest FilterMetadata](https://github.com/envoyproxy/envoy/blob/50e722cbb0486268c128b0f1d0ef76217387799f/api/envoy/service/auth/v3/external_auth.proto#L36) under a unique key: `solo.auth.passthrough.config`.

## Summary

In this guide, we installed Gloo Mesh Enterprise and created an unauthenticated Virtual Gateway that routes requests to our ratings service. We spun up an example gRPC authentication service that uses a simple header for authentication. We then created an `AuthConfig` and configured it to use Passthrough Auth, pointing it to the IP of our example gRPC service. In doing so, we instructed gloo mesh to pass through requests from the external authentication server to the grpc authentication service provided by the user.