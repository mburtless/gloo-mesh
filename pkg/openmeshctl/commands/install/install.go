package install

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/utils"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
)

// Command creates a new install command that can be attached to the root command tree.
func Command(rootCtx runtime.Context) *cobra.Command {
	ctx := NewContext(rootCtx)
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "Install Gloo Mesh",
		Long:         `Install the Gloo Mesh management plan to a Kubernetes cluster.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Install(ctx)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// TODO(ryantking): Update the meshctl config file
		},
	}

	ctx.AddToFlags(cmd.Flags())

	return cmd
}

// Install runs the Gloo Mesh installation.
func Install(ctx Context) error {
	mgmtContext, err := ctx.KubeContext()
	if err != nil {
		return eris.Wrap(err, "Error determining desired kube context")
	}

	// Install Gloo Mesh on the desired mgmt cluster
	if err := utils.SwitchContext(mgmtContext); err != nil {
		return eris.Wrapf(err, "Error switching to management context %s", mgmtContext)
	}
	output.DebugPrint(ctx, "Using chart %s\n", ctx.Chart())
	output.DebugPrint(ctx, "Using version %s\n", ctx.Version())
	s, cancel := output.MakeSpinner(ctx, "Installing Gloo Mesh version %s...", ctx.Version())
	s.Start()
	if err := install(ctx); err != nil {
		cancel()
		return err
	}
	s.Stop()

	if ctx.Register() {
		return registerCluster(ctx, mgmtContext)
	}

	return nil
}

func install(ctx Context) error {
	helmInstaller, err := ctx.HelmInstaller()
	if err != nil {
		return err
	}
	chartSpec := helm.ChartSpec{
		Name:          ctx.Chart(),
		Namespace:     ctx.Namespace(),
		Version:       ctx.Version(),
		ValuesOptions: ctx.ChartOptions(),
	}

	return helmInstaller.Install(ctx, chartSpec, ctx.ReleaseName(), "")
}

func registerCluster(ctx Context, context string) error {
	if !ctx.Register() {
		output.DebugPrint(ctx, "skipping registration of management cluster")
		return nil
	}

	return register.Register(ctx, ctx.ClusterName(), context)
}
