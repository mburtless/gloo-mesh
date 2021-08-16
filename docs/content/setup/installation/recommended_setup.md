---
title: Best Practice Installation
menuTitle: Best Practice Installation
description: Best practice guide for running Gloo Mesh Enterprise in production
weight: 20
---

{{% notice note %}} Basic Gateway features are available to anyone with a Gloo Mesh Enterprise license. Some advanced features require a Gloo Mesh Gateway license. {{% /notice %}}

### Dashboard Authentication

We recommend securing the Gloo Mesh Enterprise Dashboard by requiring authentication with an OpenID Connect identity provider.
Users accessing the dashboard will be required to authenticate with the OIDC provider and all requests to retrieve data from the API will also be authenticated.

### Manage Certs

We recommend you do not utilize Gloo Mesh to issue certificates or manage Istio CA certificates in production
and instead add automation so that the certificates can be rotated easily in the future as described in the [certificate management guide]({{% versioned_link_path fromRoot="/guides/certificate_management/without_gloo_mesh/" %}}).

When certificates are issued, Istio-controlled pods need to be bounced (restarted) to ensure they pick up the new certificates.
The certificate issuer will create a [PodBounceDirective]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.gloo-mesh.api.certificates.v1.pod_bounce_directive/#certificates.mesh.gloo.solo.io.PodBounceDirectiveSpec" >}}) 
containing the namespaces and labels of the pods that need to be bounced in order to pick up the new certs. We recommend the PodBounceDirective feature is turned off by setting the `autoRestartPods` field to `false` in the VirtualMesh as shown in this example:

```shell
apiVersion: networking.mesh.gloo.solo.io/v1
kind: VirtualMesh
metadata:
  name: virtual-mesh
  namespace: gloo-mesh
spec:
  mtlsConfig:
    autoRestartPods: false # disable autoRestartPods in production
    shared:
      rootCertificateAuthority:
        generated: {}
  federation:
    # federate all Destinations to all external meshes
    selectors:
    - {}
  meshes:
  - name: istiod-istio-system-cluster1
    namespace: gloo-mesh
```


## Setting Up Ratelimiting and External Authentication

In the [getting started guide]({{< versioned_link_path fromRoot="/setup/installation/enterprise_installation" >}}) 
we provided a simple setup to demonstrate a simple Gloo Mesh setup with Ratelimit and Extauth features enabled. 
This document describes the best practice method for installing the Gloo Mesh Enterprise management plane components and data plane components via Helm.
Helm is the recommended method of installing Gloo Mesh to your production environment as it offers rich customization.

In a typical deployment, Gloo Mesh Enterprise uses a single Kubernetes cluster to host the management plane. 
To enable mTLS with Ratelimiting and Extauth, we need to add an injection directive for those components. 
An injection directive on the `gloo-mesh` namespace itself makes the management plane components dependent on the functionality of 
Istio’s mutating webhook, which may be a fragile coupling and is not recommended as best practice. 

Instead, we want to separate the management plane and data plane components into two separate namespaces and **label only our data plane components** with the injection directive. 
This requires three separate steps: 

- Install the Gloo Mesh Enterprise chart with just enterprise-agent enabled to `gloo-mesh` namespace.
- Install the Gloo Mesh Enterprise chart with just RateLimit and ExtAuth enabled to the `gloo-mesh-addons` namespace.
- Label the `gloo-mesh-addons` namespace for istio injection.

## Setup

First, let's set a variable for the license key.

```shell
GLOO_MESH_LICENSE_KEY=<your_key_here> # You'll need to supply your own key
```

For this guide we will use `mgmt-cluster` as the management cluster where Gloo Mesh Enterprise is installed and no agents. We will have two 
worker clusters, `cluster-1` and `cluster-2` which will be registered with the managment plane.

## Install Gloo Mesh Enterprise via Helm

First install the Gloo Mesh Gateway chart with only the enterprise-agent enabled to `gloo-mesh` namespace:

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

4. Install

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

5. Register `cluster-1` and `cluster-2` using the `gloo-mesh` namespace. 

Follow the [enterprise cluster registration guide]({{% versioned_link_path fromRoot="/setup/cluster_registration/enterprise_cluster_registration/" %}}) to create the enterprise-agent. 
You will only need to do this for the `gloo-mesh` namespace.

Now you should have the Gloo Mesh Enterprise management plane installed along with the enterprise-agent setup.

## Install Extauth and RateLimit

The RateLimit and Extauth are part of the Gloo Mesh Enterprise `enterprise-agent` helm chart. This chart is installed to the 
`gloo-mesh` namespace when you register a cluster. We will now install just the Extauth and Ratelimit to the `gloo-mesh-addons` namespace with just the data plane components enabled.

RateLimit and ExtAuth are disabled by default on installation. They can be enabled via helm as follows:

```yaml
rate-limiter: 
  enabled: true
ext-auth-service: 
  enabled: true
```

The Enterprise Agent is enabled by default. The Enterprise Agent should be disabled via helm with:

```yaml
enterpriseAgent:
  enabled: false
```

Create a new namespace and install the RateLimit and Extauth, without the `enterprise-agent` using helm on the worker clusters `cluster-1` and `cluster-2`:

```shell
helm install enterprise-agent enterprise-agent/enterprise-agent --create-namespace --namespace gloo-mesh-addons \
  --set licenseKey=${GLOO_MESH_LICENSE_KEY} --set rate-limiter.enabled=true --set ext-auth-service.enabled=true --set enterpriseAgent.enabled=false
```

Now you should have the Ratelimit and Extauth data plane components installed in the `gloo-mesh-addons` namespace.

Finally, we need to add the label to enable istio-injection for the data plane components. 
To label the `gloo-mesh-addons` namespace for istio injection, run the following on both worker clusters, `cluster-2` and `cluster-3`:

```shell
kubectl --context cluster-2 label ns gloo-mesh-addons istio-injection=enabled --overwrite
kubectl --context cluster-3 label ns gloo-mesh-addons istio-injection=enabled --overwrite
```

Remember you will need to label the `gloo-mesh-addons` namespace for all clusters with ExtAuth or RateLimiter deployments . Check that the injection label
has been applied:

```shell
kubectl get pods -n gloo-mesh-addons
```

The output should contain the Ratelimit and Extauth components successfully installed and injected:

```shell
NAME                                     READY   STATUS    RESTARTS   AGE
rate-limit-3d62244cdb-fcrvd              2/2     Running   0          4m2s
ext-auth-service-3d62244cdb-fcrvd        2/2     Running   0          4m2s
```

## Next Steps

Great! Ratelimit and ExtAuth are up and running. Check out the guides for [ratelimiting]({{% versioned_link_path fromRoot="/guides/gateway/ratelimiting/" %}}) and [external authentication]({{% versioned_link_path fromRoot="/guides/gateway/auth/" %}})
to use these features. 

If you have any questions about running Gloo Mesh in production or need help setting up Gloo Mesh join us on our [Slack channel](https://slack.solo.io/).