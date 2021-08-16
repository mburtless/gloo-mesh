---
title: Custom Auth server
weight: 30
description: External Authentication with your own auth server
---

Gloo Mesh Gateway ships with an external auth server that implements a wide array of authentication and authorization models. 
If for some reason you'd rather roll your own external auth server, you can configure Gloo Mesh Gateway to use it to secure your routes.

In this guide we will see how to create such a custom external auth service. For simplicity, we will implement an HTTP 
service. With minor adjustments, you should be able to use the contents of this guide to deploy a gRPC server that implements
the Envoy spec for an [external authorization server](https://github.com/envoyproxy/envoy/blob/master/api/envoy/service/auth/v2/external_auth.proto).

## Setup

First, we need to install Gloo Mesh Enterprise (minimum version `1.1`) with extauth enabled. Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

## Creating a Virtual Gateway
Now let's configure Gloo Mesh Gateway to route requests to the destination we just created. To do that, we define a simple Virtual
Gateway to match all requests that:

- contain a `Host` header with value `www.example.com` and
- have a path that starts with `/ratings`

Apply the following virtual gateway and auth config:
{{< readfile file="guides/gateway/auth/extauth/basic_auth/test-no-auth-vg.yaml" markdown="true">}}

Let's send a request that matches the above route to the Gloo Mesh gateway and make sure it works:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

The above command should produce the following output:

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

## Creating a simple HTTP Authentication service

When a request matches a route that defines an `extauth` configuration, Gloo Mesh will forward the request to the external 
auth service. If the HTTP service returns a `200 OK` response, the request will be considered authorized and sent to 
its original destination. Otherwise, the request will be denied.
You can fine tune which headers are sent to the auth service, and whether the body is forwarded as well, 
by editing the gateway extauth settings in the http route options.

For reference, here's the code for the authorization server used in this tutorial:

```python
import http.server
import socketserver

class Server(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        path = self.path
        print("path", path)
        if path.startswith("/ratings/1"):
            self.send_response(200, 'OK')
        else:
            self.send_response(401, 'Not authorized')
        self.send_header('x-server', 'pythonauth')
        self.end_headers()

def serve_forever(port):
    socketserver.TCPServer(('', port), Server).serve_forever()

if __name__ == "__main__":
    serve_forever(8000)
```

As you can see, this service will allow requests to `/ratings/1` and will deny everything else.

### Deploy auth service

To deploy this service to your cluster, apply the following (assumes bookinfo has auto-injection turned on)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sample-auth
  namespace: bookinfo
  labels:
    app: sample-auth
spec:
  replicas: 1
  selector:
    matchLabels:
      app: sample-auth
  template:
    metadata:
      labels:
        app: sample-auth
    spec:
      containers:
      - name: sample-auth
        image: docker.io/kdorosh/sample-auth:0.0.1
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8000
---
kind: Service
apiVersion: v1
metadata:
  name: auth-service
  namespace: bookinfo
spec:
  selector:
    app: sample-auth
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8000
    name: http # this signals to Gloo Mesh that this is the http port for extauth to hit
```

## Configure Gloo Mesh to use your server

### Configure Gloo Mesh settings

Apply the following `VirtualGateway` to mark it with custom auth:

{{< readfile file="guides/gateway/auth/custom_auth/test-custom-auth-vg.yaml" markdown="true">}}

{{% notice tip %}}
When using a gRPC auth service, remove the `httpService` attribute from the configuration above.
{{% /notice %}}

This configuration also sets other configuration parameters:

- `requestBody` - When this attribute is set, the request body will also be sent to the auth service. With the above configuration, 
a body up to 10KB will be buffered and sent to the service. This is useful in use cases where the request body is relevant 
to the authentication logic, e.g. when it is used to compute an HMAC.
- `requestTimeout` - A timeout for the auth service response. If the service takes longer to respond, the request will be denied.


## Test

Let's verify that our configuration has been accepted by Gloo Mesh. Requests to `/ratings/1` should be allowed:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

Any request with a path that is not `/ratings/1` should be denied.

```shell
curl -v  $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/2 -H "Host: www.example.com"
```

## Conclusion

Gloo Mesh's extensible architecture allows follows the 'batteries included but replaceable' approach.
while you can use Gloo Mesh's built-in auth services for OpenID Connect and Basic Auth, you can also
extend Gloo Mesh with your own custom auth logic.
