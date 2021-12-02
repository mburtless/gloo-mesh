package runtime

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// RESTClientGetter is an implementation of `genericclioptions.RESTClientGetter` that returns Kubernetes clients
// build from the values stored in the struct.
type RESTClientGetter struct {
	KubeConfig  string
	KubeContext string
}

// ToRESTConfig returns a Kubernetes REST client configuration.
// If one is already stored in the context, it returns that one, otherwise it builds a new one from the Kubernetes
// configuration values stored in the context.
// Part of the `genericclioptions.RESTClientGetter` interface.
func (g RESTClientGetter) ToRESTConfig() (*rest.Config, error) {
	cfg, err := g.ToRawKubeConfigLoader().ClientConfig()
	if err != nil {
		return nil, err
	}
	cfg.WarningHandler = noopWarningHandler{}
	return cfg, nil
}

// silence warnings
type noopWarningHandler struct{}

func (f noopWarningHandler) HandleWarningHeader(code int, agent string, message string) {}

// ToDiscoveryClient returns a Kubernetes discovery client.
// If one is already stored in the context, it returns that one, otherwise it builds a new one from the Kubernetes
// configuration values stored in the context.
// Part of the `genericclioptions.RESTClientGetter` interface.
func (g RESTClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	cfg, err := g.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	client, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return nil, err
	}
	cachedClient := memory.NewMemCacheClient(client)
	return cachedClient, nil
}

// ToRESTMapper returns a Kubernetes REST mapper.
// If one is already stored in the context, it returns that one, otherwise it builds a new one from the Kubernetes
// configuration values stored in the context.
// Part of the `genericclioptions.RESTClientGetter` interface.
func (g RESTClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	client, err := g.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(client)
	expander := restmapper.NewShortcutExpander(mapper, client)
	return expander, nil
}

// ToRawConfigLoader returns a Kubernetes config loader.
// If one is already stored in the context, it returns that one, otherwise it builds a new one from the Kubernetes
// configuration values stored in the context.
// Part of the `genericclioptions.RESTClientGetter` interface.
func (g RESTClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = append(loadingRules.Precedence, clientcmd.RecommendedHomeFile)
	if g.KubeConfig != "" {
		loadingRules.ExplicitPath = g.KubeConfig
	}
	overrides := &clientcmd.ConfigOverrides{}
	if g.KubeContext != "" {
		overrides.CurrentContext = g.KubeContext
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)
}
