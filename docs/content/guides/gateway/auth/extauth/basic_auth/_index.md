---
title: Basic Auth
weight: 10
description: Authenticating using a dictionary of usernames and passwords on a virtual gateway. 
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

In certain cases - such as during testing or when releasing a new API to a small number of known users - it may be 
convenient to secure gateway routes using [Basic Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication). 
With this simple authentication mechanism the encoded user credentials are sent along with the request in a standard header.

To secure your routes using basic authentication, you first need to provide Gloo Mesh Gateway with a set of known users and 
their passwords. You can then use this information to decide who is allowed to access which routes.
If a request matches a route on which basic authentication is configured, Gloo Mesh Gateway will verify the credentials in the 
standard `Authorization` header before sending the request to its destination. If the user associated with the credentials 
is not explicitly allowed to access that route, Gloo Mesh Gateway will return a 401 response to the downstream client.


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

## Securing the Route

As we just saw, we were able to reach the destination without having to provide any credentials. This is because by default 
Gloo Mesh Gateway allows any request on routes that do not specify authentication configuration. Let's change this behavior. 
We will update the Virtual Gateway so that only requests by the user `user` with password `password` are allowed.
Gloo Mesh expects password to be hashed and [salted](https://en.wikipedia.org/wiki/Salt_(cryptography)) using the
[APR1](https://httpd.apache.org/docs/2.4/misc/password_encryptions.html) format. Passwords in this format follow this pattern:

> $apr1$**SALT**$**HASHED_PASSWORD**

To generate such a password you can use the `htpasswd` utility:

```shell
htpasswd -nbm user password
```

Running the above command returns a string like `user:$apr1$TYiryv0/$8BvzLUO9IfGPGGsPnAgSu1`, where:

- `TYiryv0/` is the salt and
- `8BvzLUO9IfGPGGsPnAgSu1` is the hashed password.

Now that we have a password in the required format, let's go ahead and create an `AuthConfig` CR and update our virtual
gateway to reference the auth config:

{{< readfile file="guides/gateway/auth/extauth/basic_auth/test-basic-auth-vg.yaml" markdown="true">}}


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
< www-authenticate: Basic realm=""
< date: Mon, 07 Oct 2019 13:36:58 GMT
< server: envoy
< content-length: 0
{{< /highlight >}}

### Testing authenticated requests
For a request to be allowed, it must now include the user credentials inside the expected header, which has the 
following format:

```
Authorization: basic <base64_encoded_credentials>
```

To encode the credentials, just run:

```shell
echo -n "user:password" | base64
```

For our example, this outputs `dXNlcjpwYXNzd29yZA==`. Let's include the header with this value in our request:

```shell
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: basic dXNlcjpwYXNzd29yZA=="
```

We are now able to reach the ratings service again!

```json
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

### Logging

The ext-auth-service logs can be viewed with:
```
kubectl logs -n gloo-mesh deploy/ext-auth-service -c ext-auth-service -f
```
If the auth config has been received successfully, you should see the log line:
```
"logger":"extauth","caller":"runner/run.go:179","msg":"got new config"
```

## Summary

In this tutorial, we installed Gloo Mesh Gateway and created an unauthenticated Virtual Gateway that routes requests to
our ratings service. We then created a Basic Authentication `AuthConfig` object and used it to secure our Virtual Gateway. 
We first showed how unauthenticated requests fail with a `401 Unauthorized` response, and then showed how to send 
authenticated requests successfully to the destination. 

Cleanup the resources by running:

```
kubectl delete ac -n bookinfo basic-auth
kubectl delete vg -n bookinfo test-inlined-gateway
```
