package check

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type serverOpts struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	localPort   uint32
	remotePort  uint32
}

func (o *serverOpts) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
	flags.Uint32Var(&o.localPort, "local-port", defaults.MetricsPort, "local port used to open port-forward to enterprise mgmt pod (enterprise only)")
	flags.Uint32Var(&o.remotePort, "remote-port", defaults.MetricsPort, "remote port used to open port-forward to enterprise mgmt pod (enterprise only). set to 0 to disable checks on the mgmt server")
}

func serverCmd(ctx context.Context, inCluster bool) *cobra.Command {
	opts := &serverOpts{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Perform post-install checks on the Gloo Mesh management plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServerChecks(ctx, opts, inCluster)
		},
	}

	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	return cmd
}

func runServerChecks(ctx context.Context, opts *serverOpts, inCluster bool) error {
	var checkCtx checks.CheckContext
	var err error

	if inCluster {
		checkCtx, err = validation.NewInClusterCheckContext()
		if err != nil {
			return eris.Wrapf(err, "could not construct in cluster check context")
		}
	} else {
		checkCtx, err = validation.NewOutOfClusterCheckContext(opts.kubeconfig, opts.kubecontext, opts.namespace, opts.localPort, opts.remotePort, nil, false, nil)
		if err != nil {
			return eris.Wrapf(err, "invalid kubeconfig")
		}
	}

	if foundFailure := checks.RunChecks(ctx, checkCtx, checks.Server, checks.PostInstall); foundFailure {
		return eris.New("Encountered failed checks.")
	}
	return nil
}
