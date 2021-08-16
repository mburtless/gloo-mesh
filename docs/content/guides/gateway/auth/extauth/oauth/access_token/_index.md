---
title: Authenticate with an Access Token
weight: 30
description: Integrating Gloo Mesh and Access Tokens
---

{{% notice note %}}
{{< readfile file="static/content/gateway_only_feature_disclaimer" markdown="true">}}
{{% /notice %}}

You may already have an OIDC compliant authentication system in place at your organization which can issue and validate access tokens. In that case, Gloo Mesh can rely on your existing system by accepting requests with an access token and validating that token against an introspection endpoint.

In this guide we will deploy ORY Hydra, a simple OpenID Connect Provider. Hydra will serve as our existing OIDC compliant authentication system. We will generate a valid access token from the Hydra deployment and have Gloo Mesh validate that token using Hyrda's introspection endpoint.

## Setup
First, we need to install Gloo Mesh Enterprise (minimum version `1.1`) with extauth enabled. Please refer to the corresponding
[installation guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) for details.

This guide makes use of the Bookinfo sample application. You can install the application by following the steps in the [Bookinfo deployment section]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployment" %}}).

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

## Securing the Virtual Gateway
As we just saw, we were able to reach our application without having to provide any credentials. This is because by default Gloo Mesh allows any request on routes that do not specify authentication configuration. Let's change this behavior.

We will update the Virtual Gateway so that each request to the sample application is authenticated using an **OpenID Connect** flow.

### Install Hydra
To implement the authentication flow, we need an OpenID Connect provider available to Gloo Mesh. For demonstration purposes, will deploy the [Hydra](https://www.ory.sh/hydra/docs/) provider in the same cluster, as it easy to install and configure.

Let's start by adding the Ory helm repository.

```bash
helm repo add ory https://k8s.ory.sh/helm/charts
helm repo update
```

Now we are going to deploy Hydra using the helm chart:

```bash
helm install \
    --set 'hydra.config.secrets.system=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | base64 | head -c 32)' \
    --set 'hydra.config.dsn=memory' \
    --set 'hydra.config.urls.self.issuer=http://public.hydra.localhost/' \
    --set 'hydra.config.urls.login=http://example-idp.localhost/login' \
    --set 'hydra.config.urls.consent=http://example-idp.localhost/consent' \
    --set 'hydra.config.urls.logout=http://example-idp.localhost/logout' \
    --set 'ingress.public.enabled=true' \
    --set 'ingress.admin.enabled=true' \
    --set 'hydra.dangerousForceHttp=true' \
    hydra-example \
    ory/hydra --version 0.4.5
```

In the above command, we are using an in-memory database of Hydra and setting `hydra.dangerousForceHttp` to `true`, disabling SSL. This is for demonstration purposes and should not be used outside of a development context.

We should now see the two Hydra pods running in the default namespace:

```bash
NAME                                     READY   STATUS    RESTARTS   AGE
hydra-example-58cd5bf699-9jgz5           1/1     Running   0          10m
hydra-example-maester-75c985dd5b-s4b27   1/1     Running   0          10m
```

The administrative endpoint is running on port 4445 and the public endpoint is running on port 4444. We will be using the former to create a client id and password and validate the token, and then the latter to generate an access token.

#### Create the Client and Access Token

Now that we have Hydra up and running, we need to create a client id and client secret by interfacing with the administrative endpoint on Hydra. First we will make the administrative endpoint accessible by forwarding port 4445 of the Hydra pod to our localhost.

```bash
kubectl port-forward deploy/hydra-example 4445 &
portForwardPid1=$! # Store the port-forward pid so we can kill the process later
```

```bash
[1] 1417
~ >Forwarding from 127.0.0.1:4445 -> 4445
```

Now we can use `curl` to interact with the administration endpoint and create a client id and client secret.

```bash
curl -X POST http://127.0.0.1:4445/clients \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json' \
  -d '{"client_id": "my-client", "client_secret": "secret", "grant_types": ["client_credentials"]}'
```

You should see output similar to this:

```json
{"client_id":"my-client","client_name":"","client_secret":"secret","redirect_uris":null,"grant_types":["client_credentials"],"response_types":null,"scope":"offline_access offline openid","audience":null,"owner":"","policy_uri":"","allowed_cors_origins":null,"tos_uri":"","client_uri":"","logo_uri":"","contacts":null,"client_secret_expires_at":0,"subject_type":"public","token_endpoint_auth_method":"client_secret_basic","userinfo_signed_response_alg":"none","created_at":"2020-10-01T19:46:51Z","updated_at":"2020-10-01T19:46:51Z"}
```

Now we will using the public endpoint to generate our access token. First we will port-forward the hydra pod on port 4444.

```bash
kubectl port-forward deploy/hydra-example 4444 &
portForwardPid2=$! # Store the port-forward pid so we can kill the process later
```

```bash
[2] 1431
~ >Forwarding from 127.0.0.1:4444 -> 4444
```

Let's use curl again to create our access token using the client id and secret we just registered on the administrative endpoint.

```bash
curl -X POST http://127.0.0.1:4444/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Accept: application/json' \
  -u my-client:secret \
  -d 'grant_type=client_credentials' | jq .access_token -r
```

The command should render the access token as output which we can set as a variable:

```bash
#OUTPUT
vn83zER2AjyOPbzoVXS3A3S65OCC2LvdGcsz3i5CxlY.NWWWsEixtTLSxN7E0Yk5NsWEZvVZEIjlOCtre0T-s4Q

#SET VARIABLE
ACCESS_TOKEN=vn83zER2AjyOPbzoVXS3A3S65OCC2LvdGcsz3i5CxlY.NWWWsEixtTLSxN7E0Yk5NsWEZvVZEIjlOCtre0T-s4Q
```

We can validate the token using the introspection path of the administrative endpoint:

```bash
curl -X POST http://127.0.0.1:4445/oauth2/introspect \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -H 'Accept: application/json' \
  -d "token=$ACCESS_TOKEN" | jq
```

This is the same path that Gloo Mesh will use to check on the validity of tokens. The next step is to take the introspection URL and add it to an *AuthConfig* and then associate that AuthConfig with the Virtual Gateway we created earlier.

#### Create an AuthConfig

Now that all the necessary resources are in place we can create the `AuthConfig` resource that we will use to secure our Virtual Gateway.

{{< highlight yaml "hl_lines=8-10" >}}
apiVersion: enterprise.gloo.solo.io/v1
kind: AuthConfig
metadata:
  name: oidc-hydra
  namespace: bookinfo
spec:
  configs:
  - oauth2:
      accessTokenValidation:
        introspectionUrl: http://hydra-example-admin.default:4445/oauth2/introspect
{{< /highlight >}}

The above configuration instructs Gloo Mesh to use the `introspectionUrl` to validate access tokens that are submitted with the request. If the token is missing or invalid, Gloo Mesh will deny the request. We can use the internal hostname of the Hydra administrative service, since the request will come from Gloo Mesh's extauth pod which has access to Kubernetes DNS.

#### Update the Virtual Gateway
Once the AuthConfig has been created, we can use it to secure our Virtual Gateway:

{{< readfile file="guides/gateway/auth/extauth/oauth/access_token/test-accesstoken-auth-vg.yaml" markdown="true">}}


### Testing our configuration
The authentication flow to get the access token happens outside of Gloo Mesh's purview. To access the bookinfo site, we will simply include the access token in our request. Gloo Mesh will validate that the token is active using the URL we specified in the AuthConfig.

Now we are ready to test our complete setup! We are going to use `curl` instead of the browser to access the bookinfo page, since we need to include the access token in the request.

First let's try and access the site without a token value set:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" 
```

We will receive a 403 (Forbidden) message letting us know that our access was not authorized. Now let's try an invalid access token value:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: Bearer qwertyuio23456789"
```

Again we will receive a 403 message. Finally, let's try using the access token we generated earlier. Be sure to paste in your proper access token value:

```bash
curl -v $(BOOKINFO_INGRESS_GATEWAY_URL)/ratings/1 -H "Host: www.example.com" -H "Authorization: Bearer $ACCESS_TOKEN"
```

You will receive a 200 HTTP response and the body of the bookinfo homepage. 

### Logging

The ext-auth-service logs can be viewed with:
```
kubectl logs -n gloo-mesh deploy/ext-auth-service -c ext-auth-service -f
```
If the auth config has been received successfully, you should see the log line:
```
"logger":"extauth","caller":"runner/run.go:179","msg":"got new config"
```

### Cleanup
You can clean up the resources created in this guide by running:

```
kill $portForwardPid1
kill $portForwardPid2
kill $portForwardPid3
helm delete --purge hydra-example
kubectl delete vg -n bookinfo test-inlined-gateway
kubectl delete authconfig -n bookinfo oidc-hydra
```

## Summary and Next Steps

In this guide you saw how Gloo Mesh could be used with an existing OIDC system to validate access tokens and grant access to a `VirtualGateway`. You may want to also check out the authentication guides that use [Google]({{< versioned_link_path fromRoot="/guides/gateway/auth/extauth/oauth/google/" >}}) for more alternatives when it comes to OAuth-based authentication.