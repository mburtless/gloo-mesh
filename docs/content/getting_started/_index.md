---
title: "Getting Started"
menuTitle: Getting Started
description: How to get started using Gloo Mesh
weight: 10
---

Welcome to Gloo Mesh, the open-source, multi-cluster, multi-mesh management plane. Gloo Mesh simplifies service-mesh operations and lets you manage multiple clusters of a service mesh from a centralized management plane. Gloo Mesh takes care of things like shared-trust/root CA federation, workload discovery, unified multi-cluster/global traffic policy, access policy, and more. 

## Getting `meshctl`

Install the latest version of `meshctl`. Make sure to add `meshctl` to your PATH (see [Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/), [macOS](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/), or [Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/) for specific instructions). For more information, see [`meshctl` CLI]({{< versioned_link_path fromRoot="/setup/installation/meshctl_cli_install/" >}}).

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

## Deploying Gloo Mesh

In this section, we detail a few ways to get you up and running with Gloo Mesh and Gloo Mesh Enterprise. For detailed
information on each aspect of Gloo Mesh installation, check out the [setup guides]({{% versioned_link_path fromRoot="/setup/" %}}).

{{% children description="true" %}}
