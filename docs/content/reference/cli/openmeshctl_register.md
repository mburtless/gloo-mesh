---
title: "openmeshctl register"
weight: 5
---
## openmeshctl register

Register a Gloo Mesh data plane cluster

### Synopsis

Registering a cluster installs the cert agent to the remote cluster and
configures the management plan to discover resources on the remote cluster,
which can then be configured using the Gloo Mesh API.

The context used is by default the same name as the cluster but can be changed
via the additional REMOTE CONTEXT argument.

The KubernetesCluster resource may not be installed properly if the --context flag is not included.

```
openmeshctl register NAME [REMOTE CONTEXT] [flags]
```

### Options

```
      --agent-chart string                  Name of the cert agent chart to install.
                                            Can be a URI to a local or remote chart or the name of a chart known to Helm.
                                            Defaults to the URI to the official Gloo Mesh Helm repository.
      --agent-crds-chart string             Name of the cert agent CRDs chart to install.
                                            Can be a URI to a local or remote chart or the name of a chart known to Helm.
                                            Defaults to the URI to the official Gloo Mesh Helm repository.
      --agent-crds-release-name string      Helm release name for the cert agent CRDs chart. Defaults to 'agent-crds' (default "agent-crds")
      --agent-crds-set stringArray          Set values on the command line for the agent CRDs chart (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --agent-crds-set-file stringArray     Set values from respective files specified via the command line for the agent CRDs chart (can specify multiple or separate values with commas: key1=path1,key2=path2)
      --agent-crds-set-string stringArray   Set STRING values on the command line for the agent CRDs chart (can specify multiple or separate values with commas: key1=val1,key2=val2)
      --agent-crds-values strings           Specify values in a YAML file or a URL (can specify multiple) for the agent CRDs chart
      --agent-release-name string           Helm release name for the cert agent chart. Defaults to 'cert-agent' (default "cert-agent")
      --agent-version string                Specific version of the cert agent chart to install.
                                            Defaults to the installed CLI version
      --api-server-address string           Swap out the address of the remote cluster's k8s API server for the value of this flag.
                                            Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that specified in the local kubeconfig.
      --cluster-domain string               The cluster domain used by the Kubernetes DNS Service in the registered cluster.
                                            Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/ (default "cluster.local")
  -h, --help                                help for register
      --remote-kubeconfig string            Path to the kubeconfig from which the remote cluster well be accessed if different from the management cluster.
      --remote-namespace string             Namespace on the remote cluster that the agent will be installed to. (default "gloo-mesh")
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

