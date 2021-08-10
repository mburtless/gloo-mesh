---
title: API Keys
weight: 40
description: How to setup ApiKey authentication. 
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

Sometimes when you need to protect a service, the set of users that will need to access it is known in advance and does 
not change frequently. For example, these users might be other services or specific persons or teams in your organization. 

You might also want to retain direct control over how credentials are generated and when they expire. If one of these 
conditions applies to your use case, you should consider securing your service using 
[API keys](https://en.wikipedia.org/wiki/Application_programming_interface_key). API keys are secure, long-lived UUIDs 
that clients must provide when sending requests to a service that is protected using this method. 

{{% notice warning %}}
It is important to note that **your services are only as secure as your API keys**; securing API keys and proper API key 
rotation is up to the user, thus the security of the routes is up to the user.
{{% /notice %}}

To secure your services using API keys, you first need to provide Gloo Mesh with your secret API keys in the form of `Secrets`. After your API key secrets are in place, you can configure authentication on your routes by referencing the secrets in one of two ways:

1. You can specify a **label selector** that matches one or more labelled API key secrets (this is the preferred option), or
1. You can **explicitly reference** a set of secrets by their identifier (namespace and name).

When Gloo Mesh matches a request to a route secured with API keys, it looks for a valid API key in the request headers. 
The name of the header that is expected to contain the API key is configurable. If the header is not present, 
or if the API key it contains does not match one of the API keys in the secrets referenced on the route, 
Gloo Mesh will deny the request and return a 401 response to the downstream client.

Internally, Gloo Mesh will generate a mapping of API keys to _user identities_ for all API keys present in the system. The _user identity_ for a given API key is the name of the `Secret` which contains the API key. The _user identity_ will be added to the request as a header, `x-user-id` by default, which can be utilized in subsequent filters. For security reasons, this header will be sanitized from the response before it leaves the proxy.

## Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`). Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).


## Creating a Virtual Gateway
Now let's configure Gloo Mesh Gateway to route requests to the destination we just created. To do that, we define a simple Virtual
Gateway to match all requests that:

- contain a `Host` header with value `www.example.com` and
- have a path that starts with `/ratings`

{{< readfile file="guides/gateway/auth/extauth/basic_auth/test-no-auth-vg.yaml" markdown="true">}}

Let's send a request that matches the above route to the Gloo Mesh gateway and make sure it works:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

The above command should return:

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

## Securing the Route

As we just saw, we were able to reach the destination without having to provide any credentials. 
This is because by default Gloo Mesh allows any request on routes that do not specify authentication configuration. 
Let's change this behavior. We will update the Virtual Gateway so that only requests containing 
a valid API key are allowed.

We start by creating the following API key secret:

```yaml
apiVersion: v1
kind: Secret
type: extauth.solo.io/apikey
metadata:
  labels:
    team: infrastructure
  name: infra-apikey
  namespace: bookinfo
data:
  api-key: TjJZd01ESXhaVEV0TkdVek5TMWpOemd6TFRSa1lqQXRZakUyWXpSa1pHVm1OamN5
```

The value of `data.api-key` is base64-encoded. You can decode it to get the API key we will use later in requests:

```shell
echo TjJZd01ESXhaVEV0TkdVek5TMWpOemd6TFRSa1lqQXRZakUyWXpSa1pHVm1OamN5 | base64 -D
```

The command should return `N2YwMDIxZTEtNGUzNS1jNzgzLTRkYjAtYjE2YzRkZGVmNjcy`, which is our API key.

Now that we have a valid API key secret, let's go ahead and create an `AuthConfig` Custom Resource (CR) with our API key authentication configuration:

{{< highlight yaml "hl_lines=8-13" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: apikey-auth
  namespace: bookinfo
spec:
  configs:
  - apiKeyAuth:
      # This is the name of the header that is expected to contain the API key.
      # This field is optional and defaults to `api-key` if not present.
      headerName: api-key
      labelSelector:
        team: infrastructure
{{< /highlight >}}

Once the `AuthConfig` has been created, we can use it to secure our Virtual Gateway:

{{< readfile file="guides/gateway/auth/extauth/apikey_auth/test-apikey-auth-vg.yaml" markdown="true">}}

### Testing denied requests
Let's try and resend the same request we sent earlier:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

You will see that the response now contains a **401 Unauthorized** code, indicating that Gloo Mesh denied the request.

{{< highlight shell "hl_lines=6" >}}
> GET /posts/1 HTTP/1.1
> Host: foo
> User-Agent: curl/7.54.0
> Accept: */*
>
< HTTP/1.1 401 Unauthorized
< www-authenticate: API key is missing or invalid
< date: Mon, 07 Oct 2019 19:28:14 GMT
< server: envoy
< content-length: 0
{{< /highlight >}}

### Testing authenticated requests
For a request to be allowed, it must include a header named `api-key` with the value set to the 
API key we previously stored in our secret. Now let's add the authorization headers:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "api-key: N2YwMDIxZTEtNGUzNS1jNzgzLTRkYjAtYjE2YzRkZGVmNjcy"
```

We are now able to reach the ratings service again!

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

## Summary

In this tutorial, we installed Gloo Mesh Gateway and created an unauthenticated Virtual Gateway that routes requests to 
our ratings service. We then created an API key `AuthConfig` object and used it to secure our Virtual Gateway. 
We first showed how unauthenticated requests fail with a `401 Unauthorized` response, and then showed how to send 
authenticated requests successfully to the destination. 

Cleanup the resources by running:

```
kubectl delete ac -n bookinfo apikey-auth
kubectl delete vg -n bookinfo test-inlined-gateway
kubectl delete secret -n bookinfo infra-apikey
```
