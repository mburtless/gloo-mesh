package resource

import (
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Config contains information and functions for working with a resource generically.
type Config struct {
	// Name of the resource that's registered
	Name string

	// GVK is the Kubernetes GroupVersionKind of the resource.
	GVK schema.GroupVersionKind

	// Plural form of the resource name.
	// Defaults to the name with an s appended if not changed.
	Plural string

	// Aliases for the resource name for things like commands.
	Aliases []string

	// Factory handles the creation of new resources.
	Factory Factory

	// Formatter formats resources into various shapes.
	Formatter Formatter

	// Applier applies resources to the cluster.
	Applier Applier
}

//go:generate mockgen -destination mocks/factory.go -package mock_resource . Factory

// Factory can build new resources.
type Factory interface {
	// New returns a new instance of the resource.
	New() client.Object

	// NewList returns a new instance of the object list type.
	NewList() client.ObjectList
}

//go:generate mockgen -destination mocks/formatter.go -package mock_resource . Formatter

// Formatter is capable of formatting resources into various representations.
type Formatter interface {
	// ToSummary returns a summary of the object in a human readable format.
	ToSummary(obj runtime.Object) *output.Summary

	// ToTable returns a tabular representation of a list of objects.
	ToTable(objs []runtime.Object, includeNS, wide bool) *output.Table
}

//go:generate mockgen -destination mocks/applier.go -package mock_resource . Applier

// Applier is capable of applying resources to a cluster.
// If custom apply logic is required for a type, such as validation, then it
// can be added via an implementation of this interface.
type Applier interface {
	// Apply applies an object to the cluster.
	Apply(ctx apply.Context, obj metav1.Object) error
}
