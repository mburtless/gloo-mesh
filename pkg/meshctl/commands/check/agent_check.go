package check

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
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
		// triggered by helm test

		checkCtx, err = validation.NewInClusterCheckContext()
		if err != nil {
			return eris.Wrapf(err, "could not construct in cluster check context")
		}

		if foundFailure := checks.RunChecks(ctx, checkCtx, checks.Agent, checks.PostInstall); foundFailure {
			return eris.New("Encountered failed checks.")
		}
	} else {

		// fetch the Helm chart corresponding to the version of the enterprise-agent deployment
		// which will be used to subsequently run its Helm test
		version, err := fetchEnterpriseAgentVersionFromDeployment(ctx, opts.kubeconfig, opts.kubecontext, opts.namespace)
		if err != nil {
			return err
		}

		// if out of cluster, execute Helm test so that the checks run in-cluster
		installer := &helm.Installer{
			KubeConfig:  opts.kubeconfig,
			KubeContext: opts.kubecontext,
			ChartUri:    fmt.Sprintf(gloomesh.EnterpriseAgentChartUriTemplate, version),
			Namespace:   opts.namespace,
			ReleaseName: "enterprise-agent-test",
			Verbose:     true,
			Output:      os.Stdout,
		}

		if err := installer.ExecuteHelmTest(ctx, 60*time.Second); err != nil {
			return err
		}
	}

	return nil
}

// fetch the Helm chart corresponding to the version found in the enterprise-agent deployment
func fetchEnterpriseAgentVersionFromDeployment(
	ctx context.Context,
	kubeconfig, kubecontext, namespace string,
) (string, error) {
	client, err := utils.BuildClient(kubeconfig, kubecontext)
	if err != nil {
		return "", err
	}

	deploymentClient := v1.NewDeploymentClient(client)
	deployment, err := deploymentClient.GetDeployment(ctx, client2.ObjectKey{
		Name:      consts.AgentDeployName,
		Namespace: namespace,
	})
	if err != nil {
		return "", eris.Wrapf(err, "could not find Gloo Mesh agent deployment")
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == consts.AgentDeployName {
			image, err := dockerutils.ParseImageName(container.Image)
			if err != nil {
				return "", eris.Wrapf(err, "parsing Gloo Mesh agent deployment image string")
			}
			return image.Tag, nil
		}
	}

	return "", eris.New("Gloo Mesh agent deployment not found")
}
