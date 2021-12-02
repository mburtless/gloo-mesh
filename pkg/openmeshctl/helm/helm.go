package helm

import (
	"fmt"

	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/pflag"
	"helm.sh/helm/v3/pkg/cli/values"
)

//go:generate mockgen -destination mocks/installer.go -package mock_helm . Installer

// Installer is capable of installing a Helm chart on a cluster.
type Installer interface {
	// Install a chart with the given release name.
	Install(ctx runtime.Context, chart ChartSpec, releaseName string, kubeContext string) error
}

//go:generate mockgen -destination mocks/uninstaller.go -package mock_helm . Uninstaller

// Uninstaller is capable of uninstalling a Helm release.
type Uninstaller interface {
	// Uninstall a Helm release .
	Uninstall(ctx runtime.Context, releaseName string) error
}

// ChartSpec contains the configuration options for a Helm chart installation.
type ChartSpec struct {
	// Name of the chart to install, can also be a URL to a local or remote chart.
	Name string

	// Namespace to install the chart in.
	Namespace string

	// Version of the chart to install.
	// Only used when installing a name chart from a repository.
	Version string

	// Options for the chart values.
	ValuesOptions *values.Options
}

// AddChartOptionToFlags adds the chart name to the given flag set.
// Supports a prefix to enable adding multiple charts to a single flagset.
func AddChartOptionToFlags(flags *pflag.FlagSet, chart *string, prefix, chartName string) {
	const chartUsageFmt = "Name of the %s chart to install.\n" +
		"Can be a URI to a local or remote chart or the name of a chart known to Helm.\n" +
		"Defaults to the URI to the official Gloo Mesh Helm repository."

	flags.StringVar(chart, prefix+"chart", "", fmt.Sprintf(chartUsageFmt, chartName))
}

// AddVersionOptionToFlags adds the chart version to the given flag set.
// Supports a prefix to enable adding multiple charts to a single flagset.
func AddVersionOptionToFlags(flags *pflag.FlagSet, version *string, prefix, chartName string) {
	const versionUsageFmt = "Specific version of the %s chart to install.\nDefaults to the installed CLI version"

	flags.StringVar(version, prefix+"version", "", fmt.Sprintf(versionUsageFmt, chartName))
}

// AddValueOptionsToFlags adds the value options to the given flag sets.
// Supports a prefix to enable adding multiple sets of values to a single flagset.
func AddValueOptionsToFlags(flags *pflag.FlagSet, opts *values.Options, prefix, chartName string) {
	const (
		valuesUsageFmt = "Specify values in a YAML file or a URL (can specify multiple) for the %s chart"
		setUsageString = "Set values on the command line for the %s chart " +
			"(can specify multiple or separate values with commas: key1=val1,key2=val2)"
		setStringUsageString = "Set STRING values on the command line for the %s chart " +
			"(can specify multiple or separate values with commas: key1=val1,key2=val2)"
		setFileUsageString = "Set values from respective files specified via the command line for the %s chart " +
			"(can specify multiple or separate values with commas: key1=path1,key2=path2)"
	)

	flags.StringSliceVar(&opts.ValueFiles, prefix+"values", []string{}, fmt.Sprintf(valuesUsageFmt, chartName))
	flags.StringArrayVar(&opts.Values, prefix+"set", []string{}, fmt.Sprintf(setUsageString, chartName))
	flags.StringArrayVar(&opts.StringValues, prefix+"set-string", []string{},
		fmt.Sprintf(setStringUsageString, chartName))
	flags.StringArrayVar(&opts.FileValues, prefix+"set-file", []string{}, fmt.Sprintf(setFileUsageString, chartName))
}

// AddReleaseNameOptionToFlags adds the value options to the given flag sets.
// Supports a prefix to enable adding multiple sets of values to a single flagset.
func AddReleaseNameOptionToFlags(flags *pflag.FlagSet, releaseName *string, prefix, chartName, defValue string) {
	const releaseNameUsageFmt = "Helm release name for the %s chart. Defaults to '%s'"

	flags.StringVar(releaseName, prefix+"release-name", defValue, fmt.Sprintf(releaseNameUsageFmt, chartName, defValue))
}
