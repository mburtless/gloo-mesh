package apply

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
)

// Command returns a new apply command that can be attached to the root command tree.
func Command(rootCtx runtime.Context) *cobra.Command {
	ctx := NewContext(rootCtx)
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "Apply a Gloo Mesh resource",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return Apply(ctx)
		},
	}

	ctx.AddToFlags(cmd.Flags())
	return cmd
}

// Apply creates/updates all the Gloo Mesh resources in the given
func Apply(ctx Context) error {
	for _, fname := range ctx.Filenames() {
		if err := applyFile(ctx, fname); err != nil {
			return err
		}
	}

	return nil
}

func applyFile(ctx Context, fname string) error {
	f, err := openFile(ctx, fname)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(splitYamls)
	for scanner.Scan() {
		if err := applyData(ctx, scanner.Bytes()); err != nil {
			return err
		}

	}

	return nil
}

func openFile(ctx Context, fname string) (io.ReadCloser, error) {
	if !strings.HasPrefix(fname, "http://") && !strings.HasPrefix(fname, "https://") {
		return ctx.FS().Open(fname)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fname, nil)
	if err != nil {
		return nil, err
	}
	res, err := ctx.HttpClient().Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		// TODO(ryantking): Debug print
		io.ReadAll(res.Body)
		res.Body.Close()
		return nil, eris.Errorf("unable to load remote file: %d %s", res.StatusCode, res.Status)
	}

	return res.Body, nil
}

func splitYamls(data []byte, atEOF bool) (int, []byte, error) {
	for i := 0; i <= len(data)-4; i++ {
		if data[i] == '-' && data[i+1] == '-' && data[i+2] == '-' {
			advance := i + 3
			if i < len(data)-3 && data[i+3] == '\n' {
				advance++
			}

			return advance, data[:i], nil
		}
	}
	if !atEOF {
		return 0, nil, nil
	}

	return 0, data, bufio.ErrFinalToken
}

var decoder = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func decodeYAML(ctx Context, data []byte) (*unstructured.Unstructured, *schema.GroupVersionKind, error) {
	var obj unstructured.Unstructured
	_, gvk, err := decoder.Decode(data, nil, &obj)
	if err != nil {
		return nil, nil, err
	}

	return &obj, gvk, nil
}

func applyData(ctx Context, data []byte) error {
	obj, gvk, err := decodeYAML(ctx, data)
	if err != nil {
		return err
	}

	if err := ctx.Applier(gvk).Apply(ctx, obj); err != nil {
		return err
	}

	fmt.Fprintf(ctx.Out(), "%s/%s applied\n", gvk.Kind, obj.GetName())
	return nil
}
