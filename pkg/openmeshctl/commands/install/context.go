package install

import (
	"github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/defaults"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/cli/values"
)

//go:generate mockgen -destination mocks/context.go -package mock_install . Context

// Context contains all the values for installing Gloo Mesh.
type Context interface {
	runtime.Context
	register.Context

	// AddToFlags adds the configurable context values to the given flag set.
	AddToFlags(flags *pflag.FlagSet)

	// Chart returns the chart to install. Will either be a name or URL.
	Chart() string

	// ChartOptions returns the values options for a chart.
	ChartOptions() *values.Options

	// Version returns the version of the chart to install.
	Version() string

	// ReleaseName returns the name of the release to use when installing the Helm chart.
	ReleaseName() string

	// Register returns whether or not to also register the management cluster.
	Register() bool

	// ClusterName is the cluster name to use when registering the management
	// cluster as a data plane cluster.
	ClusterName() string

	// HelmInstaller returns a client to install Helm charts.
	HelmInstaller() (helm.Installer, error)
}

type context struct {
	register.Context

	chartOverride   string
	chartOptions    values.Options
	versionOverride string
	releaseName     string
	register        bool
	clusterName     string
}

// NewContext creates a new installation context from the root context.
func NewContext(rootCtx runtime.Context) Context {
	regCtx := register.NewContext(rootCtx)
	return &context{Context: regCtx}
}

// AddToFlags implements the Context interface.
func (ctx *context) AddToFlags(flags *pflag.FlagSet) {
	helm.AddChartOptionToFlags(flags, &ctx.chartOverride, "", "Gloo Mesh")
	helm.AddVersionOptionToFlags(flags, &ctx.versionOverride, "", "Gloo Mesh")
	helm.AddReleaseNameOptionToFlags(flags, &ctx.releaseName, "", "Gloo Mesh", defaults.ReleaseName)
	helm.AddValueOptionsToFlags(flags, &ctx.chartOptions, "", "Gloo Mesh")
	flags.BoolVarP(&ctx.register, "register", "r", false, "Register the management cluster as a data plane cluster.")
	flags.StringVar(&ctx.clusterName, "cluster-name", "",
		"When register is enabled, the name of the cluster to register the management cluster as.\n"+
			"Defaults to the name of the context.")
	ctx.Context.AddToFlags(flags)
}

// Chart implements the Context interface.
func (ctx *context) Chart() string {
	if ctx.chartOverride != "" {
		return ctx.chartOverride
	}

	return defaults.GlooMeshChartURI(ctx.Version())
}

// ChartOptions implements the Context interface.
func (ctx *context) ChartOptions() *values.Options {
	return &ctx.chartOptions
}

// Version implements the Context interface.
func (ctx *context) Version() string {
	if ctx.versionOverride != "" {
		return ctx.versionOverride
	}

	return version.Version
}

// ReleaseName implements the Context interface.
func (ctx *context) ReleaseName() string {
	return ctx.releaseName
}

// Register implements the Context interface.
func (ctx *context) Register() bool {
	return ctx.register
}

// ClusterName implements the Context interface.
func (ctx *context) ClusterName() string {
	return ctx.clusterName
}

// HelmInstaller implements the Context interface.
func (ctx *context) HelmInstaller() (helm.Installer, error) {
	return helm.NewClient(ctx)
}
