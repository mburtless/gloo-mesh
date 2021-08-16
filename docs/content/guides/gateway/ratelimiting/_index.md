---
title: Rate Limiting
weight: 30
description: Control the rate of traffic sent to your services.
---

#### Why Rate Limit in API Gateway Environments
API Gateways act as a control point for the outside world to access the various application services
(monoliths, microservices, serverless functions) running in your environment. In microservices or hybrid application
architecture, any number of these workloads will need to accept incoming requests from external end users (clients).
Incoming requests can be numerous and varied -- protecting backend services and globally enforcing business limits
can become incredibly complex being handled at the application level. Using an API gateway we can define client
request limits to these varied services in one place.

#### Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`) with Ratelimit enabled. Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

#### Rate Limiting in Gloo Mesh

Gloo Mesh exposes Envoy's rate-limit API, which allows users to provide their own implementation of an Envoy gRPC rate-limit
service. Lyft provides an example implementation of this gRPC rate-limit service
[here](https://github.com/lyft/ratelimit). To configure Gloo Mesh to use your rate-limit server implementation,
install Gloo Mesh Gateway and then modify the VirtualGateway to use your rate limit server destination:

First create the patch:

```yaml
cat > patch-file.yaml << EOF
spec:
  connectionHandlers:
  - http:
    routeConfig:
      - virtualHost:
        routeOptions:
          rateLimit:
            requestTimeout: 120ms      # optional, default 100ms
            denyOnFail: true           # optional, default false
        ratelimitServerRef:
          name: rate-limiter # rate-limit server destination name
          namespace: gloo-mesh  # rate-limit server upstream namespace
EOF
```

Then apply the patch file with:

```shell script
kubectl patch vg my-virtual-gateway --namespace gloo-mesh --type merge --patch "$(cat patch-file.yaml)"
```

Gloo Mesh Enterprise provides an enhanced version of [Lyft's rate limit service](https://github.com/lyft/ratelimit) that
supports the full Envoy rate limit server API (with some additional enhancements, e.g. rule priority), as well as a
simplified API built on top of this service. Gloo Mesh uses this rate-limit service to enforce rate-limits.

### Logging

If Gloo Mesh is running on kubernetes, the rate limiting logs can be viewed with:
```
kubectl logs -n gloo-mesh deploy/rate-limiter -f
```

Note that these logs will only be in agent clusters.

When it starts up correctly, you should see a log line similar to:
```
"caller":"server/server_impl.go:48","msg":"Listening for HTTP on ':18080'"
```

### Rate Limit Configuration

Check out the guides for each of the Gloo Mesh rate-limit APIs and configuration options for Gloo Mesh Enterprise's rate-limit
service:

{{% children description="true" %}}
