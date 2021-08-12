package check

import (
	"context"
	"os"

	"github.com/rotisserie/eris"
	v1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/checks"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/validation/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
)

// helm release name annotation source of truth is here, https://github.com/helm/helm/blob/a499b4b179307c267bdf3ec49b880e3dbd2a5591/pkg/action/validate.go#L36
const helmReleaseNameAnnotation = "meta.helm.sh/release-name"

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
		releaseName, err := fetchEnterpriseAgentHelmReleaseName(ctx, opts.kubeconfig, opts.kubecontext, opts.namespace)
		if err != nil {
			return err
		}

		// if out of cluster, execute Helm test so that the checks run in-cluster
		installer := &helm.Installer{
			ReleaseName: releaseName,
			KubeConfig:  opts.kubeconfig,
			KubeContext: opts.kubecontext,
			Namespace:   opts.namespace,
			Verbose:     true,
			Output:      os.Stdout,
		}

		if err := installer.ExecuteHelmTest(); err != nil {
			return eris.Wrapf(err, "executing Helm test for release \"%s\" in namespace \"%s\"", releaseName, opts.namespace)
		}
	}

	return nil
}

// fetch the Helm chart corresponding to the version found in the enterprise-agent deployment
func fetchEnterpriseAgentHelmReleaseName(
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

	releaseName := deployment.GetAnnotations()[helmReleaseNameAnnotation]
	if releaseName == "" {
		return "", eris.Errorf(
			"Gloo Mesh agent deployment %s.%s is missing Helm release annotation \"%s\" (meshctl check requires that the component is installed via Helm)",
			deployment.GetName(),
			deployment.GetNamespace(),
			helmReleaseNameAnnotation,
		)
	}

	return releaseName, nil
}
