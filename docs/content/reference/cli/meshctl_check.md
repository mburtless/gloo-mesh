---
title: "meshctl check"
weight: 5
---
## meshctl check

Perform health checks on the Gloo Mesh system

```
meshctl check [flags]
```

### Options

```
      --config string        set the path to the meshctl config file (default "<home_directory>/.gloo-mesh/meshctl-config.yaml")
  -h, --help                 help for check
      --local-port uint32    local port used to open port-forward to enterprise-networking pod (default 9091)
      --remote-port uint32   remote port used to open port-forward to enterprise-networking pod (default 9091)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl](../meshctl)	 - The Command Line Interface for managing Gloo Mesh.
* [meshctl check agent](../meshctl_check_agent)	 - Perform post-install checks on a Gloo Mesh agent
* [meshctl check server](../meshctl_check_server)	 - Perform post-install checks on the Gloo Mesh management plane

