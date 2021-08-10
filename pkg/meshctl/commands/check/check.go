package check

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/logrusorgru/aurora/v3"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// inCluster is true if running from a Helm Hook context, else for meshctl it's false
func Command(ctx context.Context, inCluster bool) *cobra.Command {
	opts := &opts{}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Gloo Mesh system",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAllChecks(ctx, opts, inCluster)
		},
	}

	cmd.AddCommand(
		serverCmd(ctx, inCluster),
		agentCmd(ctx, inCluster),
	)

	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

// options for running `meshctl check`, which checks all clusters
type opts struct {
	configFile string
	localPort  uint32
	remotePort uint32
}

func (o *opts) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.configFile, "config", utils.DefaultConfigPath, "set the path to the meshctl config file")
	flags.Uint32Var(&o.localPort, "local-port", defaults.MetricsPort, "local port used to open port-forward to enterprise-networking pod")
	flags.Uint32Var(&o.remotePort, "remote-port", defaults.MetricsPort, "remote port used to open port-forward to enterprise-networking pod")
}

// run post install checks for all clusters in the system (i.e. management cluster and all remote clusters)
func runAllChecks(ctx context.Context, opts *opts, inCluster bool) error {
	config, err := utils.ParseMeshctlConfig(opts.configFile)
	if err != nil {
		return eris.Wrapf(err, "invalid meshctl config file, configure it with `meshctl cluster configure` and try again")
	}

	var errs error

	// check management cluster
	managementCluster := config.MgmtCluster()

	logrus.Info(aurora.Bold("🔎 Checking Management Plane\n"))
	if err := runServerChecks(ctx, &serverOpts{
		kubeconfig:  managementCluster.KubeConfig,
		kubecontext: managementCluster.KubeContext,
		namespace:   managementCluster.Namespace,
		localPort:   opts.localPort,
		remotePort:  opts.remotePort,
	}, inCluster); err != nil {
		errs = multierror.Append(errs, err)
	}

	// check remote clusters
	for clusterName, remoteCluster := range config.RemoteClusters() {
		logrus.Print("\n")
		logrus.Info(aurora.Bold(fmt.Sprintf("🔎 Checking Remote Cluster \"%s\"\n", clusterName)))

		if err := runAgentChecks(ctx, &agentOpts{
			kubeconfig:  remoteCluster.KubeConfig,
			kubecontext: remoteCluster.KubeContext,
			namespace:   remoteCluster.Namespace,
		}, inCluster); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}
