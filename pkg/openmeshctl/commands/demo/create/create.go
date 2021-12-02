package create

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"

	cliversion "github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/demo/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
)

const (
	// The default version of k8s under Linux is 1.18 https://github.com/solo-io/gloo-mesh/issues/700
	kindImage        = "kindest/node:v1.17.17@sha256:66f1d0d91a88b8a001811e2f1054af60eef3b669a9a74f9b6db871f2f1eeed00"
	managementPort   = "32001"
	remotePort       = "32000"
	mgmtCluster      = "mgmt-cluster"
	remoteCluster    = "remote-cluster"
	agentReleaseName = "cert-agent"
)

func Command(ctx runtime.Context) *cobra.Command {
	opts := flags.Options{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Bootstrap a multicluster Istio demo with Gloo Mesh",
		Long: `
Bootstrap a multicluster Istio demo with Gloo Mesh.

Running the Gloo Mesh demo setup locally requires 4 tools to be installed and 
accessible via your PATH: kubectl >= v1.18.8, kind >= v0.8.1, istioctl, and docker.
We recommend allocating at least 8GB of RAM for Docker.

This command will bootstrap 2 clusters, one of which will run the Gloo Mesh
management-plane as well as Istio, and the other will just run Istio.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return initIstioCmd(ctx, mgmtCluster, remoteCluster, opts)
		},
	}
	opts.AddToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

func initIstioCmd(ctx runtime.Context, mgmtCluster, remoteCluster string, opts flags.Options) error {
	box := packr.NewBox("./scripts")

	// make sure istio version is supported
	if err := checkIstioVersion(box); err != nil {
		return err
	}

	// create management cluster in kind
	if err := createKindCluster(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	// install istio on management cluster
	if err := installIstio(mgmtCluster, managementPort, box); err != nil {
		return err
	}

	// create remote cluster in kind
	if err := createKindCluster(remoteCluster, remotePort, box); err != nil {
		return err
	}

	// install istio on remote cluster
	if err := installIstio(remoteCluster, remotePort, box); err != nil {
		return err
	}

	if opts.SkipGMInstall {
		// Skip installing Gloo Mesh, finish up now.
		return nil
	}

	// install GlooMesh on management cluster
	if err := switchContext(mgmtCluster, box); err != nil {
		return err
	}
	if err := installGlooMesh(ctx, mgmtCluster, opts, box); err != nil {
		return err
	}

	// register management cluster
	registerCmd := register.Command(ctx)
	registerMgmtWorkerArgs := []string{
		mgmtCluster,
		fmt.Sprintf("--agent-release-name=%s", agentReleaseName),
		fmt.Sprintf("--agent-version=%s", opts.Version),
	}
	registerCmd.SetArgs(registerMgmtWorkerArgs)
	err := registerCmd.Execute()
	if err != nil {
		return err
	}

	// register remote cluster
	switchContext(mgmtCluster, box)
	registerRemoteWorkerArgs := []string{
		remoteCluster,
		fmt.Sprintf("--agent-version=%s", opts.Version),
		fmt.Sprintf("--agent-release-name=%s", agentReleaseName),
	}
	registerCmd.SetArgs(registerRemoteWorkerArgs)
	err = registerCmd.Execute()
	if err != nil {
		return err
	}

	// set context to management cluster
	return switchContext(mgmtCluster, box)
}

func installIstio(cluster string, port string, box packr.Box) error {
	fmt.Printf("Installing Istio to cluster %s\n", cluster)
	if err := runScript(box, os.Stdout, "install_istio.sh", cluster, port); err != nil {
		return eris.Wrapf(err, "Error installing Istio on cluster %s", cluster)
	}

	fmt.Printf("Successfully installed Istio on cluster %s\n", cluster)
	return nil
}

func checkIstioVersion(box packr.Box) error {
	var buf bytes.Buffer
	if err := runScript(box, &buf, "check_istio_version.sh"); err != nil {
		return eris.Wrap(err, buf.String())
	}

	return nil
}

func createKindCluster(cluster string, port string, box packr.Box) error {
	fmt.Printf("Creating cluster %s with ingress port %s\n", cluster, port)

	script, err := box.FindString("create_kind_cluster.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", "-c", script, cluster, port, kindImage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return eris.Wrapf(err, "Error creating cluster %s", cluster)
	}

	fmt.Printf("Successfully created cluster %s\n", cluster)
	return nil
}

func getGlooMeshVersion(opts flags.Options) (string, error) {
	// Use user provided version first
	if opts.Version != "" {
		return opts.Version, nil
	}

	// Default to the CLI version
	return cliversion.Version, nil

}

func installGlooMesh(ctx runtime.Context, cluster string, opts flags.Options, box packr.Box) error {
	version, err := getGlooMeshVersion(opts)
	if err != nil {
		return err
	}

	return installGlooMeshCommunity(ctx, cluster, version, opts.Chart, box)
}

func installGlooMeshCommunity(ctx runtime.Context, cluster, version string, chart string, box packr.Box) error {
	fmt.Printf("Deploying Gloo Mesh to %s from images\n", cluster)
	installCmd := install.Command(ctx)
	if chart != "" {
		installCmd.SetArgs([]string{fmt.Sprintf("--chart=%s", chart)})
	} else {
		installCmd.SetArgs([]string{fmt.Sprintf("--version=%s", version)})

	}
	err := installCmd.Execute()
	if err != nil {
		return err
	}

	if err = glooMeshPostInstall(cluster, box); err != nil {
		return err
	}

	fmt.Printf("Successfully set up Gloo Mesh on cluster %s\n", cluster)
	return nil
}

func switchContext(cluster string, box packr.Box) error {
	if err := runScript(box, os.Stdout, "switch_context.sh", cluster); err != nil {
		return eris.Wrapf(err, "Could not switch context to %s", fmt.Sprintf("kind-%s", cluster))
	}

	return nil
}

func glooMeshPostInstall(cluster string, box packr.Box) error {
	if err := runScript(box, os.Stdout, "post_install_gloomesh.sh", cluster); err != nil {
		return eris.Wrap(err, "Error running post-install script")
	}

	return nil
}

func runScript(box packr.Box, out io.Writer, script string, args ...string) error {
	script, err := box.FindString(script)
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", append([]string{"-c", script}, args...)...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
