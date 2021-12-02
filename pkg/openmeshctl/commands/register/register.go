package register

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/utils"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/cli/values"
)

// Command creates a new install command that can be attached to the root command tree.
func Command(rootCtx runtime.Context) *cobra.Command {
	ctx := NewContext(rootCtx)
	cmd := &cobra.Command{
		Use:   "register NAME [REMOTE CONTEXT]",
		Short: "Register a Gloo Mesh data plane cluster",
		Long: `Registering a cluster installs the cert agent to the remote cluster and
configures the management plan to discover resources on the remote cluster,
which can then be configured using the Gloo Mesh API.

The context used is by default the same name as the cluster but can be changed
via the additional REMOTE CONTEXT argument.

The KubernetesCluster resource may not be installed properly if the --context flag is not included.`,
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

			return Register(ctx, name, context)
		},
		PostRun: func(cmd *cobra.Command, args []string) {
			// TODO(ryantking): Update config
		},
	}

	ctx.AddToFlags(cmd.Flags())
	return cmd
}

//go:generate mockgen -destination mocks/registry.go -package mock_register . ClusterRegistry

// ClusterRegistry is capable of registering data plane clusters.
type ClusterRegistry interface {
	RegisterCluster(ctx Context, name, context string) error
}

// Register installs the cert agent on the remote cluster and configures the
// management plane to discover the remote resources.
func Register(ctx Context, name, remoteContext string) error {
	// Install the CRDs and the Agent on the remote cluster which we are registering
	if err := utils.SwitchContext(remoteContext); err != nil {
		return eris.Wrapf(err, "Error switching to remote context %s", remoteContext)
	}
	output.DebugPrint(ctx, "Using agent CRDs chart %s\n", ctx.AgentCRDsChart())
	output.DebugPrint(ctx, "Using agent chart %s\n", ctx.AgentChart())
	output.DebugPrint(ctx, "Using agent version %s\n", ctx.AgentVersion())
	s, cancel := output.MakeSpinner(ctx, "Installing agent CRDs version %s on kube context %s...", ctx.AgentVersion(), remoteContext)
	s.Start()
	if err := installCRDs(ctx, remoteContext); err != nil {
		cancel()
		return err
	}
	s.Stop()
	s, cancel = output.MakeSpinner(ctx, "Installing cert agent version %s on kube context %s...", ctx.AgentVersion(), remoteContext)
	s.Start()
	if err := installAgent(ctx, remoteContext); err != nil {
		cancel()
		return err
	}
	s.Stop()

	// Register the KubernetesCluster with the management cluster
	s, cancel = output.MakeSpinner(ctx, "Registering cluster with name '%s'...", name)
	s.Start()
	if err := registerCluster(ctx, name, remoteContext); err != nil {
		cancel()
		return err
	}
	s.Stop()

	return nil
}

func installCRDs(ctx Context, remoteContext string) error {
	helmInstaller, err := ctx.HelmInstaller()
	if err != nil {
		return err
	}

	agentCRDsSpec := helm.ChartSpec{
		Name:          ctx.AgentCRDsChart(),
		Namespace:     ctx.Namespace(),
		Version:       ctx.AgentVersion(),
		ValuesOptions: &values.Options{},
	}

	return helmInstaller.Install(ctx, agentCRDsSpec, ctx.AgentCRDsReleaseName(), remoteContext)
}

func installAgent(ctx Context, remoteContext string) error {
	helmInstaller, err := ctx.HelmInstaller()
	if err != nil {
		return err
	}

	agentSpec := helm.ChartSpec{
		Name:          ctx.AgentChart(),
		Namespace:     ctx.Namespace(),
		Version:       ctx.AgentVersion(),
		ValuesOptions: ctx.AgentChartOptions(),
	}

	return helmInstaller.Install(ctx, agentSpec, ctx.AgentReleaseName(), remoteContext)
}

func registerCluster(ctx Context, name, context string) error {
	registry, err := ctx.ClusterRegistry()
	if err != nil {
		return err
	}

	return registry.RegisterCluster(ctx, name, context)
}
