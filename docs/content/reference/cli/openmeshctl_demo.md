---
title: "openmeshctl demo"
weight: 5
---
## openmeshctl demo

Bootstrap environments for various demos demonstrating Gloo Mesh functionality.

### Options

```
  -h, --help   help for demo
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
* [openmeshctl demo create](../openmeshctl_demo_create)	 - Bootstrap a multicluster Istio demo with Gloo Mesh
* [openmeshctl demo destroy](../openmeshctl_demo_destroy)	 - Clean up bootstrapped local resources. This will delete the kind clusters created by the "openmeshctl demo create"

