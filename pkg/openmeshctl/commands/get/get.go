package get

import (
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Command creates a new get command that can be attached to the root command tree.
func Command(ctx runtime.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many resources",
	}

	for _, cfg := range ctx.Registry().List() {
		cmd.AddCommand(Subcommand(ctx, cfg))
	}

	return cmd
}

// Subcommand builds a new command for a specific resource type.
func Subcommand(rootCtx runtime.Context, cfg resource.Config) *cobra.Command {
	ctx := NewContext(rootCtx, cfg)
	cmd := &cobra.Command{
		Use:          cfg.Name + " [NAME]",
		Aliases:      append(cfg.Aliases, cfg.Plural),
		SilenceUsage: true,
		Short:        "Get information about managed " + cfg.Plural,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return eris.Errorf("unexpected arguments: %s", args[1:])
			} else if len(args) == 0 {
				return All(ctx)
			}

			return One(ctx, args[0])
		},
	}

	ctx.AddToFlags(cmd.Flags())
	return cmd
}

// One prints a named resource to the output.
func One(ctx Context, name string) error {
	if ctx.Namespace() == "" {
		return eris.New("a resource cannot be retrieved by name across all namespaces")
	}

	kubeClient, err := ctx.KubeClient()
	if err != nil {
		return err
	}
	obj := ctx.Factory().New()
	key := client.ObjectKey{Name: name, Namespace: ctx.Namespace()}
	if err := kubeClient.Get(ctx, key, obj); err != nil {
		return err
	}

	return printResources(ctx, obj)
}

// All prints all resources of a type.
func All(ctx Context) error {
	kubeClient, err := ctx.KubeClient()
	if err != nil {
		return err
	}
	list := ctx.Factory().NewList()
	listOpts := client.ListOptions{Namespace: ctx.Namespace()}
	if err := kubeClient.List(ctx, list, &listOpts); err != nil {
		return err
	}

	return printResources(ctx, list)
}

func printResources(ctx Context, obj k8s_runtime.Object) error {
	switch output.Format(ctx.OutputFormat()) {
	case output.Default, output.Wide:
		return printTable(ctx, obj)
	default:
		return ctx.Printer().PrintRaw(obj, ctx.OutputFormat())
	}
}

func printTable(ctx Context, obj k8s_runtime.Object) error {
	objs, err := toList(obj)
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		if ctx.Namespace() != "" {
			fmt.Fprintf(ctx.Out(), "No resources found in %s namespace.\n", ctx.Namespace())
		} else {
			fmt.Fprintf(ctx.Out(), "No resources found.\n")
		}
		return nil
	}

	table := ctx.Formatter().ToTable(objs, ctx.Namespace() == "", ctx.OutputFormat() == output.Wide)
	return ctx.Printer().PrintTable(table)
}

func toList(obj k8s_runtime.Object) ([]k8s_runtime.Object, error) {
	if _, ok := obj.(client.ObjectList); !ok {
		return []k8s_runtime.Object{obj}, nil
	}

	return meta.ExtractList(obj)
}
