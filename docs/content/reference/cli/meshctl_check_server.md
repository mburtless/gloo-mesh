---
title: "meshctl check server"
weight: 5
---
## meshctl check server

Perform post-install checks on the Gloo Mesh management plane

```
meshctl check server [flags]
```

### Options

```
  -h, --help                 help for server
      --kubeconfig string    Path to the kubeconfig from which the management cluster will be accessed
      --kubecontext string   Name of the kubeconfig context to use for the management cluster
      --local-port uint32    local port used to open port-forward to enterprise mgmt pod (enterprise only) (default 9091)
      --namespace string     namespace that Gloo Mesh is installed in (default "gloo-mesh")
      --remote-port uint32   remote port used to open port-forward to enterprise mgmt pod (enterprise only). set to 0 to disable checks on the mgmt server (default 9091)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl check](../meshctl_check)	 - Perform health checks on the Gloo Mesh system

