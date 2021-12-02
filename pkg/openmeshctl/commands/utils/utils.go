package utils

import (
	"io"
	"os"
	"os/exec"

	"github.com/gobuffalo/packr"
	"github.com/rotisserie/eris"
)

var box = packr.NewBox("./scripts")

func SwitchContext(kubeCtx string) error {
	if err := runScript(box, os.Stdout, "context_switch.sh", kubeCtx); err != nil {
		return eris.Wrapf(err, "Could not switch context to %s", kubeCtx)
	}

	return nil
}

func runScript(box packr.Box, out io.Writer, script string, args ...string) error {
	script, err := box.FindString(script)
	if err != nil {
		return eris.Wrap(err, "Error loading script")
	}

	cmd := exec.Command("bash", append([]string{"-c", script}, args...)...)
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
