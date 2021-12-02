package registry

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var global = NewTypeRegistry()

// Global returns the global type registry.
func Global() TypeRegistry {
	return global
}

// RegisterType registers a type to the global registry.
func RegisterType(desc resource.Config) {
	if err := global.Register(desc); err != nil {
		panic(err)
	}
}

// TypeRegistry is a registry that allows resource types to be registered.
type TypeRegistry interface {
	Register(desc resource.Config) error
	Get(gvk *schema.GroupVersionKind) (resource.Config, bool)
	List() []resource.Config
}

type typeRegistry struct {
	configs map[string]resource.Config
	order   []string
}

// NewTypeRegistry returns a new type registry.
func NewTypeRegistry() TypeRegistry {
	return &typeRegistry{configs: make(map[string]resource.Config)}
}

// RegisterType registers a new type config.
func (r *typeRegistry) Register(cfg resource.Config) error {
	key := cfg.GVK.String()
	if _, ok := r.configs[key]; ok {
		return eris.Errorf("resource already registered for %s", key)
	}

	r.configs[key] = cfg
	r.order = append(r.order, key)
	return nil
}

// Get returns a config for the given GVK.
func (r *typeRegistry) Get(gvk *schema.GroupVersionKind) (resource.Config, bool) {
	cfg, ok := r.configs[gvk.String()]
	return cfg, ok
}

// List returns all registered type configs.
func (r *typeRegistry) List() []resource.Config {
	cfgs := make([]resource.Config, len(r.configs))
	for i, key := range r.order {
		cfgs[i] = r.configs[key]
	}

	return cfgs
}
