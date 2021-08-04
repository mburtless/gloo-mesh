---
title: "meshctl check agent"
weight: 5
---
## meshctl check agent

Perform post-install checks on a Gloo Mesh agent

```
meshctl check agent [flags]
```

### Options

```
  -h, --help                 help for agent
      --kubeconfig string    Path to the kubeconfig from which the remote cluster will be accessed.
      --kubecontext string   Name of the kubeconfig context to use for the remote cluster.
      --namespace string     Namespace that Gloo Mesh is installed in. (default "gloo-mesh")
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl check](../meshctl_check)	 - Perform health checks on the Gloo Mesh system

