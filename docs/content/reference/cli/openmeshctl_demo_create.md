---
title: "openmeshctl demo create"
weight: 5
---
## openmeshctl demo create

Bootstrap a multicluster Istio demo with Gloo Mesh

### Synopsis


Bootstrap a multicluster Istio demo with Gloo Mesh.

Running the Gloo Mesh demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, istioctl, and docker.
We recommend allocating at least 8GB of RAM for Docker.

This command will bootstrap 2 clusters, one of which will run the Gloo Mesh
management-plane as well as Istio, and the other will just run Istio.


```
openmeshctl demo create [flags]
```

### Options

```
      --chart string      Gloo Mesh helm chart to install on the management plane
  -h, --help              help for create
      --skip-gm-install   If set to true, the local kind clusters, Istio installation, and bookinfo applications are all installed - but Gloo Mesh is NOT installed. Useful for simluating an example environment to use for trying out manual installation of Gloo Mesh.
      --version string    Gloo Mesh version to install. defaults to meshctl version
```

### Options inherited from parent commands

```
      --context string      Name of the kubeconfig context to use for the management cluster
      --kubeconfig string   Path to the kubeconfig from which the management cluster will be accessed
  -n, --namespace string    Namespace that the management plan is installed in on the management cluster (default "gloo-mesh")
  -v, --verbose             Show more detailed output information.
```

### SEE ALSO

* [openmeshctl demo](../openmeshctl_demo)	 - Bootstrap environments for various demos demonstrating Gloo Mesh functionality.

