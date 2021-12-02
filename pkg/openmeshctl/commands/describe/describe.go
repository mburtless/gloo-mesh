package describe

import (
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/meta"
	k8s_runtime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Command creates a new describe command that can be attached to the root command tree.
func Command(ctx runtime.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Human readable description of discovered resources and applicable configuration",
	}

	for _, cfg := range ctx.Registry().List() {
		cmd.AddCommand(Subcommand(ctx, cfg))
	}

	return cmd
}

// Subcommand returns a get command for a specific resource type.
func Subcommand(rootCtx runtime.Context, cfg resource.Config) *cobra.Command {
	ctx := NewContext(rootCtx, cfg)
	cmd := &cobra.Command{
		Use:          cfg.Name + " [NAME]",
		Aliases:      append(cfg.Aliases, cfg.Plural),
		SilenceUsage: true,
		Short:        fmt.Sprintf("Description of managed %s", cfg.Plural),
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
	objs, err := toList(obj)
	if err != nil {
		return err
	}
	if len(objs) == 0 {
		fmt.Fprintf(ctx.Out(), "No resources found in %s namespace.\n", ctx.Namespace())
		return nil
	}
	for i, obj := range objs {
		if i != 0 {
			fmt.Fprint(ctx.Out(), "\n\n")
		}

		summary := ctx.Formatter().ToSummary(obj)
		ctx.Printer().PrintSummary(summary)
	}

	return nil
}

func toList(obj k8s_runtime.Object) ([]k8s_runtime.Object, error) {
	if _, ok := obj.(client.ObjectList); !ok {
		return []k8s_runtime.Object{obj}, nil
	}

	return meta.ExtractList(obj)
}
