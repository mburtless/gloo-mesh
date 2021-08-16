---
title: Secure Gateways
weight: 30
description: How to up HTTPS on an ingress gateway
---

This guide will walk you through configuring SSL on your `VirtualGateway`, so that your gateway can serve HTTPS traffic.

### Before you begin

This guide assumes the following:

  * Gloo Mesh Enterprise is [installed in relay mode and running on the `cluster-1`]({{% versioned_link_path fromRoot="/setup/installation/enterprise_installation/" %}})
  * `gloo-mesh` is the installation namespace for Gloo Mesh
  * `enterprise-networking` is deployed on `cluster-1` in the `gloo-mesh` namespace and exposes its gRPC server on port 9900
  * `enterprise-agent` is deployed on both clusters and exposes its gRPC server on port 9977
  * Both `cluster-1` and `cluster-2` are [registered with Gloo Mesh]({{% versioned_link_path fromRoot="/guides/#two-registered-clusters" %}})
  * Istio is [installed on both clusters]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) clusters
  * `istio-system` is the root namespace for both Istio deployments
  * `istio-ingressgateway` is deployed on `cluster-1` in the `istio-system` namespace (this is the default with an istio install)
  * The `istio-ingressgateway` `Deployment` has port `443` open on the container
  * The `bookinfo` app is [installed into the two clusters]({{% versioned_link_path fromRoot="/guides/#bookinfo-deployed-on-two-clusters" %}}) under the `bookinfo` namespace
  * The following environment variables are set:
    ```shell
    CONTEXT_1=cluster_1_context
    CONTEXT_2=cluster_2_context
    ```
  * It is recommended you read the [Gateway Concepts Overview]({{% versioned_link_path fromRoot="/guides/gateway/concepts" %}}) beforehand to understand the custom resources being used in this guide.
  * If you are planning to generate your own certs during this walkthrough, you'll need to have the `openssl` CLI tool installed.

### Creating Certificates and Keys

You are most likely bringing your own set of certs & keys that you'd like to use for TLS on your edge gateway. If you need to create your own for demo purposes or for a test environment, you can follow the steps below. If you already have certs & keys, you can skip forward to the next section.

First, let's set up some environment variables representing our certificate name, the domain the cert will be valid for:

```shell
ROOT_CERT_NAME=gateway-root
DNS_NAME='www.example.com'
SERVER_CERT_NAME=gw-ssl
GATEWAY_NAMESPACE=istio-system
GATEWAY_KUBE_CONTEXT=kind-cluster-1
```

Next, we'll be creating our root cert:
```shell
# root cert
openssl req -new -newkey rsa:4096 -x509 -sha256 \
	-days 3650 -nodes -out ${ROOT_CERT_NAME}.crt -keyout ${ROOT_CERT_NAME}.key \
	-subj "/CN=www.example.com/O=${ROOT_CERT_NAME}" \
	-addext "subjectAltName = DNS:www.example.com"
```

Finally, we use the root cert to create the server cert:
```shell
# server cert
cat > "${SERVER_CERT_NAME}.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS = ${DNS_NAME}
EOF


openssl genrsa -out "${SERVER_CERT_NAME}.key" 2048
openssl req -new -key "${SERVER_CERT_NAME}.key" -out ${SERVER_CERT_NAME}.csr -subj "/CN=${DNS_NAME}/O=${SERVER_CERT_NAME}" -config "${SERVER_CERT_NAME}.conf"
openssl x509 -req \
-days 3650 \
-CA ${ROOT_CERT_NAME}.crt -CAkey ${ROOT_CERT_NAME}.key \
-set_serial 0 \
-in ${SERVER_CERT_NAME}.csr -out ${SERVER_CERT_NAME}.crt \
-extensions v3_req -extfile "${SERVER_CERT_NAME}.conf"
```

### Adding the cert to the cluster

Now that we have our server cert from the steps above, we need to be able to use it in our service mesh. To do this, we use a Kubernetes `Secret`. It is important to note that the secret must be added to the same cluster that the ingress gateway is deployed to, and it must also be in the same `namespace` as the ingress gateway's `Deployment` resource.

Let's create a Kubernetes secret from our server cert - and apply that secret to the cluster & namespace that contain our gateway deployment:
```shell
# create secret from server cert
kubectl create secret generic ${SERVER_CERT_NAME}-secret \
--from-file=tls.key=${SERVER_CERT_NAME}.key \
--from-file=tls.crt=${SERVER_CERT_NAME}.crt \
--dry-run=client -oyaml | kubectl apply -f- \
--context ${GATEWAY_KUBE_CONTEXT} \
--namespace ${GATEWAY_NAMESPACE}
```

{{% notice note %}} If you are running multiple ingress gateways in multiple clusters or namespaces, _each_ cluster will need a copy of this secret in _each_ namespace that an ingress gateway is installed, assuming you want to use the same SSL config for each. The name of the secret must be the same in all clusters, although it can live in different namespaces as appropriate. {{% /notice %}}

### Configuring TLS on the gateway

Now that we have our cert installed into our cluster as a Kubernetes secret, we can configure our gateway to use it. Here is an example `VirtualGateway` with ssl configured to handle https traffic. Specifically, it will handle traffic to `www.example.com`, and on the route `/ratings`, meaning this should serve https requests to `https://www.example.com/ratings/1`, for example. Note that we are specifically selecting the gateway destination on `portName: https` here, rather than the `http2` used in previous examples:

{{< highlight yaml "hl_lines=9-13 15-17 32" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: demo-gateway
  namespace: gloo-mesh
spec:
  connectionHandlers:
  - connectionOptions:
      sslConfig:
        # Note this secret must be located on the same cluster
        # in the same namespace as the gateway deployment
        secretName: gw-ssl-secret
        tlsMode: SIMPLE
    connectionMatch: 
      serverNames:
      # This SNI should match the DNS name your cert is using
      - www.example.com
    http:
      routeConfig:
      - virtualHostSelector:
          namespaces:
          - "gloo-mesh"
  ingressGatewaySelectors:
  - destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
    portName: https
---

# No TLS-specific config needed on the VirtualHost
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualHost
metadata:
  name: demo-virtualhost
  namespace: gloo-mesh
spec:
  domains:
  - "www.example.com"
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

### Testing it out

If you're using your own certs with your own domain, and DNS etc is all set up, and port 443 is exposed as a LoadBalancer service on your cluster sending traffic to your ingress gateway, then you should now be able to curl your https address and see the full flow working. Try it out with: `curl https://your-domain.com/ratings/1`.

If you are using the certs generated at the start of this guide, or you don't yet have DNS etc set up, you can still test and see this working end to end with a few tweaks. First, if your ingress gateway is not accessible via a `LoadBalancer` or `NodePort` service, you can port forward directly with a port forward. In our case, our `istio-ingressgateway` is serving this traffic on port `443`, as we specified in the `bindPort` above. We will be port forwarding port `443` to access this locally, however you can use a port number greater than `1024` if you don't have sufficient admin priviledges to listen on port `443` locally.

{{< tabs >}}
{{< tab name="With Sudo" codelang="shell" >}}
# Note: You need sudo priviliges to listen on port 443 locally:
sudo kubectl port-forward -n istio-system service/istio-ingressgateway 443
{{< /tab >}}
{{< tab name="Without Sudo" codelang="shell" >}}
# Listen on port 8443 locally:
kubectl port-forward -n istio-system service/istio-ingressgateway 8443:443
{{< /tab >}}
{{< /tabs >}}


Next, in order to resolve your domain for the scope of the curl request, you can add the parameters below, replacing `www.example.com` with whichever domain you have been using throughout this guide:


{{< tabs >}}
{{< tab name="With Sudo" codelang="shell" >}}
# Note: Assumes you are listening on port 443 locally
curl -k --resolve www.example.com:443:127.0.0.1 https://www.example.com/ratings/1
{{< /tab >}}
{{< tab name="Without Sudo" codelang="shell" >}}
# Note: Assumes you are listening on port 8443 locally
curl -k --resolve www.example.com:8443:127.0.0.1 https://www.example.com:8443/ratings/1
{{< /tab >}}
{{< /tabs >}}


If everything works, you should get a response from our ratings service:
```shell
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}
```

We can also print more verbose output by adding the -v arg to our curl command, which makes it even more clear that we're using TLS in our request:

{{< tabs >}}
{{< tab name="With Sudo" codelang="shell" >}}
# Note: Assumes you are listening on port 443 locally
curl -vik --resolve www.example.com:443:127.0.0.1 https://www.example.com/ratings/1
* Added www.example.com:443:127.0.0.1 to DNS cache
* Hostname www.example.com was found in DNS cache
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to www.example.com (127.0.0.1) port 443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/cert.pem
  CApath: none
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=www.example.com; O=gw-ssl
*  start date: Aug  6 18:27:18 2021 GMT
*  expire date: Aug  4 18:27:18 2031 GMT
*  issuer: CN=www.example.com; O=gateway-root
*  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7fbae280d600)
> GET /ratings/1 HTTP/2
> Host: www.example.com
> User-Agent: curl/7.64.1
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS == 2147483647)!
< HTTP/2 200
HTTP/2 200
< content-type: application/json
content-type: application/json
< date: Fri, 06 Aug 2021 21:45:35 GMT
date: Fri, 06 Aug 2021 21:45:35 GMT
< x-envoy-upstream-service-time: 2
x-envoy-upstream-service-time: 2
< server: istio-envoy
server: istio-envoy

<
* Connection #0 to host www.example.com left intact
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}* Closing connection 0
{{< /tab >}}
{{< tab name="Without Sudo" codelang="shell" >}}
# Note: Assumes you are listening on port 8443 locally
curl -vik --resolve www.example.com:8443:127.0.0.1 https://www.example.com:8443/ratings/1
* Added www.example.com:8443:127.0.0.1 to DNS cache
* Hostname www.example.com was found in DNS cache
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to www.example.com (127.0.0.1) port 8443 (#0)
* ALPN, offering h2
* ALPN, offering http/1.1
* successfully set certificate verify locations:
*   CAfile: /etc/ssl/cert.pem
  CApath: none
* TLSv1.2 (OUT), TLS handshake, Client hello (1):
* TLSv1.2 (IN), TLS handshake, Server hello (2):
* TLSv1.2 (IN), TLS handshake, Certificate (11):
* TLSv1.2 (IN), TLS handshake, Server key exchange (12):
* TLSv1.2 (IN), TLS handshake, Server finished (14):
* TLSv1.2 (OUT), TLS handshake, Client key exchange (16):
* TLSv1.2 (OUT), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (OUT), TLS handshake, Finished (20):
* TLSv1.2 (IN), TLS change cipher, Change cipher spec (1):
* TLSv1.2 (IN), TLS handshake, Finished (20):
* SSL connection using TLSv1.2 / ECDHE-RSA-CHACHA20-POLY1305
* ALPN, server accepted to use h2
* Server certificate:
*  subject: CN=www.example.com; O=gw-ssl
*  start date: Aug  6 18:27:18 2021 GMT
*  expire date: Aug  4 18:27:18 2031 GMT
*  issuer: CN=www.example.com; O=gateway-root
*  SSL certificate verify result: unable to get local issuer certificate (20), continuing anyway.
* Using HTTP2, server supports multi-use
* Connection state changed (HTTP/2 confirmed)
* Copying HTTP/2 data in stream buffer to connection buffer after upgrade: len=0
* Using Stream ID: 1 (easy handle 0x7fb3c7010a00)
> GET /ratings/1 HTTP/2
> Host: www.example.com:8443
> User-Agent: curl/7.64.1
> Accept: */*
>
* Connection state changed (MAX_CONCURRENT_STREAMS == 2147483647)!
< HTTP/2 200
HTTP/2 200
< content-type: application/json
content-type: application/json
< date: Fri, 06 Aug 2021 21:46:05 GMT
date: Fri, 06 Aug 2021 21:46:05 GMT
< x-envoy-upstream-service-time: 2
x-envoy-upstream-service-time: 2
< server: istio-envoy
server: istio-envoy

<
* Connection #0 to host www.example.com left intact
{"id":1,"ratings":{"Reviewer1":5,"Reviewer2":4}}* Closing connection 0
{{< /tab >}}
{{< /tabs >}}

## HTTPS Redirect

Now that we have traffic being served over HTTPS on our gateway, we may also want to make sure that any http requests that reach us on port 80 get automatically redirected to https on port 443. To accomplish this with the VirtualGateway API, we create a new VirtualGateway, and specify the `httpsRedirect` field to `true` in the connection handler options. Here is an example:

{{< highlight yaml "hl_lines=7-9" >}}
apiVersion: networking.enterprise.mesh.gloo.solo.io/v1beta1
kind: VirtualGateway
metadata:
  name: test-virtual-gateway-2
  namespace: bookinfo
spec:
  connectionHandlers:
  - connectionOptions:
      httpsRedirect: true
  ingressGatewaySelectors:
  - destinationSelectors:
    - kubeServiceMatcher:
        clusters:
        - cluster-1
        labels:
          istio: ingressgateway
        namespaces:
        - istio-system
    portName: http2
{{< /highlight >}}

After applying this gateway, you should now receive a HTTP 301 response (permanently moved), with the new https url being the new path:

{{< highlight shell "hl_lines=10-11" >}}
curl --resolve www.example.com:80:127.0.0.1 http://www.example.com/ratings/1
*   Trying ::1...
* TCP_NODELAY set
* Connected to localhost (::1) port 80 (#0)
> GET /ratings/1 HTTP/1.1
> Host: www.example.com
> User-Agent: curl/7.64.1
> Accept: */*
>
< HTTP/1.1 301 Moved Permanently
< location: https://www.example.com/ratings/1
< date: Fri, 13 Aug 2021 20:12:56 GMT
< server: istio-envoy
< content-length: 0
<
* Connection #0 to host localhost left intact
* Closing connection 0
{{< /highlight >}}

### Summary

In this guide we walked through:
- Generating demo certificates for use in local and test environments.
- Adding those certs to a Kubernetes cluster as a secret
- Configuring a VirtualGateway to use those certificates to serve traffic over HTTPS
- Setting up HTTPS Redirect