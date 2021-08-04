package check

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type agentOpts struct {
	kubeconfig  string
	kubecontext string
	namespace   string
}

func (o *agentOpts) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.kubeconfig, "kubeconfig", "", "Path to the kubeconfig from which the remote cluster will be accessed.")
	flags.StringVar(&o.kubecontext, "kubecontext", "", "Name of the kubeconfig context to use for the remote cluster.")
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "Namespace that Gloo Mesh is installed in.")
}

// run agent post-install checks
func agentCmd(ctx context.Context, inCluster bool) *cobra.Command {
	opts := &agentOpts{}

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Perform post-install checks on a Gloo Mesh agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAgentChecks(ctx, opts, inCluster)
		},
	}

	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	return cmd
}

func runAgentChecks(ctx context.Context, opts *agentOpts, inCluster bool) error {
	var checkCtx checks.CheckContext
	var err error

	if inCluster {
		checkCtx, err = validation.NewInClusterCheckContext()
		if err != nil {
			return eris.Wrapf(err, "could not construct in cluster check context")
		}
	} else {
		checkCtx, err = validation.NewOutOfClusterCheckContext(opts.kubeconfig, opts.kubecontext, opts.namespace, 0, 0, nil, false, nil)
		if err != nil {
			return eris.Wrapf(err, "could not construct out of cluster check context")
		}
	}

	if foundFailure := checks.RunChecks(ctx, checkCtx, checks.Agent, checks.PostInstall); foundFailure {
		return eris.New("Encountered failed checks.")
	}

	return nil
}
