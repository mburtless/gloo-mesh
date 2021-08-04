package check

import (
	"context"

	"github.com/spf13/cobra"
)

// inCluster is true if running from a Helm Hook context, else for meshctl it's false
func Command(ctx context.Context, inCluster bool) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Perform health checks on the Gloo Mesh system",
	}

	cmd.AddCommand(
		serverCmd(ctx, inCluster),
		agentCmd(ctx, inCluster),
		// TODO add comprehensive system wide check based on meshctl config file
	)

	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}
