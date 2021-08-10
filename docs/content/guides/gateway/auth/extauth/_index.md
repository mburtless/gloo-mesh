---
title: External Authentication
menuTitle: Ext Auth
weight: 20
description: Authenticate and authorize requests to your services using Gloo Mesh's external auth service.
---

{{% notice note %}}
This section refers specifically to the **Gloo Mesh Enterprise** external auth server.
{{% /notice %}}

Gloo Mesh Gateway provides a variety of authentication options to meet the needs of your environment. They range from supporting basic use cases to complex and fine-grained secure access control. Architecturally, Gloo Mesh uses a dedicated auth server to verify the user credentials and determine their permissions. Gloo Mesh provides an auth server that can support several authN/Z implementations and also allows you to provide your auth server to implement custom logic.

While some authentication solutions, such as JWT verification, can occur directly in Envoy, many use cases are better served by an external service. Envoy supports an external auth filter, where it reaches out to another service to authenticate and authorize a request, as a general solution for handling a large number of auth use cases at scale. Gloo Mesh Enterprise comes with an external auth (Ext Auth) server that has built-in support for all standard authentication and authorization use cases, and a plugin and passthrough framework for customization.
### Implementations

We have seen how `AuthConfigs` can be used to define granular authentication configurations for routes. For a detailed overview of the authN/authZ models implemented by Gloo Mesh Gateway, check out the other guides:

{{% children description="true" %}}

### Extauth within or external to the mesh

By default, extauth installs as a part of your mesh. This means you get all the benefits of being part of the mesh (observability, telemetry, mtls, etc.), but also means you get the drawbacks (extra network hop). The location for the extauth service can be configured to use an external service destination, i.e., external to the mesh.

### Auth Configuration Overview

Authentication configuration is defined in `AuthConfig` resources. `AuthConfig`s are top-level resources, which means that if you are running in Kubernetes, they will be stored in a dedicated CRD. Here is an example of a simple `AuthConfig` CRD:

```yaml
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: basic-auth
  namespace: gloo-system
spec:
  configs:
  - basicAuth:
      apr:
        users:
          user:
            hashedPassword: 14o4fMw/Pm2L34SvyyA2r.
            salt: 0adzfifo
      realm: gloo
```

The resource `spec` consists of an array of `configs` that will be executed in sequence when a request matches the `AuthConfig` (more on how requests are matched to `AuthConfigs` shortly). If any one of these "authentication steps" fails, the request will be denied (this behavior is [configurable](https://github.com/solo-io/solo-apis/blob/v1.6.31/pkg/api/enterprise.gloo.solo.io/v1/auth_config.pb.go#L154-L157)). In most cases an `AuthConfig` will contain a single `config`.

#### Configuration Format

Once an `AuthConfig` is created, it can be used to authenticate your routes. You can define authentication configuration your gateway at three different levels:

- on **VirtualGateways**
- on **VirtualHosts**, and
- on **Routes**

The configuration format is the same in all three cases. It must be specified under the relevant `options` attribute and can take one of two forms. The first is used to enable authentication and requires you to reference an existing `AuthConfig`. An example configuration of this kind follows:

```yaml
options:
  extauth:
    configRef:
      # references the example AuthConfig we defined earlier
      name: basic-auth
      namespace: gloo-system
```

In the case of a route or weighted destination, the top attribute would be named `options` as well.

The second form is used to disable authentication explicitly:

```yaml
options:
  extauth:
    disable: true
```
