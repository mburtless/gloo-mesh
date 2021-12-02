package destroy

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/rotisserie/eris"

	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"
)

var (
	clusters = []string{"mgmt-cluster", "remote-cluster"}
)

func Command(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "destroy",
		Short: "Clean up bootstrapped local resources. This will delete the kind clusters created by the \"openmeshctl demo create\"",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cleanup(ctx)
		},
	}

	cmd.SilenceUsage = true
	return cmd
}

func cleanup(ctx context.Context) error {
	fmt.Println("Cleaning up clusters")

	box := packr.NewBox("./scripts")
	script, err := box.FindString("delete_clusters.sh")
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	args := []string{"-c", script}
	args = append(args, clusters...)
	cmd := exec.CommandContext(ctx, "bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return err
}
