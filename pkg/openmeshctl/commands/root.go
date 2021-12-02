package commands

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/demo"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/deregister"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/describe"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/get"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/install"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/register"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/uninstall"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/commands/version"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"

	// register resources
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/accesspolicy"
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/destination"
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/mesh"
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/trafficpolicy"
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/virtualmesh"
	_ "github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/workload"
)

// RootCommand returns the meshctl root command.
func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "openmeshctl",
		Short:         "The Command Line Interface for managing Gloo Mesh.",
		SilenceErrors: true,
	}
	ctx := runtime.DefaultContext(cmd.PersistentFlags())
	cmd.AddCommand(
		install.Command(ctx),
		register.Command(ctx),
		uninstall.Command(ctx),
		deregister.Command(ctx),
		describe.Command(ctx),
		get.Command(ctx),
		apply.Command(ctx),
		version.Command(ctx),
		demo.Command(ctx),
	)

	return cmd
}
