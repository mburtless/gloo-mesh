package virtualmesh

import (
	"fmt"

	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/output"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/apply"
	"github.com/solo-io/gloo-mesh/pkg/openmeshctl/resource/registry"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	registry.RegisterType(resource.Config{
		Name:      "virtualmesh",
		GVK:       networkingv1.VirtualMeshGVK,
		Plural:    "virtualmeshes",
		Aliases:   []string{"vm", "vms"},
		Factory:   Factory{},
		Formatter: Formatter{},
		Applier:   apply.ServerSideApplier(&networkingv1.VirtualMeshGVK),
	})
}

// Factory creates new virtual mesh objects.
type Factory struct{}

// New returns a new virtual mesh.
func (f Factory) New() client.Object {
	return &networkingv1.VirtualMesh{}
}

// NewList returns a new virtual mesh list.
func (f Factory) NewList() client.ObjectList {
	return &networkingv1.VirtualMeshList{}
}

// Formatter formats virtual mesh objects.
type Formatter struct{}

// ToSummary converts a virtual mesh to a summary representation.
func (f Formatter) ToSummary(obj runtime.Object) *output.Summary {
	vm := obj.(*networkingv1.VirtualMesh)
	fieldSet := output.FieldSet{}
	fieldSet.AddField("Meshes", vm.Spec.GetMeshes())
	// TODO(ryantking): MTLS config
	// TODO(ryantking): federation
	// TODO(ryantking): global access policy
	// TODO(ryantking): applied destinations
	return &output.Summary{Meta: vm, Fields: fieldSet}
}

// ToTable converts a list of meshes into a table
func (f Formatter) ToTable(objs []runtime.Object, includeNS, wide bool) *output.Table {
	return &output.Table{
		Headers: f.makeHeaders(includeNS),
		NextRow: f.makeNextTableRowFunc(objs, includeNS),
	}
}

func (f Formatter) makeHeaders(includeNS bool) []string {
	headers := make([]string, 0, 8)
	if includeNS {
		headers = append(headers, "namespace")
	}
	headers = append(headers, "name", "meshes", "destinations", "age")

	return headers
}

func (f Formatter) makeNextTableRowFunc(objs []runtime.Object, includeNS bool) func() []string {
	ndx := 0
	return func() []string {
		if ndx == len(objs) {
			return nil
		}

		vm := objs[ndx].(*networkingv1.VirtualMesh)
		ndx++

		data := make([]string, 0, 8)
		if includeNS {
			data = append(data, vm.GetNamespace())
		}
		data = append(
			data,
			vm.GetName(),
			f.buildMeshesCell(vm),
			f.buildDestinationCell(vm),
			output.FormatAge(vm.GetCreationTimestamp()),
		)

		return data
	}
}

func (f Formatter) buildMeshesCell(vm *networkingv1.VirtualMesh) string {
	meshes := vm.Status.GetMeshes()
	healthy := 0
	for _, dest := range meshes {
		if dest.GetState() == commonv1.ApprovalState_ACCEPTED {
			healthy++
		}
	}

	return fmt.Sprintf("%d/%d", healthy, len(meshes))
}

func (f Formatter) buildDestinationCell(vm *networkingv1.VirtualMesh) string {
	dests := vm.Status.GetDestinations()
	healthy := 0
	for _, dest := range dests {
		if dest.GetState() == commonv1.ApprovalState_ACCEPTED {
			healthy++
		}
	}

	return fmt.Sprintf("%d/%d", healthy, len(dests))
}
