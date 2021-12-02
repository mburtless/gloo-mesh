package deregister

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/defaults"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/pflag"
)

//go:generate mockgen -destination mocks/context.go -package mock_deregister . Context

// Context contains all the values for getting resources.
type Context interface {
	runtime.Context

	// AddToFlags adds the configurable context values to the given flag set.
	AddToFlags(flags *pflag.FlagSet)

	// AgentReleaseName returns the name of the release to use when uninstalling
	// the cert agent the Helm chart.
	AgentReleaseName() string

	// AgentCRDsReleaseName returns the name of the release to use when
	// uninstalling the cert agent CRDs Helm chart.
	AgentCRDsReleaseName() string

	// HelmUninstaller returns a client to uninstall Helm charts.
	HelmUninstaller() (helm.Uninstaller, error)

	// ClusterDeregistry returns a cluster registry that is used to add remote
	// clusters to the management plane.
	ClusterDeregistry() (ClusterDeregistry, error)
}

type context struct {
	runtime.Context

	remoteKubeCfg        string
	remoteNamespace      string
	agentReleaseName     string
	agentCRDsReleaseName string
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
	helm.AddReleaseNameOptionToFlags(flags, &ctx.agentReleaseName, "agent-", "cert agent", defaults.AgentReleaseName)
	helm.AddReleaseNameOptionToFlags(flags, &ctx.agentCRDsReleaseName, "agent-crds-", "cert agent CRDs",
		defaults.AgentCRDsReleaseName)
}

// AgentReleaseName implements the Context interface.
func (ctx *context) AgentReleaseName() string {
	return ctx.agentReleaseName
}

// AgentCRDsReleaseName implements the Context interface.
func (ctx *context) AgentCRDsReleaseName() string {
	return ctx.agentCRDsReleaseName
}

// HelmUninstaller implements the Context interface.
func (ctx *context) HelmUninstaller() (helm.Uninstaller, error) {
	return helm.NewClient(ctx)
}

// ClusterDeregistry implements the Context interface.
func (ctx *context) ClusterDeregistry() (ClusterDeregistry, error) {
	clusterRegistry := &register.RegistrationOptions{
		KubeCfg:         ctx.ToRawKubeConfigLoader(),
		Namespace:       ctx.Namespace(),
		RemoteNamespace: ctx.remoteNamespace,
	}

	return DeregistryFunc(func(subCtx Context, name, context string) error {
		// copy clusterRegistry to avoid races between calls to this function
		clusterRegistryLocal := *clusterRegistry

		clusterRegistryLocal.RemoteCtx = context
		clusterRegistryLocal.ClusterName = name

		remoteCfg, err := kubeconfig.GetClientConfigWithContext(ctx.remoteKubeCfg, context, "")
		if err != nil {
			return err
		}
		clusterRegistryLocal.RemoteKubeCfg = remoteCfg

		return clusterRegistryLocal.DeregisterCluster(subCtx)
	}), nil
}

// DeregistryFunc converts a function to a ClusterDeregistry object.
type DeregistryFunc func(ctx Context, name, context string) error

// RegisterCluster implements the ClusterDeregistry interface.
func (f DeregistryFunc) DeregisterCluster(ctx Context, name, context string) error {
	return f(ctx, name, context)
}
