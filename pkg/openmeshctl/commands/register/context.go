package register

import (
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/defaults"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/cli/values"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockgen -destination mocks/context.go -package mock_register . Context

// Context contains all the values for getting resources.
type Context interface {
	runtime.Context

	// AddToFlags adds the configurable context values to the given flag set.
	AddToFlags(flags *pflag.FlagSet)

	// AgentChart returns the agent chart to install. Will either be a name or URL.
	AgentChart() string

	// AgentChartOptions returns the values options for a chart.
	AgentChartOptions() *values.Options

	// AgentVersion returns the version of the cert agent chart to install.
	AgentVersion() string

	// AgentReleaseName returns the name of the release to use when installing
	// cert agent the Helm chart.
	AgentReleaseName() string

	// AgentCRDsChart returns the agent CRDs chart to install.
	// Will either be a name or URL.
	AgentCRDsChart() string

	// AgentCRDsReleaseName returns the name of the release to use when
	// installing the cert agent CRDs Helm chart.
	AgentCRDsReleaseName() string

	// HelmInstaller returns a client to install Helm charts.
	HelmInstaller() (helm.Installer, error)

	// ClusterRegistry returns a cluster registry that is used to add remote
	// clusters to the management plna.
	ClusterRegistry() (ClusterRegistry, error)
}

type context struct {
	runtime.Context

	remoteKubeCfg          string
	remoteNamespace        string
	agentChartOverride     string
	agentChartOptions      values.Options
	agentVersionOverride   string
	agentCRDsChartOverride string
	agentReleaseName       string
	agentCRDsReleaseName   string
	clusterDomain          string
	apiServerAddress       string
}

// NewContext creates a new installation context from the root context.
func NewContext(rootCtx runtime.Context) Context {
	return &context{Context: rootCtx}
}

// AddToFlags implements the Context interface.
func (ctx *context) AddToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&ctx.remoteKubeCfg, "remote-kubeconfig", "",
		"Path to the kubeconfig from which the remote cluster well be accessed if different from the management cluster.")
	flags.StringVar(&ctx.remoteNamespace, "remote-namespace", defaults.Namespace,
		"Namespace on the remote cluster that the agent will be installed to.")
	helm.AddChartOptionToFlags(flags, &ctx.agentChartOverride, "agent-", "cert agent")
	helm.AddChartOptionToFlags(flags, &ctx.agentCRDsChartOverride, "agent-crds-", "cert agent CRDs")
	helm.AddValueOptionsToFlags(flags, &ctx.agentChartOptions, "agent-crds-", "agent CRDs")
	helm.AddVersionOptionToFlags(flags, &ctx.agentVersionOverride, "agent-", "cert agent")
	helm.AddReleaseNameOptionToFlags(flags, &ctx.agentReleaseName, "agent-", "cert agent", defaults.AgentReleaseName)
	helm.AddReleaseNameOptionToFlags(flags, &ctx.agentCRDsReleaseName, "agent-crds-", "cert agent CRDs",
		defaults.AgentCRDsReleaseName)
	flags.StringVar(&ctx.clusterDomain, "cluster-domain", defaults.ClusterDomain,
		"The cluster domain used by the Kubernetes DNS Service in the registered cluster.\n"+
			"Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	flags.StringVar(&ctx.apiServerAddress, "api-server-address", "",
		"Swap out the address of the remote cluster's k8s API server for the value of this flag.\n"+
			"Set this flag when the address of the cluster domain used by the Gloo Mesh is different than that "+
			"specified in the local kubeconfig.")
}

// AgentChart implements the Context interface.
func (ctx *context) AgentChart() string {
	if ctx.agentChartOverride != "" {
		return ctx.agentChartOverride
	}

	return defaults.CertAgentChartURI(ctx.AgentVersion())
}

// AgentChartOptions implements the Context interface.
func (ctx *context) AgentChartOptions() *values.Options {
	return &ctx.agentChartOptions
}

// AgentVersion implements the Context interface.
func (ctx *context) AgentVersion() string {
	if ctx.agentVersionOverride != "" {
		return ctx.agentVersionOverride
	}

	return version.Version
}

// AgentReleaseName implements the Context interface.
func (ctx *context) AgentReleaseName() string {
	return ctx.agentReleaseName
}

// AgentCRDsChart implements the Context interface.
func (ctx *context) AgentCRDsChart() string {
	if ctx.agentCRDsChartOverride != "" {
		return ctx.agentCRDsChartOverride
	}

	return defaults.AgentCRDsChartURI(ctx.AgentVersion())
}

// AgentCRDsReleaseName implements the Context interface.
func (ctx *context) AgentCRDsReleaseName() string {
	return ctx.agentCRDsReleaseName
}

// HelmInstaller implements the Context interface.
func (ctx *context) HelmInstaller() (helm.Installer, error) {
	return helm.NewClient(ctx)
}

// ClusterRegistry implements the Context interface.
func (ctx *context) ClusterRegistry() (ClusterRegistry, error) {
	policyRules := []rbacv1.PolicyRule{}
	policyRules = append(policyRules, io.DiscoveryInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.LocalNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.IstioNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.SmiNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesUpdateStatus()...)

	clusterRegistry := &register.RegistrationOptions{
		KubeCfg:          ctx.ToRawKubeConfigLoader(),
		APIServerAddress: ctx.apiServerAddress,
		Namespace:        ctx.Namespace(),
		RemoteNamespace:  ctx.remoteNamespace,
		ClusterDomain:    ctx.clusterDomain,
		ClusterRoles: []*rbacv1.ClusterRole{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gloomesh-remote-access",
					Namespace: ctx.remoteNamespace,
				},
				Rules: policyRules,
			},
		},
	}

	return RegistryFunc(func(subCtx Context, name, context string) error {
		// copy clusterRegistry to avoid races between calls to this function
		clusterRegistryLocal := *clusterRegistry

		clusterRegistryLocal.RemoteCtx = context
		clusterRegistryLocal.ClusterName = name

		remoteCfg, err := kubeconfig.GetClientConfigWithContext(ctx.remoteKubeCfg, context, "")
		if err != nil {
			return err
		}
		clusterRegistryLocal.RemoteKubeCfg = remoteCfg

		return clusterRegistryLocal.RegisterCluster(subCtx)
	}), nil
}

// RegistryFunc converts a function to a ClusterRegistry object.
type RegistryFunc func(ctx Context, name, context string) error

// RegisterCluster implements the ClusterRegistry interface.
func (f RegistryFunc) RegisterCluster(ctx Context, name, context string) error {
	return f(ctx, name, context)
}
