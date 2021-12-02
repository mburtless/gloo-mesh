package uninstall

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/utils"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
)

// Command creates a new install command that can be attached to the root command tree.
func Command(rootCtx runtime.Context) *cobra.Command {
	ctx := NewContext(rootCtx)
	cmd := &cobra.Command{
		Use:          "uninstall",
		Short:        "Uninstall Gloo Mesh",
		Long:         `Uninstall the Gloo Mesh management plan from a Kubernetes cluster.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Uninstall(ctx)
		},
	}

	ctx.AddToFlags(cmd.Flags())
	return cmd
}

// Setup checks the user provided values, applies defaults, and sets up the Helm client.
// Run uninstalls the Gloo Mesh Helm chart.
func Uninstall(ctx Context) error {
	mgmtContext, err := ctx.KubeContext()
	if err != nil {
		return eris.Wrap(err, "Error determining desired kube context")
	}

	// Uninstall Gloo Mesh from the desired mgmt cluster
	if err := utils.SwitchContext(mgmtContext); err != nil {
		return eris.Wrapf(err, "Error switching to management context %s", mgmtContext)
	}

	s, cancel := output.MakeSpinner(ctx, "Uninstalling Gloo Mesh...")
	s.Start()
	defer cancel()
	helmUninstaller, err := ctx.HelmUninstaller()
	if err != nil {
		return err
	}

	if err := helmUninstaller.Uninstall(ctx, ctx.ReleaseName()); err != nil {
		return err
	}
	s.Stop()
	return nil
}
