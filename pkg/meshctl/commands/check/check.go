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

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Gloo Mesh system",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChecks(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
	localPort   uint32
	remotePort  uint32
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	utils.AddManagementKubeconfigFlags(&o.kubeconfig, &o.kubecontext, flags)
	flags.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Gloo Mesh is installed in")
	flags.Uint32Var(&o.localPort, "local-port", defaults.MetricsPort, "local port used to open port-forward to enterprise mgmt pod (enterprise only)")
	flags.Uint32Var(&o.remotePort, "remote-port", defaults.MetricsPort, "remote port used to open port-forward to enterprise mgmt pod (enterprise only). set to 0 to disable checks on the mgmt server")
}

func runChecks(ctx context.Context, opts *options) error {
	checkCtx, err := validation.NewOutOfClusterCheckContext(
		opts.kubeconfig,
		opts.kubecontext,
		opts.namespace,
		opts.localPort,
		opts.remotePort,
		nil, // server post install check doesn't require validating install parameters
		false,
	)
	if err != nil {
		return eris.Wrapf(err, "invalid kubeconfig")
	}

	if foundFailure := checks.RunChecks(ctx, checkCtx, checks.Server, checks.PostInstall); foundFailure {
		return eris.New("Encountered failed checks.")
	}
	return nil
}
