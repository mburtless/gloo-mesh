package deregister

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
		Use:   "deregister NAME [REMOTE CONTEXT]",
		Short: "Deregister a data plane cluster",
		Long: `
Deregistering a cluster removes the cert agent along with other Gloo Mesh-owned
resources such as service accounts.

The context used is by default the same name as the cluster but can be changed
via the additional REMOTE CONTEXT argument.

The KubernetesCluster resource may not be uninstalled properly if the --context flag is not included.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return eris.New("must provide a cluster name")
			} else if len(args) > 2 {
				return eris.Errorf("unexpected arguments: %s", args[2:])
			}

			name := args[0]
			context := args[0]
			if len(args) == 2 {
				context = args[1]
			}

			return Deregister(ctx, name, context)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// TODO(ryantking): Update config
		},
	}

	ctx.AddToFlags(cmd.Flags())
	return cmd
}

//go:generate mockgen -destination mocks/deregistry.go -package mock_deregister . ClusterDeregistry

// ClusterDeregistry is capable of deregistering data plane clusters.
type ClusterDeregistry interface {
	DeregisterCluster(ctx Context, name, context string) error
}

// Deregister installs the cert agent on the remote cluster and configures the
// management plan to discover the remote resources.
func Deregister(ctx Context, name, context string) error {
	if err := utils.SwitchContext(context); err != nil {
		return eris.Wrapf(err, "Error switching to remote context %s", context)
	}
	helmUninstaller, err := ctx.HelmUninstaller()
	if err != nil {
		return err
	}
	s, cancel := output.MakeSpinner(ctx, "Uninstalling agent CRDs...")
	s.Start()
	if err := helmUninstaller.Uninstall(ctx, ctx.AgentCRDsReleaseName()); err != nil {
		cancel()
		return err
	}
	s.Stop()
	s, cancel = output.MakeSpinner(ctx, "Uninstalling cert agent...")
	s.Start()
	if err := helmUninstaller.Uninstall(ctx, ctx.AgentReleaseName()); err != nil {
		cancel()
		return err
	}
	s.Stop()
	s, cancel = output.MakeSpinner(ctx, "Deregistering cluster with name '%s'...", name)
	s.Start()
	if err := deregisterCluster(ctx, name, context); err != nil {
		cancel()
		return err
	}
	s.Stop()

	return nil
}

func deregisterCluster(ctx Context, name, context string) error {
	registry, err := ctx.ClusterDeregistry()
	if err != nil {
		return err
	}

	return registry.DeregisterCluster(ctx, name, context)
}
