---
title: "meshctl CLI"
description: Install the meshctl CLI
weight: 10
---

Use the Gloo Mesh command line interface (CLI) tool, `meshctl`, to set up Gloo Mesh, register clusters, describe your Gloo Mesh resources, and more.

## Quick installation

Install the latest version of `meshctl`. Make sure to add `meshctl` to your PATH (see [Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/), [macOS](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/), or [Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/) for specific instructions).

```shell
curl -sL https://run.solo.io/meshctl/install | sh
```

## Install a specific version of the CLI

You can pass a variable on the command line to download a specific version of `meshctl`.

1.  Go the [Gloo Mesh releases GitHub page](https://github.com/solo-io/gloo-mesh/releases) and find the release version that you want to install.
2.  Export the release version as a variable in your command line.
    ```shell
    export GLOO_MESH_VERSION=v1.x.x
    ```
3.  Install the `meshctl` CLI.
    ```shell
    curl -sL https://run.solo.io/meshctl/install | sh -
    ```
4.  Add `meshctl` on your PATH system variable for global access on the command line. The steps vary depending on your operating system. The following eample is for macOS. For more information, see [Windows](https://helpdeskgeek.com/windows-10/add-windows-path-environment-variable/), [macOS](https://osxdaily.com/2014/08/14/add-new-path-to-path-command-line/), or [Linux](https://linuxize.com/post/how-to-add-directory-to-path-in-linux/).
    ```shell
    export PATH=$HOME/.gloo-mesh/bin:$PATH
    echo $PATH
    ```
     
    Example output:
    ```
    /Users/<name>/.gloo-mesh/bin:/usr/local/bin
    ```
5.  Verify that you can run `meshctl` commands.
    ```shell
    meshctl version
    ```

{{% notice tip %}}
On macOS, you might see the following warning: `“meshctl” cannot be opened because it is from an unidentified developer.` In the Finder app, navigate to the `~/.gloo-mesh/binmeshctl` executable file, right-click the file, click **Open**, and confirm that you want toopen the file. For more information, try searching the warning and following a guide such as[this blog](https://www.howtogeek.com/205393gatekeeper-101-why-your-mac-only-allows-apple-approved-software-by-default/).
{{% /notice %}}

Good job! You now have the version of `meshctl` that you want installed. Next, [install Gloo Mesh]({{< versioned_link_path fromRoot="/getting_started/#deploying-gloo-mesh" >}}) in your clusters.

Do you have multiple cluster environments that require different versions of Gloo Mesh, Istio, and Kubernetes? Consider downloading each `meshctl`, `istioctl`, and `kubectl` version binary file to a separate directory. Then, you can set up an alias in your local command line interface profile to point to the binary file directory that matches the version of the cluster environment that you want to work with.

## Upgrade the CLI

To upgrade, re-install the CLI. You can install the [latest]({{<ref "#quick-installation" >}})) or a [specific version]({{<ref "#install-a-specific-version-of-the-cli" >}}). Make sure that your `meshctl` and Gloo Mesh versions match.

{{% notice note %}}
Upgrading the `meshctl` CLI does _not_ [upgrade the Gloo Mesh version]({{% versioned_link_path fromRoot="/operations/upgrading/" %}}) that you run in your clusters.
{{% /notice %}}

## Uninstall the CLI

To uninstall `meshctl`, you can delete the executable file from your system, such as on macOS in the following example.

```shell
rm ~/.gloo-mesh/bin/meshctl
```

## Skew policy

Use the same version of the `meshctl` CLI as the Gloo Mesh version that is installed in your clusters.

* Slight skews within minor versions typically work, such as `meshctl` 1.0.7 and Gloo Mesh 1.0.10. 
* Compatibility across beta versions is not guaranteed, even within minor version skews.
* To resolve bugs in `meshctl`, you might have to upgrade the entire CLI to a specific or latest version.

## Configuration file

When you run certain `meshctl` commands, the commands attempt to read a configuration file located at `$HOME/.gloo-mesh/meshctl-config.yaml`. This configuration file contains information about your management and remote clusters. For example, if `cluster-1` serves as the management cluster and both `cluster-1` and `cluster-2` are registered as remote clusters, `meshctl-config.yaml` might look like the following:

```yaml
apiVersion: v1
clusters:
  cluster1: # data plane cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: kind-cluster-1
  cluster2: # data plane cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: kind-cluster-2
  managementPlane: # management cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: kind-cluster-1
```

If you change your setup, you can configure the `meshctl-config.yaml` file by using the following command, which opens an
interactive prompt. You can also configure this file non-interactively by using the `--disable-prompt` flag.
```shell
meshctl cluster configure
```

Additionally, you can override the default `$HOME/.gloo-mesh/meshctl-config.yaml` file location when you run `meshctl` commands by setting a location in the `--config` flag.

## Reference documentation

For more information about each `meshctl` command, see the [CLI documentation]({{< versioned_link_path fromRoot="/reference/cli/" >}}) or run the help flag for a command.

```shell
meshctl cluster --help
```
