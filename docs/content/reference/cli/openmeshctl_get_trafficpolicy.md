---
title: "openmeshctl get trafficpolicy"
weight: 5
---
## openmeshctl get trafficpolicy

Get information about managed trafficpolicies

```
openmeshctl get trafficpolicy [NAME] [flags]
```

### Options

```
  -A, --all-namespaces   List the requsted resource across all namespaces.
  -h, --help             help for trafficpolicy
  -o, --output string    Resource output format. One of: |wide|json|yaml
```

### Options inherited from parent commands

```
      --context string      Name of the kubeconfig context to use for the management cluster
      --kubeconfig string   Path to the kubeconfig from which the management cluster will be accessed
  -n, --namespace string    Namespace that the management plan is installed in on the management cluster (default "gloo-mesh")
  -v, --verbose             Show more detailed output information.
```

### SEE ALSO

* [openmeshctl get](../openmeshctl_get)	 - Display one or many resources

