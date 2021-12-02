package apply

import (
	"crypto/tls"
	"io/fs"
	"net/http"
	"os"
	"time"

	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/runtime"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

//go:generate mockgen -destination mocks/context.go -package mock_apply . Context

// Context contains all the options and other info for the apply command.
type Context interface {
	runtime.Context
	apply.Context

	// AddToFlags adds the configurable options to the flag set.
	AddToFlags(flags *pflag.FlagSet)

	// Filenames returns the filenames to load Kubernetes resources from to apply.
	// The filenames may be local paths or URLs to remote locations.
	Filenames() []string

	// Applier returns the applier for the given GVK from the current context.
	// If there is one set for a type, it returns that, otherwise a default server-side applier.
	Applier(gvk *schema.GroupVersionKind) resource.Applier

	// FS returns a local file system for loading local resources.
	FS() fs.FS

	// HttpClient returns an HTTP client for loading remote resources.
	HttpClient() *http.Client
}

type context struct {
	runtime.Context
	filenames     []string
	fs            fs.FS
	httpClient    *http.Client
	dynamicClient dynamic.Interface
}

// NewContext returns a new apply command context, wrapping the given root context.
func NewContext(rootCtx runtime.Context) Context {
	return &context{
		Context: rootCtx,
		fs:      os.DirFS("."),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

// AddToFlags implements the Context interface.
func (ctx *context) AddToFlags(flags *pflag.FlagSet) {
	flags.StringArrayVarP(&ctx.filenames, "filenames", "f", []string{}, "The filenames that contain the objects to apply")
}

// Filenames implements the Context interface.
func (ctx *context) Filenames() []string {
	return ctx.filenames
}

// Applier implements the Context interface.
func (ctx *context) Applier(gvk *schema.GroupVersionKind) resource.Applier {
	applier, ok := ctx.Registry().Get(gvk)
	if !ok {
		return apply.ServerSideApplier(gvk)
	}

	return applier.Applier
}

// FS implements the Context interface.
func (ctx *context) FS() fs.FS {
	return ctx.fs
}

// HttpClient implements the Context interface.
func (ctx *context) HttpClient() *http.Client {
	return ctx.httpClient
}

// DynamicClient implements the Context interface.
func (ctx *context) DynamicClient() (dynamic.Interface, error) {
	if ctx.dynamicClient != nil {
		return ctx.dynamicClient, nil
	}

	cfg, err := ctx.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	ctx.dynamicClient = client
	return client, nil
}
