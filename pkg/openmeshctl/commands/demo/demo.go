package demo

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/demo/create"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/demo/destroy"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"

	"github.com/spf13/cobra"
)

func Command(ctx runtime.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Bootstrap environments for various demos demonstrating Gloo Mesh functionality.",
	}

	cmd.AddCommand(
		create.Command(ctx),
		destroy.Command(ctx),
	)

	return cmd
}
