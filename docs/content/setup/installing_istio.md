---
title: "Install multicluster Istio"
description: Install Istio in multiple clusters in your Gloo Mesh setup
weight: 40
---

Install Istio for use with Gloo Mesh in a multicluster setting. These instructions are provided for reference only, as your Istio installation process might differ depending on your organization's policies and procedures.

## Before you begin

* Install [`istioctl`](https://istio.io/latest/docs/setup/getting-started/#download), the Istio CLI tool. Gloo Mesh Open Source currently supports Istio versions 1.7 - 1.10. Be sure to download a version of Istio that is supported for the version of your remote clusters. For example, to review Istio support of Kubernetes versions, see the [Istio documentation](https://istio.io/latest/docs/releases/supported-releases/#support-status-of-istio-releases), or to review OpenShift Kubernetes support, see the [OpenShift changelog documentation for the version you want to use](https://docs.openshift.com/container-platform/4.8/release_notes/ocp-4-8-release-notes.html). {{% notice note %}} Istio versions 1.8.0, 1.8.1, and 1.8.2 have a [known issue](https://github.com/istio/istio/issues/28620) where sidecar proxies might not start under specific circumstances. This bug might surface in sidecars configured by Failover Services. This issue is resolved in Istio 1.8.3. {{% /notice %}}
* [Prepare at least two clusters]({{< versioned_link_path fromRoot="/setup/kind_setup/" >}}) for your Gloo Mesh setup, and [install the Gloo Mesh management components]({{< versioned_link_path fromRoot="/setup/community_installation/" >}}) into one cluster.
* Save the contexts for each cluster in environment variables. This guide uses the following variables for the contexts:
  ```shell
  export MGMT_CONTEXT=<management-cluster-context>
  export REMOTE_CONTEXT=<remote-cluster-context>
  ```

## Install Istio version 1.8 or later

In the following `IstioOperator` manifests for a multicluster Istio setup, a NodePort is used to expose the Istio ingress gateway. If you must deploy Istio on different cluster setup, update your gateway settings accordingly. 

1. If you follow the deployment pattern in which the management cluster also runs a service mesh, install Istio in the cluster that runs the Gloo Mesh management components. The Gloo Mesh documentation guides follow this deployment pattern to run the Gloo Mesh management components alongside an Istio service mesh in the same cluster.
   ```shell
   cat << EOF | istioctl manifest install -y --context $MGMT_CONTEXT -f -
   apiVersion: install.istio.io/v1alpha1
   kind: IstioOperator
   metadata:
     name: example-istiooperator
     namespace: istio-system
   spec:
     profile: minimal
     meshConfig:
       enableAutoMtls: true
       defaultConfig:
         proxyMetadata:
           # Enable Istio agent to handle DNS requests for known hosts
           # Unknown hosts are automatically resolved using upstream DNS servers in resolv.conf
           ISTIO_META_DNS_CAPTURE: "true"
     components:
       # Istio Gateway feature
       ingressGateways:
       - name: istio-ingressgateway
         enabled: true
         k8s:
           env:
             - name: ISTIO_META_ROUTER_MODE
               value: "sni-dnat"
           service:
             type: NodePort
             ports:
               - port: 80
                 targetPort: 8080
                 name: http2
               - port: 443
                 targetPort: 8443
                 name: https
               - port: 15443
                 targetPort: 15443
                 name: tls
                 nodePort: 32001
     values:
       global:
         pilotCertProvider: istiod
   EOF
   ```

2. Install Istio in each remote cluster that you want to register with Gloo Mesh.
   ```shell
   cat << EOF | istioctl manifest install -y --context $REMOTE_CONTEXT -f -
   apiVersion: install.istio.io/v1alpha1
   kind: IstioOperator
   metadata:
     name: example-istiooperator
     namespace: istio-system
   spec:
     profile: minimal
     meshConfig:
       enableAutoMtls: true
       defaultConfig:
         proxyMetadata:
           # Enable Istio agent to handle DNS requests for known hosts
           # Unknown hosts are automatically resolved using upstream DNS servers in resolv.conf
           ISTIO_META_DNS_CAPTURE: "true"
     components:
       # Istio Gateway feature
       ingressGateways:
       - name: istio-ingressgateway
         enabled: true
         k8s:
           env:
             - name: ISTIO_META_ROUTER_MODE
               value: "sni-dnat"
           service:
             type: NodePort
             ports:
               - port: 80
                 targetPort: 8080
                 name: http2
               - port: 443
                 targetPort: 8443
                 name: https
               - port: 15443
                 targetPort: 15443
                 name: tls
                 nodePort: 32000
     values:
       global:
         pilotCertProvider: istiod
   EOF
   ```

3. After the installation is complete, verify that the Istio control plane pods are running in each cluster.
   ```shell
   kubectl get pods -n istio-system --context $REMOTE_CONTEXT
   ```

   Example output:
   ```shell
   NAME                                    READY   STATUS    RESTARTS   AGE
   istio-ingressgateway-746d597f7c-g6whv   1/1     Running   0          5d23h
   istiod-7795ccf9dc-vr4cq                 1/1     Running   0          5d22h
   ```

## Install Istio version 1.7

In the following `IstioOperator` manifests for a multicluster Istio setup, a NodePort is used to expose the Istio ingress gateway. If you must deploy Istio on different cluster setup, update your gateway settings accordingly. 

1. If you follow the deployment pattern in which the management cluster also serves as a registered cluster, install Istio in the cluster that also runs the Gloo Mesh management components. The guides throughout this documentation set follow this deployment pattern to run the Gloo Mesh management components alongside an Istio service mesh in the same cluster.
   ```shell
   cat << EOF | istioctl manifest install --context $MGMT_CONTEXT -f -
   apiVersion: install.istio.io/v1alpha1
   kind: IstioOperator
   metadata:
     name: mgmt-plane-istiooperator
     namespace: istio-system
   spec:
     profile: minimal
     addonComponents:
       istiocoredns:
         enabled: true
     components:
       # Istio Gateway feature
       ingressGateways:
       - name: istio-ingressgateway
         enabled: true
         k8s:
           env:
             - name: ISTIO_META_ROUTER_MODE
               value: "sni-dnat"
           service:
             ports:
               - port: 80
                 targetPort: 8080
                 name: http2
               - port: 443
                 targetPort: 8443
                 name: https
               - port: 15443
                 targetPort: 15443
                 name: tls
                 nodePort: 32001
     meshConfig:
       enableAutoMtls: true
     values:
       prometheus:
         enabled: false
       gateways:
         istio-ingressgateway:
           type: NodePort
           ports:
             - targetPort: 15443
               name: tls
               nodePort: 32001
               port: 15443
       global:
         pilotCertProvider: istiod
         controlPlaneSecurityEnabled: true
         podDNSSearchNamespaces:
         - global
   EOF
   ```

2. Install Istio in each remote cluster that you want to register with Gloo Mesh.
   ```shell
   cat << EOF | istioctl manifest install --context $REMOTE_CONTEXT -f -
   apiVersion: install.istio.io/v1alpha1
   kind: IstioOperator
   metadata:
     name: remote-cluster-istiooperator
     namespace: istio-system
   spec:
     profile: minimal
     addonComponents:
       istiocoredns:
         enabled: true
     components:
       # Istio Gateway feature
       ingressGateways:
       - name: istio-ingressgateway
         enabled: true
         k8s:
           env:
             - name: ISTIO_META_ROUTER_MODE
               value: "sni-dnat"
           service:
             ports:
               - port: 80
                 targetPort: 8080
                 name: http2
               - port: 443
                 targetPort: 8443
                 name: https
               - port: 15443
                 targetPort: 15443
                 name: tls
                 nodePort: 32000
     meshConfig:
       enableAutoMtls: true
     values:
       prometheus:
         enabled: false
       gateways:
         istio-ingressgateway:
           type: NodePort
           ports:
             - targetPort: 15443
               name: tls
               nodePort: 32000
               port: 15443
       global:
         pilotCertProvider: istiod
         controlPlaneSecurityEnabled: true
         podDNSSearchNamespaces:
         - global
   EOF
   ```

3. After the installation is complete, verify that the Istio control plane pods are running in each cluster.
   ```shell
   kubectl get pods -n istio-system --context $REMOTE_CONTEXT
   ```

   Example output:
   ```shell
   NAME                                    READY   STATUS    RESTARTS   AGE
   istio-ingressgateway-746d597f7c-g6whv   1/1     Running   0          5d23h
   istiod-7795ccf9dc-vr4cq                 1/1     Running   0          5d22h
   ```

4. Modify `coredns` to enable Istio DNS for the `.global` stub domain for multicluster communication across the management and remote clusters.
  * Management cluster:
     ```shell
     ISTIO_COREDNS=$(kubectl --context $MGMT_CONTEXT -n istio-system get svc istiocoredns -o jsonpath={.spec.clusterIP})
     kubectl --context $MGMT_CONTEXT apply -f - <<EOF
     apiVersion: v1
     kind: ConfigMap
     metadata:
       name: coredns
       namespace: kube-system
     data:
       Corefile: |
         .:53 {
             errors
             health
             ready
             kubernetes cluster.local in-addr.arpa ip6.arpa {
                pods insecure
                upstream
                fallthrough in-addr.arpa ip6.arpa
             }
             prometheus :9153
             forward . /etc/resolv.conf
             cache 30
             loop
             reload
             loadbalance
         }
         global:53 {
             errors
             cache 30
             forward . ${ISTIO_COREDNS}:53
         }
     EOF
     ```
  * Remote cluster:
     ```shell
     ISTIO_COREDNS=$(kubectl --context $REMOTE_CONTEXT -n istio-system get svc istiocoredns -o jsonpath={.spec.clusterIP})
     kubectl --context $REMOTE_CONTEXT apply -f - <<EOF
     apiVersion: v1
     kind: ConfigMap
     metadata:
       name: coredns
       namespace: kube-system
     data:
       Corefile: |
         .:53 {
             errors
             health
             ready
             kubernetes cluster.local in-addr.arpa ip6.arpa {
                pods insecure
                upstream
                fallthrough in-addr.arpa ip6.arpa
             }
             prometheus :9153
             forward . /etc/resolv.conf
             cache 30
             loop
             reload
             loadbalance
         }
         global:53 {
             errors
             cache 30
             forward . ${ISTIO_COREDNS}:53
         }
     EOF
     ```

## Next steps

Now that Istio service meshes are installed, you can [register the clusters]({{< versioned_link_path fromRoot="/setup/community_cluster_registration/" >}}) so that Gloo Mesh can identify and manage their service meshes.
