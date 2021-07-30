---
title: Manual Istio Root CA
menuTitle: Generating Istio Root CAs
description: Manually generating CA certificates for each Istio deployment.
weight: 40
---

Istio provides a tool to quickly generate CA certificates for Istio deployments.

* [Istio Cert Tools](https://github.com/istio/istio/blob/master/tools/certs/)

## Create Istio CA Certificates

1. Download Istio

```sh
ISTIO_VERSION=1.10.3
curl -L https://istio.io/downloadIstio | ISTIO_VERSION=$ISTIO_VERSION sh -
```

2. Set variables: below is the variables you can override when generating your certificates

```sh
#------------------------------------------------------------------------
# variables: root CA
ROOTCA_DAYS ?= 3650
ROOTCA_KEYSZ ?= 4096
ROOTCA_ORG ?= Istio
ROOTCA_CN ?= Root CA
KUBECONFIG ?= $(HOME)/.kube/config
ISTIO_NAMESPACE ?= istio-system
# Additional variables are defined in root-ca.conf target below.

#------------------------------------------------------------------------
# variables: intermediate CA
INTERMEDIATE_DAYS ?= 730
INTERMEDIATE_KEYSZ ?= 4096
INTERMEDIATE_ORG ?= Istio
INTERMEDIATE_CN ?= Intermediate CA
INTERMEDIATE_SAN_DNS ?= istiod.istio-system.svc

```

3. Generate the Root certificate

```sh
# Go into certs directory
cd istio-$ISTIO_VERSION/tools/certs

# Create root certificate
make -f Makefile.selfsigned.mk \
  ROOTCA_CN="Solo Root CA" \
  ROOTCA_ORG=solo.io \
  root-ca


# Delete certs and start over if needed
# make -f Makefile.selfsigned.mk clean

```

4. Generate Intermedicate CA Certificates for each cluster

```sh
CLUSTER_NAME=cluster1

make -f Makefile.selfsigned.mk \
  INTERMEDIATE_CN="Solo Intermediate CA" \
  INTERMEDIATE_ORG=solo.io \
  $CLUSTER_NAME-cacerts

# apply kubernetes secret to cluster
kubectl create secret generic cacerts -n istio-system \
      --from-file=$CLUSTER_NAME/ca-cert.pem \
      --from-file=$CLUSTER_NAME/ca-key.pem \
      --from-file=$CLUSTER_NAME/root-cert.pem \
      --from-file=$CLUSTER_NAME/cert-chain.pem
```

## Example Configurations

Below is an example configurations for generating Istio CA certificates for each Istio Clusters with the same root. This is for users who want to use the below configuration as a starting point for their own pki infrastructure.

```sh
cat > "root-ca.conf" <<EOF
[ req ]
encrypt_key = no
prompt = no
utf8 = yes
default_md = sha256
default_bits = 4096
req_extensions = req_ext
x509_extensions = req_ext
distinguished_name = req_dn
[ req_ext ]
subjectKeyIdentifier = hash
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, nonRepudiation, keyEncipherment, keyCertSign
[ req_dn ]
O = Istio
CN = Root CA
EOF

cat > "cluster-ca.conf" <<EOF
[ req ]
encrypt_key = no
prompt = no
utf8 = yes
default_md = sha256
default_bits = 4096
req_extensions = req_ext
x509_extensions = req_ext
distinguished_name = req_dn
[ req_ext ]
subjectKeyIdentifier = hash
basicConstraints = critical, CA:true, pathlen:0
keyUsage = critical, digitalSignature, nonRepudiation, keyEncipherment, keyCertSign
subjectAltName=@san
[ san ]
DNS.1 = istiod.istio-system.svc
[ req_dn ]
O = Istio
CN = Intermediate CA
L = ${cluster}
EOF

```
