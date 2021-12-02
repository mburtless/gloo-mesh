---
title: "openmeshctl deregister"
weight: 5
---
## openmeshctl deregister

Deregister a data plane cluster

### Synopsis


Deregistering a cluster removes the cert agent along with other Gloo Mesh-owned
resources such as service accounts.

The context used is by default the same name as the cluster but can be changed
via the additional REMOTE CONTEXT argument.

The KubernetesCluster resource may not be uninstalled properly if the --context flag is not included.

```
openmeshctl deregister NAME [REMOTE CONTEXT] [flags]
```

### Options

```
      --agent-crds-release-name string   Helm release name for the cert agent CRDs chart. Defaults to 'agent-crds' (default "agent-crds")
      --agent-release-name string        Helm release name for the cert agent chart. Defaults to 'cert-agent' (default "cert-agent")
  -h, --help                             help for deregister
      --remote-kubeconfig string         Path to the kubeconfig from which the remote cluster well be accessed if different from the management cluster.
      --remote-namespace string          Namespace on the remote cluster that the agent will be installed to. (default "gloo-mesh")
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

