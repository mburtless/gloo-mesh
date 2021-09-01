---
title: "Enterprise"
menuTitle: Enterprise
description: Install Gloo Mesh Enterprise
weight: 100
---

{{% notice note %}}
Gloo Mesh Enterprise is the paid version of Gloo Mesh including the Gloo Mesh UI and multi-cluster role-based access control. To complete the installation you will need a license key. You can get a trial license by [requesting a demo](https://www.solo.io/products/gloo-mesh/) from the website.
{{% /notice %}}

In a typical deployment, Gloo Mesh Enterprise uses a single Kubernetes cluster to host the management plane while each service mesh can run on its own independent cluster.
This document describes how to install the Gloo Mesh Enterprise management plane components with both `meshctl` and Helm. 
We recommend using Helm to install Gloo Mesh in a production environment as it offers rich customization.

A conceptual overview of the Gloo Mesh Enterprise architecture can be found [here]({{% versioned_link_path fromRoot="/concepts/relay" %}}).

## Installing with `meshctl`

Install the latest version of `meshctl`. Make sure to add `meshctl` to your PATH (see [Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/), [macOS](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/), or [Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/) for specific instructions). For more information, see [`meshctl` CLI]({{< versioned_link_path fromRoot="/setup/installation/meshctl_cli_install/" >}}).

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```
```

Installing Gloo Mesh Enterprise with `meshctl` is a simple process. You will use the command `meshctl install enterprise` and supply the license key, as well as any chart values you want to update, and arguments pointing to the cluster where Gloo Mesh Enterprise will be installed. For our example, we are going to install Gloo Mesh Enterprise on the cluster currently in context. First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

We are not going to change any of the default values in the underlying chart, so the only argument needed is the license key. However `meshctl install enterprise` is backed by the Gloo Mesh Enterprise Helm chart, so you can customize your installation with a Helm values override via the flag `--chart-values-file`. Review the Gloo Mesh Enterprise Helm values documentation [here]({{% versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise/" %}}).

```shell
meshctl install enterprise --license $GLOO_MESH_LICENSE_KEY
```

You should see the following output from the command:

```shell
Installing Helm chart
Finished installing chart 'gloo-mesh-enterprise' as release gloo-mesh:gloo-mesh
```

The installer has created the namespace `gloo-mesh` and installed Gloo Mesh Enterprise into the namespace using a Helm chart with default values.

{{% notice note %}}`meshctl` will create a self-signed certificate authority for mTLS if you do not supply your own certificates.{{% /notice %}}

To undo the installation, you can simply run the `uninstall` command:

```shell
meshctl uninstall
```

## Helm

You may prefer to use the Helm chart directly rather than using the `meshctl` CLI tool. This section will take you through the steps necessary to deploy a Gloo Mesh Enterprise installation from the Helm chart.

1. Add the Helm repo

```shell
helm repo add gloo-mesh-enterprise https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise
```

2. (optional) View available versions

```shell
helm search repo gloo-mesh-enterprise
```

3. (optional) View Helm values

```shell
helm show values gloo-mesh-enterprise/gloo-mesh-enterprise
```

Note that the `gloo-mesh-enterprise` Helm chart bundles multiple components, including `enterprise-networking`, `rbac-webhook`, and `gloo-mesh-ui`. Each is versioned in step with the parent `gloo-mesh-enterprise` chart, and each has its own Helm values for advanced customization. Review the Gloo Mesh Enterprise Helm values documentation [here]({{% versioned_link_path fromRoot="/reference/helm/gloo_mesh_enterprise/" %}}).

4. Install

{{% notice note %}}If you are running Gloo Mesh Enterprise's management plane on a cluster you intend to register (i.e. also run a service mesh), set the `enterprise-networking.cluster` value to the cluster name you intend to set for the management cluster at registration time. {{% /notice %}}

```shell
kubectl create ns gloo-mesh

helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY}
```

{{% notice note %}}The Helm value `selfSigned` is set to `true` by default. This means the Helm chart will create certificates for you if you do not supply them through values.{{% /notice %}}

## Verify install
Once you've installed Gloo Mesh, verify what components were installed:

```shell
kubectl get pods -n gloo-mesh
```

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
```

Running the check command from meshctl will also verify everything was installed correctly:

```shell
meshctl check server
```

```shell
Gloo Mesh Management Cluster Installation
--------------------------------------------

🟢 Gloo Mesh Pods Status

🟡 Gloo Mesh Agents Connectivity
    Hints:
    * No registered clusters detected. To register a remote cluster that has a deployed Gloo Mesh agent, add a KubernetesCluster CR.
      For more info, see: https://docs.solo.io/gloo-mesh/latest/setup/cluster_registration/enterprise_cluster_registration/

Management Configuration
---------------------------

🟢 Gloo Mesh CRD Versions

🟢 Gloo Mesh Networking Configuration Resources
```

These checks include the following:

1. That the Gloo Mesh pods are up and running.

2. That expected remote agents are connected (Because this is a first install with no registered clusters,
   the check will warn that no agents are connected. We will register a cluster in the next guide.)
   
3. That the CRD versions expected by `enterprise-networking` matches the versions of the CRDs installed on the cluster.

## Advanced install options

The following subsections describe advanced installation options for Gloo Mesh Enterprise. They are not required.

#### Manual Certificate Creation

Also see [Certificate Management](../../guides/certificate_management/_index.md) for more information

Gloo Mesh Enterprise's default behavior is to create self-signed certificates at install time to handle
bootstrapping mTLS connectivity between the [server and agent components]({{% versioned_link_path fromRoot="/concepts/relay" %}})
of Gloo Mesh Enterprise. If you would prefer to provision your own secrets for this purpose, read on. Note that this is
not required for standard Gloo Mesh Enterprise deployments.

The following script illustrates how you can create these certificates manually. If you do not create certificates yourself, both the Helm chart and `meshctl` will create self-signed certificates for you.

At a high level, the script achieves the following:

1. Creates a root certificate (`RELAY_ROOT_CERT_NAME`). This is distributed to managed clusters
   so that relay agents can use it verify the identity of the relay server.

2. Create a certificate for the relay server (`RELAY_SERVER_CERT_NAME`), derived from the root certificate,
   which is presented by the server to relay agents.

3. Create a signing certificate (`RELAY_SIGNING_CERT_NAME`), derived from the root certificate,
   which is used by the relay server to issue certificates for relay agents once initial trust has been established.

{{% notice note %}}
The following was tested with OpenSSL 1.1.1h  (released on 22 Sep 2020)

Mac users may have LibreSSL installed by default. If so, we recommend that you `brew install openssl` and verify that
`openssl version` returns an OpenSSL version before you proceed.
{{% /notice %}}

```bash
RELAY_ROOT_CERT_NAME=relay-root
RELAY_SERVER_CERT_NAME=relay-server-tls
RELAY_SIGNING_CERT_NAME=relay-tls-signing

echo "creating root cert ..."

openssl req -new -newkey rsa:4096 -x509 -sha256 \
        -days 3650 -nodes -out ${RELAY_ROOT_CERT_NAME}.crt -keyout ${RELAY_ROOT_CERT_NAME}.key \
        -subj "/CN=enterprise-networking-ca" \
        -addext "extendedKeyUsage = clientAuth, serverAuth"


echo "creating grpc server tls cert ..."

# server cert
cat > "${RELAY_SERVER_CERT_NAME}.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS = *.gloo-mesh
EOF

openssl genrsa -out "${RELAY_SERVER_CERT_NAME}.key" 2048
openssl req -new -key "${RELAY_SERVER_CERT_NAME}.key" -out ${RELAY_SERVER_CERT_NAME}.csr -subj "/CN=enterprise-networking-ca" -config "${RELAY_SERVER_CERT_NAME}.conf"
openssl x509 -req \
  -days 3650 \
  -CA ${RELAY_ROOT_CERT_NAME}.crt -CAkey ${RELAY_ROOT_CERT_NAME}.key \
  -set_serial 0 \
  -in ${RELAY_SERVER_CERT_NAME}.csr -out ${RELAY_SERVER_CERT_NAME}.crt \
  -extensions v3_req -extfile "${RELAY_SERVER_CERT_NAME}.conf"

echo "creating identity server signing cert ..."

# signing cert
cat > "${RELAY_SIGNING_CERT_NAME}.conf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = critical,CA:TRUE
keyUsage = digitalSignature, keyEncipherment, keyCertSign
extendedKeyUsage = clientAuth, serverAuth
subjectAltName = @alt_names
[alt_names]
DNS = *.gloo-mesh
EOF

openssl genrsa -out "${RELAY_SIGNING_CERT_NAME}.key" 2048
openssl req -new -key "${RELAY_SIGNING_CERT_NAME}.key" -out ${RELAY_SIGNING_CERT_NAME}.csr -subj "/CN=enterprise-networking-ca" -config "${RELAY_SIGNING_CERT_NAME}.conf"
openssl x509 -req \
  -days 3650 \
  -CA ${RELAY_ROOT_CERT_NAME}.crt -CAkey ${RELAY_ROOT_CERT_NAME}.key \
  -set_serial 0 \
  -in ${RELAY_SIGNING_CERT_NAME}.csr -out ${RELAY_SIGNING_CERT_NAME}.crt \
  -extensions v3_req -extfile "${RELAY_SIGNING_CERT_NAME}.conf"
```

Next, we'll create secrets on the management and remote clusters with all the requisite generated certificates. Let's
define the management cluster as the cluster where we'll install the Gloo Mesh control plane, and the remote cluster as
a cluster running a service mesh we would like Gloo Mesh to discover and configure. On the management cluster, we need
the root cert, server cert, and signing cert. On the remote clusters, we just need the root cert.

```bash
MGMT_CLUSTER=cluster-1
MGMT_CONTEXT=kind-cluster-1

REMOTE_CLUSTER=cluster-2
REMOTE_CONTEXT=kind-cluster-2

# ensure gloo-mesh namespace exists on both clusters
for context in ${MGMT_CONTEXT} ${REMOTE_CONTEXT}; do
  kubectl --context ${context} create namespace gloo-mesh
done

# create secrets from certs

# Note: ${RELAY_SERVER_CERT_NAME}-secret must match the server Helm value `relayTlsSecret.Name`
kubectl create secret generic ${RELAY_SERVER_CERT_NAME}-secret \
  --from-file=tls.key=${RELAY_SERVER_CERT_NAME}.key \
  --from-file=tls.crt=${RELAY_SERVER_CERT_NAME}.crt \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
   --context ${MGMT_CONTEXT} \
  --namespace gloo-mesh

# Note: ${RELAY_ROOT_CERT_NAME}-tls-secret must match the agent Helm value `relay.rootTlsSecret.Name`
for context in ${MGMT_CONTEXT} ${REMOTE_CONTEXT}; do
echo "creating matching root cert for agent in cluster context ${context}..."
  kubectl create secret generic ${RELAY_ROOT_CERT_NAME}-tls-secret \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
   --context ${context} \
  --namespace gloo-mesh
done

# Note: ${RELAY_SIGNING_CERT_NAME}-secret must match the server Helm value `signingTlsSecret.Name`
kubectl create secret generic ${RELAY_SIGNING_CERT_NAME}-secret \
  --from-file=tls.key=${RELAY_SIGNING_CERT_NAME}.key \
  --from-file=tls.crt=${RELAY_SIGNING_CERT_NAME}.crt \
  --from-file=ca.crt=${RELAY_ROOT_CERT_NAME}.crt \
  --dry-run=client -oyaml | kubectl apply -f- \
  --context ${MGMT_CONTEXT} \
  --namespace gloo-mesh
```

You can also choose to use your own internal PKI to create and assign certificates to Gloo Mesh Enterprise. At a minimum, you would need the certificate chain to your PKI root CA, a server certificate for relay server communication, and a signing certificate the relay server can use to generate relay agent certificates. Please refer to your PKI documentation for more information.

## Getting started with RateLimiting and External Authentication

{{% notice note %}}
Rate limiting and Extauth were introduced with **Gloo Mesh Enterprise**, release `v1.1.0`.
If you are using an earlier version, these features will not be available.
{{% /notice %}}

To get started using rate limiting and external authentication, you can enable these features in the `gloo-mesh` namespace. This
is **not** the recommended setup for using ratelimit and external authentication in a secure production environment, but
is a quick way to get started testing the features. Follow the [recommended setup of rate limiting and external authentication]({{% versioned_link_path fromRoot="/setup/installation/recommended_setup/" %}})
for the best practice setup environment.

For demo setup, we will use a single namespace, `gloo-mesh`, for our management plane (the gloo mesh server and agents) and
the data plane components (Ratelimit and ExtAuth). 
This is not recommended for non-demo purposes as the injection directive on the namespace makes the 
management plane components dependent on the functionality of Istio’s mutating webhook and is not recommended as best practice.

There are two steps to install Gloo Mesh with Ratelimit and Extauth enabled in the same namespace:

1. Install the Gloo Mesh Enterprise chart with the RateLimit and/or ExtAuth enabled to `gloo-mesh` namespace.
2. Label the `gloo-mesh` namespace for istio injection.

RateLimit and/or ExtAuth are disabled by default on installation. They can be enabled via helm as follows in the `enterprise-agent`:

```yaml
rate-limiter: 
  enabled: true
ext-auth-service: 
  enabled: true
```

After adding the Gloo Mesh helm chart as described above, create the `gloo-mesh` namespace and install Gloo Mesh Enterprise with:

```shell
helm install gloo-mesh-enterprise gloo-mesh-enterprise/gloo-mesh-enterprise --create-namespace --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY}
```

Once you've installed Gloo Mesh, verify what components were installed:

```shell
kubectl get pods -n gloo-mesh
```

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
```

For the simple demo setup, we will use `cluster-1` as the management and as a worker cluster and `cluster-2` just 
as a worker cluster. It is best practice to separate the management cluster and the worker clusters as described in
the [recommended setup guide]{{% versioned_link_path fromRoot="/setup/installation/recommended_setup/" %}}).

To install the `enterprise-agent` on both `cluster-1` and `cluster-2`, first create the `gloo-mesh` namespace on 
each cluster. Then install the `enterprise-agent` helm chart with Extauth and/or Ratelimit enabled:

```shell
kubectl create ns gloo-mesh

helm install enterprise-agent enterprise-agent/enterprise-agent --namespace gloo-mesh \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY} --set rate-limiter.enabled=true --set ext-auth-service.enabled=true
```

To label the `gloo-mesh` namespace for istio injection, run the following on both `cluster-1` and `cluster-2`:

```shell
kubectl --context cluster-1 label ns gloo-mesh istio-injection=enabled --overwrite
kubectl --context cluster-2 label ns gloo-mesh istio-injection=enabled --overwrite
```

Remember you will need to label the `gloo-mesh` namespace for all clusters. Check that the injection label
has been applied with:

```shell
kubectl get pods -n gloo-mesh
```

The output should contain the Ratelimit and Extauth components successfully installed and injected:

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
dashboard-6d6b944cdb-jcpvl               3/3     Running   0          4m2s
enterprise-networking-84fc9fd6f5-rrbnq   1/1     Running   0          4m2s
enterprise-agent-ccfcddd1f5-frcnf        1/1     Running   0          4m2s
rate-limit-3d62244cdb-fcrvd              2/2     Running   0          4m2s
ext-auth-service-3d62244cdb-fcrvd        2/2     Running   0          4m2s
```

Great! Ratelimit and ExtAuth are up and running. Check out the guides for setting up [ratelimiting]({{% versioned_link_path fromRoot="/guides/gateway/ratelimiting/" %}}) 
and [external authentication]({{% versioned_link_path fromRoot="/guides/gateway/auth/" %}})
to use these features. For the best practice setup of Ratelimit and Extauth in a separate namespace, 
follow the [recommended installation guide]({{% versioned_link_path fromRoot="/setup/installation/recommended_setup/" %}}).

## Next Steps

Now that we have Gloo Mesh Enterprise installed, let's [register a cluster]({{< versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration" >}}).
