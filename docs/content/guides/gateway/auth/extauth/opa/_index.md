---
title: OPA Authorization
weight: 50
description: Illustrating how to combine OpenID Connect with Open Policy Agent to achieve fine-grained policy with Gloo Mesh.
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

The [Open Policy Agent](https://www.openpolicyagent.org/) (OPA) is an open source, general-purpose policy engine that can be used to define and enforce versatile policies in a uniform way across your organization. Compared to an RBAC authorization system, OPA allows you to create more fine-grained policies. For more information, see [the official docs](https://www.openpolicyagent.org/docs/latest/comparison-to-other-systems/).

## Table of Contents
- [Setup](#setup)
- [OPA policy overview](#opa-policy-overview)
    - [OPA input structure](#opa-input-structure)
- [Validate requests attributes with Open Policy Agent](#validate-requests-attributes-with-open-policy-agent)
    - [Creating a Virtual Gateway](#creating-a-virtual-gateway)
    - [Secure the Virtual Gateway](#securing-the-virtual-gateway)
        - [Define an OPA policy](#define-an-opa-policy)
        - [Create an OPA AuthConfig CRD](#create-an-opa-authconfig-crd)
        - [Update the Virtual Gateway](#updating-the-virtual-gateway)
    - [Testing our configuration](#testing-the-configuration)

## Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`) with extauth enabled. Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).


## OPA policy overview
Open Policy Agent policies are written in [Rego](https://www.openpolicyagent.org/docs/latest/how-do-i-write-policies/). The _Rego_ language is inspired from _Datalog_, which in turn is a subset of _Prolog_. _Rego_ is more suited to work with modern JSON documents.

Gloo Mesh's OPA integration will populate an `input` document which can be used in your OPA policies. The structure of the `input` document depends on the context of the incoming request. See the following section for details.

### OPA input structure
- `input.check_request` - By default, all OPA policies will contain an [Envoy Auth Service `CheckRequest`](https://www.envoyproxy.io/docs/envoy/latest/api-v2/service/auth/v2/external_auth.proto#service-auth-v2-checkrequest). This object contains all the information Envoy has gathered of the request being processed. See the Envoy docs and [proto files for `AttributeContext`](https://github.com/envoyproxy/envoy/blob/b3949eaf2080809b8a3a6cf720eba2cfdf864472/api/envoy/service/auth/v2/attribute_context.proto#L39) for the structure of this object.
- `input.http_request` - When processing an HTTP request, this field will be populated for convenience. See the [Envoy `HttpRequest` docs](https://www.envoyproxy.io/docs/envoy/latest/api-v2/service/auth/v2/attribute_context.proto#service-auth-v2-attributecontext-httprequest) and [proto files](https://github.com/envoyproxy/envoy/blob/b3949eaf2080809b8a3a6cf720eba2cfdf864472/api/envoy/service/auth/v2/attribute_context.proto#L90) for the structure of this object.
- `input.state.jwt` - When the [OIDC auth plugin]({{< versioned_link_path fromRoot="/guides/gateway/auth/extauth/oauth/" >}}) is utilized, the token retrieved during the OIDC flow is placed into this field. See the section below on [validating JWTs](#validate-jwts-with-open-policy-agent) for an example.

# Validate requests attributes with Open Policy Agent

#### Creating a Virtual Gateway
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

### Securing the Virtual Gateway

As we just saw, we were able to reach the destination without having to provide any credentials. This is because by default Gloo Mesh allows any request on routes that do not specify authentication configuration. Let's change this behavior. We will update the Virtual Gateway so that only requests that comply with a given OPA policy are allowed.

#### Define an OPA policy 

Let's create a Policy to control which actions are allowed on our service:

```shell
cat <<EOF > policy.rego
package test

default allow = false
allow {
    startswith(input.http_request.path, "/ratings/2")
    input.http_request.method == "GET"
}
allow {
    input.http_request.path == "/ratings/3"
    any({input.http_request.method == "GET",
        input.http_request.method == "DELETE"
    })
}
EOF
```

This policy:

- denies everything by default,
- allows requests if:
  - the path starts with `/ratings/2` AND the http method is `GET` **OR**
  - the path is exactly `/ratings/3` AND the http method is either `GET` or `DELETE`

#### Create an OPA AuthConfig CRD
Gloo Mesh expects OPA policies to be stored in a Kubernetes ConfigMap, so let's go ahead and create a ConfigMap with the contents of the above policy file:

```
kubectl -n bookinfo create configmap allow-get-users --from-file=policy.rego
```

Now we can create an `AuthConfig` CRD with our OPA authorization configuration:

{{< highlight yaml "hl_lines=8-12" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: opa
  namespace: bookinfo
spec:
  configs:
  - opa_auth:
      modules:
      - name: allow-get-users
        namespace: bookinfo
      query: "data.test.allow == true"
{{< /highlight >}}

The above `AuthConfig` references the ConfigMap  (`modules`) we created earlier and adds a query that allows access only if the `allow` variable is `true`. 

#### Updating the Virtual Gateway
Once the `AuthConfig` has been created, we can use it to secure our Virtual Gateway:

{{< readfile file="guides/gateway/auth/extauth/opa/test-opa-auth-vg.yaml" markdown="true">}}

### Testing the configuration
Paths that start with `/ratings/1` are not authorized (should return 403):
```
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com"
```

Allowed to get `ratings/2`:
```
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/2 -H "Host: www.example.com"
```

#### Cleanup
You can clean up the resources created in this guide by running:

```
kubectl delete vg -n bookinfo test-inlined-gateway
kubectl delete ac -n bookinfo opa
kubectl delete cm -n bookinfo allow-get-users
rm policy.rego
```
