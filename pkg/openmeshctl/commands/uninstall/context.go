package uninstall

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/defaults"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/helm"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/pflag"
)

//go:generate mockgen -destination mocks/context.go -package mock_uninstall . Context

// Context contains all the values for uninstalling Gloo Mesh.
type Context interface {
	runtime.Context

	// AddToFlags adds the configurable context values to the given flag set.
	AddToFlags(flags *pflag.FlagSet)

	// ReleaseName returns the name of the release to use when uninstalling the Helm chart.
	ReleaseName() string

	// HelmUninstaller returns a client to install Helm charts.
	HelmUninstaller() (helm.Uninstaller, error)
}

type context struct {
	runtime.Context

	releaseName string
}

// NewContext creates a new uninstallation context from the root context.
func NewContext(rootCtx runtime.Context) Context {
	regCtx := register.NewContext(rootCtx)
	return &context{Context: regCtx}
}

// AddToFlags implements the Context interface.
func (ctx *context) AddToFlags(flags *pflag.FlagSet) {
	helm.AddReleaseNameOptionToFlags(flags, &ctx.releaseName, "", "Gloo Mesh", defaults.ReleaseName)
}

// ReleaseName implements the Context interface.
func (ctx *context) ReleaseName() string {
	return ctx.releaseName
}

// HelmUninstaller implements the Context interface.
func (ctx *context) HelmUninstaller() (helm.Uninstaller, error) {
	return helm.NewClient(ctx)
}
