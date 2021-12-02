package get

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/pflag"
)

//go:generate mockgen -destination mocks/context.go -package mock_get . Context

// Context contains all the values for getting resources.
type Context interface {
	runtime.Context

	// AddToFlags adds the configurable context values to the given flag set.
	AddToFlags(flags *pflag.FlagSet)

	// OutputFormat returns how the format to print objects in.
	OutputFormat() output.Format

	// Printer returns a printer to print objects to the console.
	Printer() output.Printer

	// Factory returns the factory to create empty resources/lists.
	Factory() resource.Factory

	// Formatter returns the formatter to format the objects in the desired shape.
	Formatter() resource.Formatter
}

type context struct {
	runtime.Context
	allNamespaces bool
	outputFormat  string
	printer       output.Printer
	factory       resource.Factory
	formatter     resource.Formatter
}

// NewContext returns a new get context that wraps the root context and uses the
// factory and formatter from the given resource config.
func NewContext(rootCtx runtime.Context, cfg resource.Config) Context {
	return &context{
		Context:   rootCtx,
		printer:   output.NewPrinter(rootCtx.Out()),
		factory:   cfg.Factory,
		formatter: cfg.Formatter,
	}
}

// AddToFlags implements the Context interface.
func (ctx *context) AddToFlags(flags *pflag.FlagSet) {
	flags.BoolVarP(&ctx.allNamespaces, "all-namespaces", "A", false,
		"List the requsted resource across all namespaces.")
	output.AddFormatFlag(flags, &ctx.outputFormat)
}

// AllNamespaces implements the Context interface.
func (ctx *context) Namespace() string {
	if ctx.allNamespaces {
		return ""
	}

	return ctx.Context.Namespace()
}

// OutputFormat implements the Context interface.
func (ctx *context) OutputFormat() output.Format {
	return output.Format(ctx.outputFormat)
}

// Printer implements the Context interface.
func (ctx *context) Printer() output.Printer {
	return ctx.printer
}

// Factory implements the Context interface.
func (ctx *context) Factory() resource.Factory {
	return ctx.factory
}

// Formatter implements the Context interface.
func (ctx *context) Formatter() resource.Formatter {
	return ctx.formatter
}
