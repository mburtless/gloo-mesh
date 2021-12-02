---
title: "openmeshctl uninstall"
weight: 5
---
## openmeshctl uninstall

Uninstall Gloo Mesh

### Synopsis

Uninstall the Gloo Mesh management plan from a Kubernetes cluster.

```
openmeshctl uninstall [flags]
```

### Options

```
  -h, --help                  help for uninstall
      --release-name string   Helm release name for the Gloo Mesh chart. Defaults to 'gloo-mesh' (default "gloo-mesh")
```

### Options inherited from parent commands

```
      --context string      Name of the kubeconfig context to use for the management cluster
      --kubeconfig string   Path to the kubeconfig from which the management cluster will be accessed
  -n, --namespace string    Namespace that the management plan is installed in on the management cluster (default "gloo-mesh")
  -v, --verbose             Show more detailed output information.
```

### SEE ALSO

* [openmeshctl](../openmeshctl)	 - The Command Line Interface for managing Gloo Mesh.

