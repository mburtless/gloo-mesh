package apply

import (
	"context"
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
)

//go:generate mockgen -package mocks -destination mocks/context.go . Context
//go:generate mockgen -package mocks -destination mocks/dynamic.go k8s.io/client-go/dynamic Interface,NamespaceableResourceInterface,ResourceInterface

// Context holds the required background information for a resource apply.
type Context interface {
	context.Context
	genericclioptions.RESTClientGetter

	// Namespace returns the current Kubernetes namespace
	Namespace() string

	// DynamicClient returns a new dynamic client.
	DynamicClient() (dynamic.Interface, error)
}

// ApplyFunc is a function that can act as an applier.
type ApplyFunc func(ctx Context, obj metav1.Object) error

// Apply runs the function on the given parameters.
func (f ApplyFunc) Apply(ctx Context, obj metav1.Object) error {
	return f(ctx, obj)
}

// ServerSideApplier returns a new applier that does a Kubernetes SSA.
// Optionally supports hooks.
func ServerSideApplier(gvk *schema.GroupVersionKind, hooks ...ApplyFunc) ApplyFunc {
	ssa := serverSideApplier{gvk: gvk}
	return func(ctx Context, obj metav1.Object) error {
		for _, hook := range hooks {
			if err := hook(ctx, obj); err != nil {
				return err
			}
		}

		return ssa.apply(ctx, obj)
	}
}

type serverSideApplier struct {
	gvk *schema.GroupVersionKind

	dynClient dynamic.Interface
}

func (a *serverSideApplier) apply(ctx Context, obj metav1.Object) error {
	mapping, err := a.mapObject(ctx, obj)
	if err != nil {
		return err
	}
	name, namespace, data, err := a.encodeObject(ctx, obj)
	if err != nil {
		return err
	}
	client, err := ctx.DynamicClient()
	if err != nil {
		return err
	}

	_, err = client.
		Resource(mapping.Resource).
		Namespace(namespace).
		Patch(ctx, name, types.ApplyPatchType, data, metav1.PatchOptions{FieldManager: "meshctl"})

	return err
}

func (a *serverSideApplier) mapObject(ctx Context, obj metav1.Object) (*meta.RESTMapping, error) {
	mapper, err := ctx.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	return mapper.RESTMapping(a.gvk.GroupKind(), a.gvk.Version)
}

func (a *serverSideApplier) encodeObject(ctx Context, obj metav1.Object) (string, string, []byte, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", "", nil, err
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = ctx.Namespace()
	}

	return obj.GetName(), namespace, data, nil
}
