package runtime

import (
	pkgcontext "context"
	"io"
	"os"

	"github.com/solo-io/gloo-mesh/pkg/common/schemes"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/registry"
	"github.com/spf13/pflag"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockgen -destination mocks/context.go -package mock_runtime . Context

// Context contains the common values to the execution of a meshctl command.
// It can be used to satisfy the `context.Context` interface and the `genericclioptions.RESTClientGetter`
type Context interface {
	pkgcontext.Context
	genericclioptions.RESTClientGetter
	output.Context

	// KubeConfig returns the path to the KubeConfig.
	KubeConfig() string

	// KubeContext returns the name of the target KubeContext. If no KubeContext is provided, it defaults to the current context.
	KubeContext() (string, error)

	// KubeClient returns a Kubernetes client.
	KubeClient() (client.Client, error)

	// Namespace is the name of Kubernetes namespace to interact with.
	Namespace() string

	// Registry returns the type registry.
	Registry() registry.TypeRegistry
}

var _ Context = &context{}

type context struct {
	pkgcontext.Context
	RESTClientGetter
	namespace   string
	registry    registry.TypeRegistry
	out, errOut io.Writer
	verbose     bool
}

// KubeConfig implements the Context interface.
func (ctx *context) KubeConfig() string {
	return ctx.RESTClientGetter.KubeConfig
}

// KubeContext implements the Context interface.
func (ctx *context) KubeContext() (string, error) {
	if ctx.RESTClientGetter.KubeContext != "" {
		return ctx.RESTClientGetter.KubeContext, nil
	}

	cfg, err := ctx.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return "", err
	}
	return cfg.CurrentContext, nil
}

// KubeClient implements the Context interface.
func (ctx *context) KubeClient() (client.Client, error) {
	scheme := scheme.Scheme
	if err := schemes.AddToScheme(scheme); err != nil {
		return nil, err
	}

	// needed for in-cluster CRD check
	if err := apiextensionsv1beta1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	cfg, err := ctx.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Namespace implements the Context interface.
func (ctx *context) Namespace() string {
	return ctx.namespace
}

// Registry implements the Context interface.
func (ctx *context) Registry() registry.TypeRegistry {
	return ctx.registry
}

// Out implements the Context interface.
func (ctx *context) Out() io.Writer {
	return ctx.out
}

// ErrOut implements the Context interface.
func (ctx *context) ErrOut() io.Writer {
	return ctx.errOut
}

// Verbose implements the Context interface.
func (ctx *context) Verbose() bool {
	return ctx.verbose
}

// DefaultContext returns a Context with default values.
// If the provided flag set is non-nil, it will have context options added to it.
func DefaultContext(flags *pflag.FlagSet) Context {
	ctx := context{
		Context:          pkgcontext.Background(),
		RESTClientGetter: RESTClientGetter{},
		out:              os.Stdout,
		errOut:           os.Stderr,
		registry:         registry.Global(),
	}
	if flags != nil {
		flags.StringVar(&ctx.RESTClientGetter.KubeConfig, "kubeconfig", "",
			"Path to the kubeconfig from which the management cluster will be accessed")
		flags.StringVar(&ctx.RESTClientGetter.KubeContext, "context", "", "Name of the kubeconfig context to use for the management cluster")
		flags.StringVarP(&ctx.namespace, "namespace", "n", "gloo-mesh",
			"Namespace that the management plan is installed in on the management cluster")
		flags.BoolVarP(&ctx.verbose, "verbose", "v", false, "Show more detailed output information.")
	}

	return &ctx
}
