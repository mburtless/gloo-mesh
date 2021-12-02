---
title: "openmeshctl demo destroy"
weight: 5
---
## openmeshctl demo destroy

Clean up bootstrapped local resources. This will delete the kind clusters created by the "openmeshctl demo create"

```
openmeshctl demo destroy [flags]
```

### Options

```
  -h, --help   help for destroy
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

